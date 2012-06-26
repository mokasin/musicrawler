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

package index

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"musicrawler/source"
	"os"
	"time"
)

type Index struct {
	Filename  string
	db        *sql.DB
	timestamp int64
	Tracks    *Tracks
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

	// initializing members
	i.Tracks = NewTracks(i)

	return i, nil
}

func (i *Index) Timestamp() int64 {
	return i.timestamp
}

// Closes the opened database.
func (i *Index) Close() {
	i.db.Close()
}

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

// Updates the timestamp if the track is in the database. The timestamp shows
// the last time the entry was touched.
func (i *Index) updateTrackTimestamp(track source.TrackInfo,
	stmtUpdateTimestamp *sql.Stmt) error {
	_, err := stmtUpdateTimestamp.Exec(i.timestamp, track.Path())
	return err
}

func (i *Index) insertArtistAlbum(tag *source.TrackTags, stmtInsertArtist *sql.Stmt,
	stmtInsertAlbum *sql.Stmt) error {
	if _, err := stmtInsertArtist.Exec(tag.Artist); err != nil {
		return err
	}

	if _, err := stmtInsertAlbum.Exec(tag.Album); err != nil {
		return err
	}
	return nil
}

// Adds a track into the database using an existing transaction tx.
func (i *Index) addTrack(track source.TrackInfo, stmtInsertArtist *sql.Stmt,
	stmtInsertAlbum *sql.Stmt, stmtAddTrack *sql.Stmt) error {
	tag, err := track.Tags()
	if err != nil {
		return err
	}

	// first make sure artist and album exist in database
	if err := i.insertArtistAlbum(tag, stmtInsertArtist, stmtInsertAlbum); err != nil {
		return err
	}

	_, err = stmtAddTrack.Exec(
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

// Update a changed track in the database using an existing transaction tx.
func (i *Index) updateTrack(track source.TrackInfo, stmtInsertArtist *sql.Stmt,
	stmtInsertAlbum *sql.Stmt, stmtUpdateTrack *sql.Stmt) error {
	tag, err := track.Tags()
	if err != nil {
		return err
	}

	// first make sure artist and album exist in database
	if err := i.insertArtistAlbum(tag, stmtInsertArtist, stmtInsertAlbum); err != nil {
		return err
	}

	_, err = stmtUpdateTrack.Exec(
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
func (i *Index) DeleteDanglingEntries() (int64, error) {
	stmt, err := i.db.Prepare("DELETE FROM Track WHERE dbmtime <> ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	r, err := stmt.Exec(i.timestamp)
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
	Path   string
	Action uint8
	Err    error
}

// Holds information if the operation was successful.
type UpdateResult struct {
	Err error
}

// Update is a wrapper for update method, that should be called when using in a
// goroutine.
//
// It makes sure everything is cleaned up nicely before the signal gets emmitted
// to prevent racing conditions when closing the database connection.
func (i *Index) Update(tracks <-chan source.TrackInfo, status chan<- *UpdateStatus,
	result chan<- *UpdateResult) {
	// signal is emitted, not untils index.Update() has cleaned up everything
	result <- i.update(tracks, status)
}

// Updates or adds tracks that are received at the tracks channel.
//
// For every track a status update UpdateStatus is emitted to the status
// channel. If the method finishes, the overall result is emitted on the result
// channel.
func (i *Index) update(tracks <-chan source.TrackInfo,
	status chan<- *UpdateStatus) *UpdateResult {

	// Get current time to set modify time of database entry
	i.timestamp = time.Now().Unix()

	tx, err := i.db.Begin()
	if err != nil {
		close(status)
		return &UpdateResult{Err: err}
	}

	// get tracks that need to be updated
	rows, err := tx.Prepare(
		"SELECT path,filemtime FROM Track WHERE path = ?")
	if err != nil {
		close(status)
		return &UpdateResult{Err: err}
	}
	defer rows.Close()

	// prepare insert statements
	stmtInsertArtist, err := tx.Prepare(SQL_INSERT_ARTIST)
	if err != nil {
		close(status)
		return &UpdateResult{Err: err}
	}
	defer stmtInsertArtist.Close()

	stmtInsertAlbum, err := tx.Prepare(SQL_INSERT_ALBUM)
	if err != nil {
		close(status)
		return &UpdateResult{Err: err}
	}
	defer stmtInsertAlbum.Close()

	stmtUpdateTimestamp, err := tx.Prepare(SQL_UPDATE_TIMESTAMP)
	if err != nil {
		close(status)
		return &UpdateResult{Err: err}
	}
	defer stmtUpdateTimestamp.Close()

	stmtAddTrack, err := tx.Prepare(SQL_ADD_TRACK)
	if err != nil {
		close(status)
		return &UpdateResult{Err: err}
	}
	defer stmtAddTrack.Close()

	stmtUpdateTrack, err := tx.Prepare(SQL_UPDATE_TRACK)
	if err != nil {
		close(status)
		return &UpdateResult{Err: err}
	}
	defer stmtUpdateTrack.Close()

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
			statusErr = i.updateTrackTimestamp(ti, stmtUpdateTimestamp)
			if statusErr == nil {
				// check if track has changed since the last time
				if ti.Mtime() != trackMtime {
					statusErr = i.updateTrack(ti, stmtInsertArtist,
						stmtInsertAlbum, stmtUpdateTrack)
					trackAction = TRACK_UPDATE
				}
			}
		case err == sql.ErrNoRows: // track is not in database
			// automatically update of timestamp when adding (performance)
			statusErr = i.addTrack(ti, stmtInsertArtist, stmtInsertAlbum,
				stmtAddTrack)
			trackAction = TRACK_ADD
		default:
			// if something is wrong update timestamp, so track is not
			// deleted the next time
			statusErr = i.updateTrackTimestamp(ti, stmtUpdateTimestamp)
		}

		status <- &UpdateStatus{
			Path:   ti.Path(),
			Action: trackAction,
			Err:    statusErr}
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		close(status)
		return &UpdateResult{Err: err}
	}

	close(status)
	return &UpdateResult{Err: nil}
}
