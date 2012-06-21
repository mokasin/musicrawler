package index

import (
	"container/list"
	"database/sql"
	"fmt"
	"musicrawler/source"
)

func rows2TrackList(rows *sql.Rows) (*list.List, error) {
	l := list.New()

	var path, title, artist, album string
	var year, track uint

	for rows.Next() {
		if err := rows.Scan(&path, &title, &year, &track, &artist, &album); err == nil {
			l.PushBack(source.TrackTags{
				Path:    path,
				Title:   title,
				Artist:  artist,
				Album:   album,
				Comment: "", // not yet in database
				Genre:   "",
				Year:    year,
				Track:   track,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return l, nil
}

// Returns list of source.TrackTags of all tracks in the database.
func (i *Index) GetAllTracks() (*list.List, error) {
	rows, err := i.db.Query(
		`SELECT tr.path, tr.title, tr.year, tr.tracknumber, ar.name, al.name
 FROM Track tr
 JOIN Artist ar ON tr.trackartist = ar.ID
 JOIN Album  al ON tr.trackalbum  = al.ID;`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rows2TrackList(rows)
}

// Does a substring match on every non empty entry in tt.
func (i *Index) Query(tt source.TrackTags) (*list.List, error) {
	query :=
		`SELECT tr.path, tr.title, tr.year, tr.tracknumber, ar.name, al.name
 FROM Track tr
 JOIN Artist ar ON tr.trackartist = ar.ID
 JOIN Album  al ON tr.trackalbum  = al.ID
 WHERE tr.path LIKE '%' || ? || '%' AND tr.title LIKE '%' || ? || '%'
 AND ar.name LIKE '%' || ? || '%' AND al.name LIKE '%' || ? || '%'`

	if tt.Year != 0 {
		query += fmt.Sprintf(" AND tr.year=%d", tt.Year)
	}
	if tt.Track != 0 {
		query += fmt.Sprintf(" AND tr.tracknumber=%d", tt.Track)
	}

	rows, err := i.db.Query(query, tt.Path, tt.Title, tt.Artist, tt.Album)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rows2TrackList(rows)
}
