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
	"fmt"
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

type dbUpdateTrack struct {
	path         string
	mtime        int64
	mtimeChanged bool
}

// Adds or updates a track to or in the database. It requires a transaction, on
// which it can act.
//
// ui.retag == false, if 
// TODO doc
func (i *Index) updatetrack(ut *dbUpdateTrack, dbmtime int64, tx *sql.Tx) error {

	// TODO use map
	const (
		ARTIST_INSERT = iota
		ALBUM_INSERT
		TRACK_NORETAG
		TRACK_UPDATE
		TRACK_INSERT
	)

	sqls := []string{
		"INSERT INTO Artist(name) VALUES (?);",
		"INSERT INTO Album(name)  VALUES (?);",
		"UPDATE Track SET dbmtime = ? WHERE path = ?;",
		`UPDATE Track SET
			title       = ?,
			trackartist = (SELECT ID FROM Artist WHERE name = ?),
			trackalbum  = (SELECT ID FROM Album  WHERE name = ?),
			tracknumber = ?,
			year        = ?,
			filemtime   = ?,
			dbmtime     = ?
			WHERE path = ?;`,
		`INSERT INTO Track(
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

	// see if the entry exists in the database and update the timestamp
	if !ut.mtimeChanged {
		r, err := tx.Exec(sqls[TRACK_NORETAG],
			dbmtime,
			ut.path,
		)
		if err != nil {
			return err
		}

		affected, _ := r.RowsAffected()

		// if there was an entry, there is nothing more todo
		if affected != 0 {
			//TODO return result of successfull operations
			return nil
		}
	}

	// (ui.mtimeChanged == true) mtime has changed
	// (ui.mtimeChanged == false) or the track is not in the database

	tag, err := gotaglib.NewTaggedFile(ut.path)
	if err != nil {
		// FIXME no database entry is created!
		return err
	}

	// FIXME the error handling really! sucks
	// first make sure artist and album exist in database	
	if _, err := tx.Exec(sqls[ARTIST_INSERT], tag.Artist); err != nil &&
		err.Error() != "column name is not unique" {
		return err
	}
	if _, err := tx.Exec(sqls[ALBUM_INSERT], tag.Album); err != nil &&
		err.Error() != "column name is not unique" {
		return err
	}

	if ut.mtimeChanged {
		// the track is in the database and it needs to be updated
		_, err = tx.Exec(sqls[TRACK_UPDATE],
			tag.Title,
			tag.Artist,
			tag.Album,
			tag.Track,
			tag.Year,
			ut.mtime,
			dbmtime,
			ut.path,
		)
		if err != nil {
			return err
		}
	} else {
		// the track is not in the database and needs to be added

		// add new entry
		_, err = tx.Exec(sqls[TRACK_INSERT],
			ut.path,
			tag.Title,
			tag.Artist,
			tag.Album,
			tag.Track,
			tag.Year,
			ut.mtime,
			dbmtime,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Index) DeleteDanglingEntries(dbmtime int64) error {
	return nil
}

// UpdateFiles takes a list of MtimeFileInfo and tries to update all entries
// concerning modified time and file path. If there is no entry to update, one
// is created.
func (i *Index) Update(list *list.List) (timestamp int64, err error) {
	// Get current time to set modify time of database entry
	timestamp = time.Now().Unix()

	tx, err := i.db.Begin()
	if err != nil {
		return timestamp, err
	}

	// get tracks that need to be updated
	stmtQuery, err := i.db.Prepare(
		"SELECT path FROM Track WHERE path = ? AND filemtime <> ?")
	if err != nil {
		return 0, err
	}
	defer stmtQuery.Close()

	// update and add…
	for e := list.Front(); e != nil; e = e.Next() {

		path, ok := e.Value.(string)
		if ok {
			fi, err := os.Stat(path)
			if err != nil {
				continue
			}

			ut := &dbUpdateTrack{
				path:         path,
				mtime:        fi.ModTime().Unix(),
				mtimeChanged: false}

			var entry string
			err = stmtQuery.QueryRow(ut.path, ut.mtime).Scan(&entry)

			// CAUTION: The query is empty, if the file hasn't changed or if
			// there is no corresponding entry in the database!

			// if file exists and has changed
			// FIXME be more precise with error handling
			if err == nil {
				ut.mtimeChanged = true
			}

			// update or add the database entry
			err = i.updatetrack(ut, timestamp, tx)
			if err != nil {
				//FIXME inform the caller that something is wrong
				continue
				//FIXME Don't write something out!
				fmt.Println(err)
			}
		}
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return timestamp, err
	}

	err = i.DeleteDanglingEntries(timestamp)
	return timestamp, nil
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
