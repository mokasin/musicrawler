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
	"musicrawler/source"
	"time"
)

const (
	sql_insert_artist = "INSERT OR IGNORE INTO Artist(name) VALUES (?);"
	sql_insert_album  = `INSERT OR IGNORE INTO Album(name, artist_id)
	VALUES (?,
			(SELECT ID FROM Artist WHERE name = ?));`

	sql_add_track = `INSERT INTO Track(
	path,
	title,
	album_id,
	tracknumber,
	year,
	length,
	genre,
	filemtime,
	dbmtime)
    VALUES( ?, ?, 
		   (SELECT ID FROM Album  WHERE name = ?), 
		    ?, ?, ?, ?, ?, ?);`

	sql_update_timestamp = "UPDATE Track SET dbmtime = ? WHERE path = ?;"

	sql_update_track = `UPDATE Track SET
	title       = ?,
	album_id    = (SELECT ID FROM Album  WHERE name = ?),
	tracknumber = ?,
	year        = ?,
	length		= ?,
	genre       = ?,
	filemtime   = ?
	WHERE path  = ?;`
)

// Updates the timestamp if the track is in the database. The timestamp shows
// the last time the entry was touched.
func (d *Database) updateTrackTimestamp(track source.TrackInfo,
	stmtUpdateTimestamp *sql.Stmt) error {
	_, err := stmtUpdateTimestamp.Exec(d.timestamp, track.Path())
	return err
}

func (d *Database) insertArtistAlbum(tag *source.TrackTags, stmtInsertArtist *sql.Stmt,
	stmtInsertAlbum *sql.Stmt) error {
	if _, err := stmtInsertArtist.Exec(tag.Artist); err != nil {
		return err
	}

	if _, err := stmtInsertAlbum.Exec(tag.Album, tag.Artist); err != nil {
		return err
	}
	return nil
}

// Adds a track into the database using an existing transaction tx.
func (d *Database) addTrack(track source.TrackInfo, stmtInsertArtist *sql.Stmt,
	stmtInsertAlbum *sql.Stmt, stmtAddTrack *sql.Stmt) error {
	tag, err := track.Tags()
	if err != nil {
		return err
	}

	// first make sure artist and album exist in database
	if err := d.insertArtistAlbum(tag, stmtInsertArtist, stmtInsertAlbum); err != nil {
		return err
	}

	_, err = stmtAddTrack.Exec(
		track.Path(),
		tag.Title,
		tag.Album,
		tag.Track,
		tag.Year,
		tag.Length,
		tag.Genre,
		track.Mtime(),
		d.timestamp,
	)
	return err
}

// Update a changed track in the database using an existing transaction tx.
func (d *Database) updateTrack(track source.TrackInfo, stmtInsertArtist *sql.Stmt,
	stmtInsertAlbum *sql.Stmt, stmtUpdateTrack *sql.Stmt) error {
	tag, err := track.Tags()
	if err != nil {
		return err
	}

	// first make sure artist and album exist in database
	if err := d.insertArtistAlbum(tag, stmtInsertArtist, stmtInsertAlbum); err != nil {
		return err
	}

	_, err = stmtUpdateTrack.Exec(
		tag.Title,
		tag.Artist,
		tag.Album,
		tag.Track,
		tag.Year,
		tag.Length,
		tag.Genre,
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
func (d *Database) Update(tracks <-chan source.TrackInfo, status chan<- *UpdateStatus,
	result chan<- *UpdateResult) {
	// signal is emitted, not untils index.Update() has cleaned up everything
	result <- d.update(tracks, status)
}

// Updates or adds tracks that are received at the tracks channel.
//
// For every track a status update UpdateStatus is emitted to the status
// channel. If the method finishes, the overall result is emitted on the result
// channel.
func (d *Database) update(tracks <-chan source.TrackInfo,
	status chan<- *UpdateStatus) *UpdateResult {

	// Get current time to set modify time of database entry
	d.timestamp = time.Now().Unix()

	tx, err := d.db.Begin()
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
	stmtInsertArtist, err := tx.Prepare(sql_insert_artist)
	if err != nil {
		close(status)
		return &UpdateResult{Err: err}
	}
	defer stmtInsertArtist.Close()

	stmtInsertAlbum, err := tx.Prepare(sql_insert_album)
	if err != nil {
		close(status)
		return &UpdateResult{Err: err}
	}
	defer stmtInsertAlbum.Close()

	stmtUpdateTimestamp, err := tx.Prepare(sql_update_timestamp)
	if err != nil {
		close(status)
		return &UpdateResult{Err: err}
	}
	defer stmtUpdateTimestamp.Close()

	stmtAddTrack, err := tx.Prepare(sql_add_track)
	if err != nil {
		close(status)
		return &UpdateResult{Err: err}
	}
	defer stmtAddTrack.Close()

	stmtUpdateTrack, err := tx.Prepare(sql_update_track)
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
			statusErr = d.updateTrackTimestamp(ti, stmtUpdateTimestamp)
			if statusErr == nil {
				// check if track has changed since the last time
				if ti.Mtime() != trackMtime {
					statusErr = d.updateTrack(ti, stmtInsertArtist,
						stmtInsertAlbum, stmtUpdateTrack)
					trackAction = TRACK_UPDATE
				}
			}
		case err == sql.ErrNoRows: // track is not in database
			// automatically update of timestamp when adding (performance)
			statusErr = d.addTrack(ti, stmtInsertArtist, stmtInsertAlbum,
				stmtAddTrack)
			trackAction = TRACK_ADD
		default:
			// if something is wrong update timestamp, so track is not
			// deleted the next time
			statusErr = d.updateTrackTimestamp(ti, stmtUpdateTimestamp)
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

/*
// Deletes all entries that have an outdated timestamp dbmtime. Also cleans up
// entries in Artist and Album table that are not referenced anymore in the
// Track-table.
//
// Returns the number of deleted rows and an error.
func (d *Database) DeleteDanglingEntries() (int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, err
	}

	r, err := tx.Exec("DELETE FROM Track WHERE dbmtime <> ?", d.timestamp)
	deletedTracks, _ := r.RowsAffected()
	if err != nil {
		return deletedTracks, err
	}

	//TODO needs rework
	if _, err := tx.Exec("DELETE FROM Artist WHERE ID IN " +
		"(SELECT Artist.ID FROM Artist LEFT JOIN Track ON " +
		"Artist.ID = Track.trackartist WHERE Track.trackartist " +
		"IS NULL);"); err != nil {
		return deletedTracks, err
	}
	if _, err := tx.Exec("DELETE FROM Album WHERE ID IN " +
		"(SELECT Album.ID FROM Album LEFT JOIN Track ON " +
		"Album.ID = Track.trackalbum WHERE Track.trackalbum " +
		"IS NULL);"); err != nil {
		return deletedTracks, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return deletedTracks, nil
}
*/
