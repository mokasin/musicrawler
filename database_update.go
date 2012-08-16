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
	"musicrawler/lib/database"
	"musicrawler/lib/database/mod"
	"musicrawler/lib/database/query"
	"musicrawler/lib/source"
	"musicrawler/model/album"
	"musicrawler/model/artist"
	"musicrawler/model/track"
)

// define databse actions
const (
	TRACK_NOUPDATE = iota
	TRACK_UPDATE
	TRACK_ADD
)

type trackMtime struct {
	ID    int   `column:"ID" set:"0"`
	Mtime int64 `column:"filemtime"`
}

type trackDBMtime struct {
	DBMtime int64 `column:"dbmtime"`
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
	Deleted int64
	Err     error
}

// Update is a wrapper for update method, that should be called when using in a
// goroutine.
//
// It makes sure everything is cleaned up nicely before the signal gets emmitted
// to prevent racing conditions when closing the database connection.
func UpdateDatabase(db *database.Database, tracks <-chan source.TrackInfo,
	status chan<- *UpdateStatus, result chan<- *UpdateResult) {
	// signal is emitted, not untils index.Update() has cleaned up everything
	result <- updateDatabase(db, tracks, status)
}

// Updates or adds tracks that are received at the tracks channel.
//
// For every track a status update UpdateStatus is emitted to the status
// channel. If the method finishes, the overall result is emitted on the result
// channel.
func updateDatabase(db *database.Database, tracks <-chan source.TrackInfo,
	status chan<- *UpdateStatus) *UpdateResult {

	err := db.BeginTransaction()
	if err != nil {
		close(status)
		return &UpdateResult{Err: err}
	}
	defer db.EndTransaction()

	var trackAction uint8

	tm := &trackMtime{}

	martists := mod.New(db, "artist")
	malbums := mod.New(db, "album")
	mtracks := mod.New(db, "track")

	// traverse all catched pathes and update or add database entries
	for ti := range tracks {
		var statusErr error

		trackAction = TRACK_NOUPDATE

		// check if mtime has changed and decide what to do
		err := query.New(db, "track").Where("path =", ti.Path()).Exec(tm)
		switch {
		case err == nil: // track is in database
			// check if track has changed since the last time
			if ti.Mtime() != tm.Mtime {
				trackAction = TRACK_UPDATE

			} else {
				tdbm := &trackDBMtime{db.Mtime()}
				statusErr = mtracks.Update(tm.ID, tdbm)
			}
		case err == sql.ErrNoRows: // track is not in database
			trackAction = TRACK_ADD
			tag, err := ti.Tags()
			if err != nil {
				statusErr = err
				break
			}

			artist := &artist.Artist{
				Name: tag.Artist,
			}

			res, err := martists.InsertIgnore(artist)
			if err != nil {
				statusErr = err
				break
			}

			aff, _ := res.RowsAffected()

			var artist_id int64

			// if entry exists
			if aff == 0 {
				err = query.New(db, "artist").
					Where("name =", tag.Artist).Limit(1).Exec(artist)
				if err != nil {
					statusErr = err
					break
				}

				artist_id = artist.Id
			} else {
				artist_id, _ = res.LastInsertId()
			}

			album := &album.Album{
				Name:     tag.Album,
				ArtistID: artist_id,
			}

			res, err = malbums.InsertIgnore(album)
			if err != nil {
				statusErr = err
				break
			}

			aff, _ = res.RowsAffected()

			var album_id int64

			// if entry exists
			if aff == 0 {
				err = query.New(db, "album").Where("name =", tag.Album).Exec(album)
				if err != nil {
					statusErr = err
					break
				}

				album_id = album.Id
			} else {
				album_id, _ = res.LastInsertId()
			}

			track := &track.RawTrack{
				Path:        ti.Path(),
				Title:       tag.Title,
				Tracknumber: tag.Track,
				Year:        tag.Year,
				Length:      tag.Length,
				Genre:       tag.Genre,
				AlbumID:     album_id,
				Filemtime:   ti.Mtime(),
				DBMtime:     db.Mtime(),
			}

			_, err = mtracks.Insert(track)
			if err != nil {
				statusErr = err
				break
			}
			statusErr = nil

		default:
			// if something is wrong update timestamp, so track is not
			// deleted the next time
			tdbm := &trackDBMtime{db.Mtime()}
			mtracks.Update(tm.ID, tdbm)
			statusErr = err
		}

		status <- &UpdateStatus{
			Path:   ti.Path(),
			Action: trackAction,
			Err:    statusErr}
	}

	//del, err := deleteDanglingEntries(db)
	del, err := int64(0), nil

	close(status)
	return &UpdateResult{Err: err, Deleted: del}
}

// Deletes all entries that have an outdated timestamp dbmtime. Also cleans up
// entries in Artist and Album table that are not referenced anymore in the
// Track-table.
//
// Returns the number of deleted rows and an error.
func deleteDanglingEntries(db *database.Database) (int64, error) {
	r, err := db.Execute("DELETE FROM Track WHERE dbmtime <> ?", db.Mtime())
	deletedTracks, _ := r.RowsAffected()
	if err != nil {
		return deletedTracks, err
	}

	if _, err := db.Execute("DELETE FROM Album WHERE ID IN " +
		"(SELECT Album.ID FROM Album LEFT JOIN Track ON " +
		"Album.ID = Track.album_id WHERE Track.album_id " +
		"IS NULL);"); err != nil {
		return deletedTracks, err
	}

	if _, err := db.Execute("DELETE FROM Artist WHERE ID IN " +
		"(SELECT Artist.ID FROM Artist LEFT JOIN Album ON " +
		"Artist.ID = Album.artist_id WHERE Album.artist_id " +
		"IS NULL);"); err != nil {
		return deletedTracks, err
	}

	return deletedTracks, nil
}
