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
	"container/list"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"gotaglib"
	"os"
	"time"
)

type Index struct {
	Filename string
	db       *sql.DB
}

// SQL queries to create the dabase schema
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
func NewDatabase(filename string) (*Index, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	return &Index{Filename: filename, db: db}, nil
}

// Closes the opened database.
func (i *Index) Close() {
	i.db.Close()
}

// Creates the basic database structure
func (i *Index) CreateDatabase() error {
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
	path   string
	mtime  int64
	action uint8 // one of consts above
}

// Adds or updates a track to or in the database. It requires a transaction, on
// which it can act.
// TODO doc
func (i *Index) updatetrack(utr *updateTrackRecord, dbmtime int64,
	tx *sql.Tx) error {

	sqls := map[string]string{
		"artist insert":   "INSERT INTO Artist(name) VALUES (?);",
		"album insert":    "INSERT INTO Album(name)  VALUES (?);",
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
			utr.path); err != nil {
			return err
		}

		// nothing more to do
		if utr.action == TRACK_NOUPDATE {
			return nil
		}
	}

	tag, err := gotaglib.NewTaggedFile(utr.path)
	if err != nil {
		return err
	}

	// FIXME the error handling really! sucks
	// first make sure artist and album exist in database	
	if _, err := tx.Exec(sqls["artist insert"], tag.Artist); err != nil &&
		err.Error() != "column name is not unique" {
		return err
	}
	if _, err := tx.Exec(sqls["album insert"], tag.Album); err != nil &&
		err.Error() != "column name is not unique" {
		return err
	}

	switch utr.action {
	case TRACK_UPDATE: // the track is in the database and it needs to be updated
		_, err = tx.Exec(sqls["track update"],
			tag.Title,
			tag.Artist,
			tag.Album,
			tag.Track,
			tag.Year,
			utr.mtime,
			utr.path,
		)
		if err != nil {
			return err
		}
	case TRACK_ADD: // the track is not in the database and needs to be added
		_, err = tx.Exec(sqls["track insert"],
			utr.path,
			tag.Title,
			tag.Artist,
			tag.Album,
			tag.Track,
			tag.Year,
			utr.mtime,
			dbmtime,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

//TODO doc, return list of delete entries
func (i *Index) deleteDanglingEntries(dbmtime int64) (sql.Result, error) {
	stmt, err := i.db.Prepare("DELETE FROM Track WHERE dbmtime <> ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return stmt.Exec(dbmtime)
}

type UpdateResult struct {
	added     int64
	updated   int64
	deleted   int64
	errors    int64
	timestamp int64
}

// TODO: Add pathes to results, return database error messages
// Updates or adds tracks in list. Delete all entries not in list.
func (i *Index) Update(list *list.List) (*UpdateResult, error) {
	// Get current time to set modify time of database entry

	result := new(UpdateResult)
	result.timestamp = time.Now().Unix()

	tx, err := i.db.Begin()
	if err != nil {
		return nil, err
	}

	// get tracks that need to be updated
	stmtQuery, err := i.db.Prepare(
		"SELECT path,filemtime FROM Track WHERE path = ?")
	if err != nil {
		return nil, err
	}
	defer stmtQuery.Close()

	// update and add…
	for e := list.Front(); e != nil; e = e.Next() {

		path, ok := e.Value.(string)
		if ok {
			fi, err := os.Stat(path)
			if err != nil {
				//FIXME is it safe to depend on mtime?
				continue
			}

			var trackAction uint8 = TRACK_NOUPDATE
			var trackPath string
			var trackFilemtime int64

			switch err := stmtQuery.QueryRow(path).Scan(&trackPath,
				&trackFilemtime); {
			case err == nil:
				if fi.ModTime().Unix() != trackFilemtime {
					trackAction = TRACK_UPDATE // update track
				}
			case err == sql.ErrNoRows:
				trackAction = TRACK_ADD // add track to db
			case err != nil:
				result.errors++
				continue //FIXME inform the caller that something is wrong
			}

			utr := &updateTrackRecord{
				path:   path,
				mtime:  fi.ModTime().Unix(),
				action: trackAction}

			// update or add the database entry
			if err := i.updatetrack(utr, result.timestamp, tx); err != nil {
				result.errors++
				continue //FIXME inform the caller that something is wrong
			}
			switch trackAction {
			case TRACK_UPDATE:
				result.updated++
			case TRACK_ADD:
				result.added++
			}
		}
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if r, err := i.deleteDanglingEntries(result.timestamp); err == nil {
		result.deleted, _ = r.RowsAffected()
	} else {
		return nil, err
	}

	return result, nil
}

// Returns a gotaglib.TaggedFile with all information about the track with
// filename 'filename'
func (i *Index) GetTrackByFile(filename string) (t *gotaglib.TaggedFile,
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

	return &gotaglib.TaggedFile{
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
