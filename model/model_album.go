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

package model

import (
	. "musicrawler/lib/database"
	. "musicrawler/lib/database/query"
)

func CreateAlbumTable(db *Database) error {
	_, err := db.Execute(`CREATE TABLE Album
	(  ID   INTEGER NOT NULL PRIMARY KEY,
	   name TEXT,
	   artist_id INTEGER REFERENCES Artist(ID) ON DELETE SET NULL
	);`)

	if err != nil {
		return err
	}

	// create tuple index to prevent double entries
	_, err = db.Execute(
		"CREATE UNIQUE INDEX 'album_artist' ON Album (name, artist_id);")
	return err
}

// Define scheme of album entry.
type Album struct {
	Id       int64  `column:"ID" set:"0"`
	Name     string `column:"name"`
	ArtistID int64  `column:"artist_id"`
}

func (self *Album) ArtistQuery(db *Database) *Query {
	return NewQuery(db, "artist").Where("ID =", self.ArtistID)
}

// Tracks returns a prepared Query reference 
func (self *Album) TracksQuery(db *Database) *Query {
	return NewQuery(db, "track").Where("album_id =", self.Id)
}
