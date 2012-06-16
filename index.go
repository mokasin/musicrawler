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

	// Make it nosync and disable the journal
	if _, err := i.db.Exec("PRAGMA synchronous=OFF"); err != nil {
		return nil, err
	}
	if _, err := i.db.Exec("PRAGMA journal_mode=OFF"); err != nil {
		return nil, err
	}

	// If databsae file does not exist
	if newdatabase {
		if err := i.createDatabase(); err != nil {
			return nil, err
		}
	}

	return i, nil
}

// Closes the opened database.
func (i *Index) Close() {
	i.db.Close()
}

// SQL queries to create the database schema
const SQL_CREATE_ARTIST = `
	CREATE TABLE Artist
	(
		ID				INTEGER		NOT NULL PRIMARY KEY,
		name			TEXT		UNIQUE
	);`
const SQL_CREATE_ALBUM = `
	CREATE TABLE Album
	(
		ID				INTEGER		NOT NULL PRIMARY KEY,
		name			TEXT		UNIQUE
	);`
const SQL_CREATE_TRACK = `
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

// Creates the basic database structure.
func (i *Index) createDatabase() error {
	sqls := []string{
		SQL_CREATE_ARTIST,
		SQL_CREATE_ALBUM,
		SQL_CREATE_TRACK,
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

const SQL_INSERT_ARTIST = "INSERT OR IGNORE INTO Artist(name) VALUES (?);"
const SQL_INSERT_ALBUM = "INSERT OR IGNORE INTO Album(name)  VALUES (?);"
const SQL_ADD_TRACK = `INSERT INTO Track(
	path,
	title,
	trackartist,
	trackalbum,
	tracknumber,
	year,
	filemtime,
	dbmtime)
    VALUES( ?, ?, 
		   (SELECT ID FROM Artist WHERE name = ?), 
		   (SELECT ID FROM Album  WHERE name = ?), 
		    ?, ?, ?, ?);`
const SQL_UPDATE_TIMESTAMP = "UPDATE Track SET dbmtime = ? WHERE path = ?;"

// Updates the timestamp if the track is in the database. The timestamp shows
// the last time the entry was touched.
func (i *Index) updateTrackTimestamp(track TrackInfo, tx *sql.Tx) error {
	_, err := tx.Exec(SQL_UPDATE_TIMESTAMP, i.timestamp, track.Path())
	return err
}

// Adds a track into the database using an existing transaction tx.
func (i *Index) addTrack(track TrackInfo, tx *sql.Tx) error {
	tag, err := track.Tags()
	if err != nil {
		return err
	}

	// first make sure artist…
	if _, err := tx.Exec(SQL_INSERT_ARTIST, tag.Artist); err != nil {
		return err
	}
	// …and album exist in database	
	if _, err := tx.Exec(SQL_INSERT_ALBUM, tag.Album); err != nil {
		return err
	}

	_, err = tx.Exec(SQL_ADD_TRACK,
		track.Path(),
		tag.Title,
		tag.Artist,
		tag.Album,
		tag.Track,
		tag.Year,
		track.Mtime(),
		i.timestamp,
	)
	return err
}

const SQL_UPDATE_TRACK = `UPDATE Track SET
	title       = ?,
	trackartist = (SELECT ID FROM Artist WHERE name = ?),
	trackalbum  = (SELECT ID FROM Album  WHERE name = ?),
	tracknumber = ?,
	year        = ?,
	filemtime   = ?
	WHERE path  = ?;`

// Update a changed track in the database using an existing transaction tx.
func (i *Index) updateTrack(track TrackInfo, tx *sql.Tx) error {
	tag, err := track.Tags()
	if err != nil {
		return err
	}

	// first make sure artist…
	if _, err := tx.Exec(SQL_INSERT_ARTIST, tag.Artist); err != nil {
		return err
	}
	// …and album exist in database
	if _, err := tx.Exec(SQL_INSERT_ALBUM, tag.Album); err != nil {
		return err
	}

	_, err = tx.Exec(SQL_UPDATE_TRACK,
		tag.Title,
		tag.Artist,
		tag.Album,
		tag.Track,
		tag.Year,
		track.Mtime(),
		track.Path(),
	)
	return err
}

// define databse actions
const (
	TRACK_NOUPDATE = iota
	TRACK_UPDATE
	TRACK_ADD
)

// Deletes all entries that have an outdated timestamp dbmtime. Also cleans up
// entries in Artist and Album table that are not referenced anymore in the
// Track-table.
//
// Returns the number of deleted rows and an error.
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

// Holds information of how the track at path was handeled. If the transaction
// was successfully err is nil.
type UpdateStatus struct {
	path   string
	action uint8
	err    error
}

// Holds information if the operation was successful.
type UpdateResult struct {
	err error
}

// Updates or adds tracks that are received at the tracks channel.
//
// For every track a status update UpdateStatus is emitted to the status
// channel. If the method finishes, the overall result is emitted on the result
// channel.
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
	rows, err := tx.Prepare(
		"SELECT path,filemtime FROM Track WHERE path = ?")
	if err != nil {
		result <- &UpdateResult{err: err}
		return
	}
	defer rows.Close()

	var trackAction uint8
	var trackPath string
	var trackMtime int64

	// traverse all catched pathes and update or add database entries
	for ti := range tracks {
		var statusErr error

		trackAction = TRACK_NOUPDATE

		// check if mtime has changed and decide what to do
		switch err := rows.QueryRow(ti.Path()).Scan(&trackPath,
			&trackMtime); {
		case err == nil: // track is in database
			// update timestamp because file still exists
			statusErr = i.updateTrackTimestamp(ti, tx)
			if statusErr == nil {
				if ti.Mtime() != trackMtime {
					statusErr = i.updateTrack(ti, tx)
					trackAction = TRACK_UPDATE
				}
			}
		case err == sql.ErrNoRows: // track is not in database
			// automatically update of timestamp when adding (performance)
			statusErr = i.addTrack(ti, tx)
			trackAction = TRACK_ADD
		default:
			// if something is wrong update timestamp, so track is not
			// deleted the next time
			statusErr = i.updateTrackTimestamp(ti, tx)
		}

		status <- &UpdateStatus{
			path:   ti.Path(),
			action: trackAction,
			err:    statusErr}
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		result <- &UpdateResult{err: err}
		return
	}

	result <- &UpdateResult{err: nil}
}

// Returns a TrackTags struct of the track at the given path.
func (i *Index) GetTrackByPath(path string) (t *TrackTags, err error) {
	stmt, err := i.db.Prepare(
		`SELECT tr.title, tr.year, tr.tracknumber, ar.name, al.name
			FROM Track tr
				JOIN Artist ar ON tr.trackartist = ar.ID
				JOIN Album  al ON tr.trackalbum  = al.ID
			WHERE tr.path = ?;`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var title, artist, album string
	var year, track uint
	if err := stmt.QueryRow(path).Scan(
		&title, &year, &track, &artist, &album); err != nil {
		return nil, err
	}

	return &TrackTags{
		Path:    path,
		Title:   title,
		Artist:  artist,
		Album:   album,
		Comment: "", // not yet in database
		Genre:   "",
		Year:    year,
		Track:   track,
	}, nil
}
