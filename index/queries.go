package index

import (
	"database/sql"
	"fmt"
	"musicrawler/source"
)

func rows2TrackList(rows *sql.Rows, array *[]source.TrackTags) error {
	var path, title, artist, album string
	var year, track uint

	i := 0
	for rows.Next() {
		if err := rows.Scan(&path, &title, &year, &track, &artist, &album); err == nil {
			(*array)[i] = source.TrackTags{
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

	return rows.Err()
}

// Returns list of source.TrackTags of all tracks in the database.
func (i *Index) GetAllTracks() (*[]source.TrackTags, error) {

	tx, err := i.db.Begin()
	if err != nil {
		return nil, err
	}

	var count int
	if err := tx.QueryRow("SELECT COUNT(path) FROM Track").Scan(&count); err != nil {
		return nil, err
	}

	// allocating big enough array
	tracks := make([]source.TrackTags, count)

	rows, err := tx.Query(
		`SELECT tr.path, tr.title, tr.year, tr.tracknumber, ar.name, al.name
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

	err = rows2TrackList(rows, &tracks)
	return &tracks, err
}

// Does a substring match on every non empty entry in tt.
func (i *Index) Query(tt source.TrackTags) (*[]source.TrackTags, error) {
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

	count, _ := i.QueryCount(tt)

	// allocating big enough array
	tracks := make([]source.TrackTags, count)

	rows, err := i.db.Query(query, tt.Path, tt.Title, tt.Artist, tt.Album)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	err = rows2TrackList(rows, &tracks)
	return &tracks, err
}

// Returns how many tracks with the given values are stored in the database
func (i *Index) QueryCount(tt source.TrackTags) (int, error) {
	query :=
		`SELECT COUNT(*)
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

	row := i.db.QueryRow(query, tt.Path, tt.Title, tt.Artist, tt.Album)

	var count int

	err := row.Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, err

}
