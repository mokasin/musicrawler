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
	"fmt"
	"musicrawler/source"
)

// Tracks represents the model of tracks in database.
type Tracks struct {
	index    *Index
	allCache cache
}

// Constructor returns instance of Tracks.
func NewTracks(i *Index) *Tracks {
	return &Tracks{index: i}
}

// columns that get queried
const columns = "tr.path, ar.name, al.name, tr.title, tr.tracknumber, tr.length, tr.year, tr.genre"

// Query does a SQL-query query with arguments args. query must be of the form
// 		SELECT %s FROM Tracks ...
// '%' in the query has to be escaped by '%%' (see fmt package)
func (t *Tracks) Query(query string, args ...interface{}) (*[]source.TrackTags, error) {
	tx, err := t.index.db.Begin()
	if err != nil {
		return nil, err
	}

	var count int
	if err := tx.QueryRow(fmt.Sprintf(query, "COUNT(*)"), args...).Scan(&count); err != nil {
		return nil, err
	}

	// allocating big enough array
	result := make([]source.TrackTags, count)

	rows, err := tx.Query(fmt.Sprintf(query, columns), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	var path, title, artist, album, genre string
	var year, track, length int

	i := 0
	for rows.Next() {
		if err := rows.Scan(&path, &artist, &album, &title,
			&track, &length, &year, &genre); err == nil {
			result[i] = source.TrackTags{
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

	return &result, rows.Err()
}

const tracks_sql_all = `SELECT %s
 FROM Track tr
 JOIN Artist ar ON tr.trackartist = ar.ID
 JOIN Album  al ON tr.trackalbum  = al.ID;`

// All returns a pointer to an array of source.TrackTags for all tracks in the
// database
func (t *Tracks) All() (*[]source.TrackTags, error) {
	if t.allCache.data == nil || t.allCache.ctime != t.index.Timestamp() {
		var err error
		t.allCache.data, err = t.Query(tracks_sql_all)
		if err != nil {
			return nil, err
		}
		t.allCache.ctime = t.index.Timestamp()
	}

	val, ok := t.allCache.data.(*[]source.TrackTags)
	if !ok {
		return t.Query(tracks_sql_all)
	}

	return val, nil
}

const track_sql_bytag = `SELECT %s
 FROM Track tr
 JOIN Artist ar ON tr.trackartist = ar.ID
 JOIN Album  al ON tr.trackalbum  = al.ID
 WHERE UPPER(tr.path) LIKE UPPER('%%' || ? || '%%')
 AND UPPER(tr.title) LIKE UPPER('%%' || ? || '%%')
 AND UPPER(tr.genre) LIKE UPPER('%%' || ? || '%%')
 AND UPPER(ar.name) LIKE UPPER(? || '%%')
 AND UPPER(al.name) LIKE UPPER(? || '%%')`

// ByTag return a pointer to an array of source.TrackTags for tracks filtered
// constrained by entries in tt. Empty filds of tt are considered as wildcards.
func (t *Tracks) ByTag(tt source.TrackTags) (*[]source.TrackTags, error) {
	query := track_sql_bytag

	if tt.Year != 0 {
		query += fmt.Sprintf(" AND tr.year=%d", tt.Year)
	}
	if tt.Length != 0 {
		query += fmt.Sprintf(" AND tr.length=%d", tt.Length)
	}
	if tt.Track != 0 {
		query += fmt.Sprintf(" AND tr.tracknumber=%d", tt.Track)
	}

	return t.Query(query, tt.Path, tt.Title, tt.Genre, tt.Artist, tt.Album)
}
