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
	"fmt"
	"musicrawler/source"
)

func rows2TrackList(rows *sql.Rows, array *[]source.TrackTags) error {
	var path, title, genre, artist, album string
	var year, track, length uint

	i := 0
	for rows.Next() {
		if err := rows.Scan(&path, &artist, &album, &title,
			&track, &length, &year, &genre); err == nil {
			(*array)[i] = source.TrackTags{
				Path:    path,
				Title:   title,
				Artist:  artist,
				Album:   album,
				Comment: "", // not yet in database
				Genre:   genre,
				Year:    year,
				Track:   track,
				Length:  length,
			}
		}
		i++
	}

	return rows.Err()
}

var cacheGetAllTracks []source.TrackTags
var ctimeGetAllTracks int64 = -1

// Returns list of source.TrackTags of all tracks in the database.
func (i *Index) GetAllTracks() (*[]source.TrackTags, error) {
	// if nothing has changed, just return cached array
	if ctimeGetAllTracks == i.timestamp {
		return &cacheGetAllTracks, nil
	}

	tx, err := i.db.Begin()
	if err != nil {
		return nil, err
	}

	var count int
	if err := tx.QueryRow("SELECT COUNT(path) FROM Track").Scan(&count); err != nil {
		return nil, err
	}

	// allocating big enough array
	cacheGetAllTracks := make([]source.TrackTags, count)

	rows, err := tx.Query(
		`SELECT tr.path, ar.name, al.name, tr.title,
		        tr.tracknumber, tr.length, tr.year, tr.genre
 		FROM Track tr
 		JOIN Artist ar ON tr.trackartist = ar.ID
 		JOIN Album  al ON tr.trackalbum  = al.ID;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	err = rows2TrackList(rows, &cacheGetAllTracks)
	return &cacheGetAllTracks, err
}

// Does a substring match on every non empty entry in tt.
func (i *Index) QueryTrack(tt source.TrackTags) (*[]source.TrackTags, error) {
	query :=
		`SELECT %s
 FROM Track tr
 JOIN Artist ar ON tr.trackartist = ar.ID
 JOIN Album  al ON tr.trackalbum  = al.ID
 WHERE tr.path LIKE '%%' || ? || '%%' AND tr.title LIKE '%%' || ? || '%%'
 AND tr.gentre LIKE '%%' || ? || '%%'
 AND ar.name LIKE ? || '%%' AND al.name LIKE ? || '%%'`

	if tt.Length != 0 {
		query += fmt.Sprintf(" AND tr.length=%d", tt.Length)
	}
	if tt.Year != 0 {
		query += fmt.Sprintf(" AND tr.year=%d", tt.Year)
	}
	if tt.Track != 0 {
		query += fmt.Sprintf(" AND tr.tracknumber=%d", tt.Track)
	}

	tx, err := i.db.Begin()
	if err != nil {
		return nil, err
	}

	var count int
	queryCount := fmt.Sprintf(query, "COUNT(path)")
	if err := tx.QueryRow(queryCount, tt.Path, tt.Title,
		tt.Artist, tt.Album).Scan(&count); err != nil {
		return nil, err
	}

	// allocating big enough array
	tracks := make([]source.TrackTags, count)

	querySelect := fmt.Sprintf(query, "tr.path, ar.name, al.name, tr.title,"+
		"tr.tracknumber, tr.length, tr.year tr.genre")
	rows, err := tx.Query(querySelect, tt.Path, tt.Title, tt.Genre, tt.Artist, tt.Album)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	err = rows2TrackList(rows, &tracks)
	return &tracks, err
}
