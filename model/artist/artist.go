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

package artist

import (
	"errors"
	. "musicrawler/lib/database"
	"musicrawler/lib/database/query"
)

func CreateArtistTable(db *Database) error {
	_, err := db.Execute(`CREATE TABLE Artist
	( ID   INTEGER  NOT NULL PRIMARY KEY,
	  name TEXT     UNIQUE
	);`)

	return err
}

// Define scheme of artist entry.
type Artist struct {
	Id   int64  `column:"ID" set:"0"`
	Name string `column:"name"`
	Link string
}

// Albums returns a prepared Query to query the albums of the artist.
func (self *Artist) AlbumsQuery(db *Database) *query.Query {
	return query.New(db, "album").Where("artist_id =", self.Id)
}

var ErrNoEntries = errors.New("No entries in database")

func FirstLetters(db *Database) (alpha, nonalpha string, err error) {
	// get first letter of artists
	q := query.New(db, "artist").Order("name")
	alpha, nonalpha, err = q.Letters("name")

	if err != nil {
		return
	}

	if len(alpha) == 0 && len(nonalpha) == 0 {
		err = ErrNoEntries
	}

	return
}

var alphabet = []interface{}{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O",
	"P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

func NonAlphaArtists(db *Database) *query.Query {
	return query.New(db, "artist").WhereIn(
		"SUBSTR(UPPER(name),1,1) NOT", alphabet...,
	)
}
