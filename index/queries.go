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

	err = rows2TrackList(rows, &cacheGetAllTracks)
	return &cacheGetAllTracks, err
}

// Return an array of all artist names
func (i *Index) GetAllArtists() (*[]string, error) {
	query := "select %s from Artist;"

	var count int

	tx, err := i.db.Begin()

	err = tx.QueryRow(fmt.Sprintf(query, "COUNT(*)")).Scan(&count)

	if err != nil {
		return nil, err
	}

	artists := make([]string, count)

	rows, err := tx.Query(fmt.Sprintf(query, "name"))

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	err = tx.Commit()

	var artist string

	c := 0
	for rows.Next() {
		if err = rows.Scan(&artist); err != nil {
			return nil, err
		}
		artists[c] = artist
		c++
	}
	return &artists, err
}

// Does a substring match on every non empty entry in tt.
func (i *Index) QueryTrack(tt source.TrackTags) (*[]source.TrackTags, error) {
	query :=
		`SELECT %s
 FROM Track tr
 JOIN Artist ar ON tr.trackartist = ar.ID
 JOIN Album  al ON tr.trackalbum  = al.ID
 WHERE tr.path LIKE '%%' || ? || '%%' AND tr.title LIKE '%%' || ? || '%%'
 AND ar.name LIKE ? || '%%' AND al.name LIKE ? || '%%'`

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

	querySelect := fmt.Sprintf(query,
		"tr.path, tr.title, tr.year, tr.tracknumber, ar.name, al.name")
	rows, err := tx.Query(querySelect, tt.Path, tt.Title, tt.Artist, tt.Album)
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
