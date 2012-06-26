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

type tracksCache struct {
	data  *[]source.TrackTags
	ctime int64
}

// Tracks represents the model of tracks in database.
type Tracks struct {
	index    *Index
	allCache tracksCache
}

// Constructor returns instance of Tracks.
func NewTracks(i *Index) *Tracks {
	return &Tracks{index: i}
}

// columns that get queried
const columns = "tr.path, tr.title, tr.year, tr.tracknumber, ar.name, al.name"

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

	var path, title, artist, album string
	var year, track uint

	i := 0
	for rows.Next() {
		if err := rows.Scan(&path, &title, &year, &track, &artist, &album); err == nil {
			result[i] = source.TrackTags{
				Path:    path,
				Title:   title,
				Artist:  artist,
				Album:   album,
				Comment: "", // not yet in database
				Genre:   "",
				Year:    year,
				Track:   track,
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
	if t.allCache.data != nil && t.allCache.ctime == t.index.timestamp {
		return t.allCache.data, nil
	}

	return t.Query(tracks_sql_all)
}

const track_sql_bytag = `SELECT %s
 FROM Track tr
 JOIN Artist ar ON tr.trackartist = ar.ID
 JOIN Album  al ON tr.trackalbum  = al.ID
 WHERE tr.path LIKE '%%' || ? || '%%' AND tr.title LIKE '%%' || ? || '%%'
 AND ar.name LIKE ? || '%%' AND al.name LIKE ? || '%%'`

// ByTag return a pointer to an array of source.TrackTags for tracks filtered
// constrained by entries in tt. Empty filds of tt are considered as wildcards.
func (t *Tracks) ByTag(tt source.TrackTags) (*[]source.TrackTags, error) {
	query := track_sql_bytag

	if tt.Year != 0 {
		query += fmt.Sprintf(" AND tr.year=%d", tt.Year)
	}
	if tt.Track != 0 {
		query += fmt.Sprintf(" AND tr.tracknumber=%d", tt.Track)
	}

	return t.Query(query)
}
