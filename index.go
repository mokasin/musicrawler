/*  Copyright 2012, mokasin
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

type Index struct {
	Filename  string
	db        *sql.DB
	timestamp int64
}

// SQL queries to create the database schema
const sql_create_artist = `
	CREATE TABLE Artist
	(
		ID				INTEGER		NOT NULL PRIMARY KEY,
		name			TEXT		UNIQUE
	);`
const sql_create_album = `
	CREATE TABLE Album
	(
		ID				INTEGER		NOT NULL PRIMARY KEY,
		name			TEXT		UNIQUE
	);`
const sql_create_track = `
	CREATE TABLE Track
	(
		path			TEXT		NOT NULL PRIMARY KEY,
		title			TEXT,
		tracknumber		INTEGER,
		year			INTEGER,
		trackartist		INTEGER	REFERENCES Album(ID) ON DELETE SET NULL,
		trackalbum		INTEGER	REFERENCES Artist(ID) ON DELETE SET NULL,
		filemtime		INTEGER,
		dbmtime			INTEGER
	);`

// Creates a new Index struct and connects it to the database at filename.
// Needs to be closed with method Close()!
func NewIndex(filename string) (*Index, error) {
	_, err := os.Stat(filename)
	newdatabase := os.IsNotExist(err)

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	i := &Index{Filename: filename, db: db}
	if newdatabase {
		return i, i.createDatabase()
	}

	return i, nil
}

// Closes the opened database.
func (i *Index) Close() {
	i.db.Close()
}

// Creates the basic database structure
func (i *Index) createDatabase() error {
	sqls := []string{
		sql_create_artist,
		sql_create_album,
		sql_create_track,
	}

	tx, err := i.db.Begin()
	if err != nil {
		return err
	}

	for _, sql := range sqls {
		_, err := tx.Exec(sql)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// define updateTrackRecord actions
const (
	TRACK_NOUPDATE = iota
	TRACK_UPDATE
	TRACK_ADD
)

// Information needed for updating an entry in the database.
// action is.
type updateTrackRecord struct {
	track  TrackInfo
	action uint8 // one of consts above
}

// Adds or updates a track to or in the database. It requires a transaction, on
// which it can act.
// TODO doc
func (i *Index) updatetrack(utr *updateTrackRecord, dbmtime int64,
	tx *sql.Tx) error {

	// SQL queries for all possible actions in a map
	sqls := map[string]string{
		"artist insert":   "INSERT OR IGNORE INTO Artist(name) VALUES (?);",
		"album insert":    "INSERT OR IGNORE INTO Album(name)  VALUES (?);",
		"track timestamp": "UPDATE Track SET dbmtime = ? WHERE path = ?;",
		"track update": `UPDATE Track SET
			title       = ?,
			trackartist = (SELECT ID FROM Artist WHERE name = ?),
			trackalbum  = (SELECT ID FROM Album  WHERE name = ?),
			tracknumber = ?,
			year        = ?,
			filemtime   = ?
			WHERE path = ?;`,
		"track insert": `INSERT INTO Track(
			path,
			title,
			trackartist,
			trackalbum,
			tracknumber,
			year,
			filemtime,
			dbmtime)
		 VALUES(
			 ?, ?, 
			 (SELECT ID FROM Artist WHERE name = ?), 
			 (SELECT ID FROM Album  WHERE name = ?), 
			 ?, ?, ?, ?);`,
	}

	// update the timestamp if the track is in the database
	if utr.action == TRACK_NOUPDATE || utr.action == TRACK_UPDATE {
		if _, err := tx.Exec(sqls["track timestamp"], dbmtime,
			utr.track.Path()); err != nil {
			return err
		}

		// nothing more to do, everything is up to date
		if utr.action == TRACK_NOUPDATE {
			return nil
		}
	}

	tag, err := utr.track.Tags()
	if err != nil {
		return err
	}

	// first make sure artist and album exist in database	
	if _, err := tx.Exec(sqls["artist insert"], tag.Artist); err != nil {
		return err
	}
	if _, err := tx.Exec(sqls["album insert"], tag.Album); err != nil {
		return err
	}

	switch utr.action {
	case TRACK_UPDATE: // the track is in the database and needs to be updated
		_, err := tx.Exec(sqls["track update"],
			tag.Title,
			tag.Artist,
			tag.Album,
			tag.Track,
			tag.Year,
			utr.track.Mtime(),
			utr.track.Path(),
		)
		if err != nil {
			return err
		}
	case TRACK_ADD: // the track is not in the database and needs to be added
		_, err := tx.Exec(sqls["track insert"],
			utr.track.Path(),
			tag.Title,
			tag.Artist,
			tag.Album,
			tag.Track,
			tag.Year,
			utr.track.Mtime(),
			dbmtime,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// Deletes all entries that have an outdated timestamp dbmtime. Also cleans up
// entries in Artist and Album table that are not referenced evermore in Track.
//
// Returns the number of deleted rows.
func (i *Index) DeleteDanglingEntries(dbmtime int64) (int64, error) {
	stmt, err := i.db.Prepare("DELETE FROM Track WHERE dbmtime <> ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	r, err := stmt.Exec(dbmtime)
	deletedTracks, _ := r.RowsAffected()
	if err != nil {
		return deletedTracks, err
	}

	if _, err := i.db.Exec("DELETE FROM Artist WHERE ID IN " +
		"(SELECT Artist.ID FROM Artist LEFT JOIN Track ON " +
		"Artist.ID = Track.trackartist WHERE Track.trackartist " +
		"IS NULL);"); err != nil {
		return deletedTracks, err
	}
	if _, err := i.db.Exec("DELETE FROM Album WHERE ID IN " +
		"(SELECT Album.ID FROM Album LEFT JOIN Track ON " +
		"Album.ID = Track.trackalbum WHERE Track.trackalbum " +
		"IS NULL);"); err != nil {
		return deletedTracks, err
	}

	return deletedTracks, nil
}

// Reports if update on path with action was successful.
type UpdateStatus struct {
	path   string
	action uint8
	err    error
}

// Reports how many tracks were deleted and if the operation was successful.
type UpdateResult struct {
	err error
}

// Updates or adds tracks in list. Delete all entries not in list.
//
// Requires a channel for updates on tracks and a result channel. If the
// function finishes, an UpdateResult goes down the result channel.
func (i *Index) Update(tracks <-chan TrackInfo, status chan<- *UpdateStatus,
	result chan<- *UpdateResult) {

	// Get current time to set modify time of database entry
	i.timestamp = time.Now().Unix()

	tx, err := i.db.Begin()
	if err != nil {
		result <- &UpdateResult{err: err}
		return
	}

	// get tracks that need to be updated
	rows, err := i.db.Prepare(
		"SELECT path,filemtime FROM Track WHERE path = ?")
	if err != nil {
		result <- &UpdateResult{err: err}
		return
	}
	defer rows.Close()

	var trackAction uint8
	var trackPath string
	var trackMtime int64

	var utr *updateTrackRecord = &updateTrackRecord{}

	// traverse filelist and update or add database entries
	for ti := range tracks {
		// default action
		trackAction = TRACK_NOUPDATE

		// check if mtime has changed and decide what to do
		switch err := rows.QueryRow(ti.Path()).Scan(&trackPath,
			&trackMtime); {
		case err == nil:
			if ti.Mtime() != trackMtime {
				trackAction = TRACK_UPDATE // update track
			}
		case err == sql.ErrNoRows:
			trackAction = TRACK_ADD // add track to db
		case err != nil:
			goto STATUS // something went wrong
		}

		// prepare the record
		utr = &updateTrackRecord{
			track:  ti,
			action: trackAction}

		// update or add the database entry
		err = i.updatetrack(utr, i.timestamp, tx)

	STATUS:
		status <- &UpdateStatus{path: ti.Path(), action: trackAction, err: err}
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		result <- &UpdateResult{err: err}
		return
	}

	result <- &UpdateResult{err: err}
}

// Returns a gotaglib.TaggedFile with all information about the track with
// filename 'filename'.
func (i *Index) GetTrackByFile(filename string) (t *TrackTags,
	err error) {

	stmt, err := i.db.Prepare(
		`SELECT tr.path, tr.title, tr.year, tr.tracknumber, ar.name, al.name
			FROM Track tr
				JOIN Artist ar ON tr.trackartist = ar.ID
				JOIN Album  al ON tr.trackalbum  = al.ID
			WHERE tr.path = ?;`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var path, title, artist, album string
	var year, track uint
	if err := stmt.QueryRow(filename).Scan(
		&path, &title, &year, &track, &artist, &album); err != nil {
		return nil, err
	}

	return &TrackTags{
		Filename: path,
		Title:    title,
		Artist:   artist,
		Album:    album,
		Comment:  "", // not yet in database
		Genre:    "",
		Year:     year,
		Track:    track,
	}, nil
}
