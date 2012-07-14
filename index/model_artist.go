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

// Define artist model.
type Artists struct {
	Model
}

func NewArtists(db *Database) *Artists {
	// feed it with index and table name
	return &Artists{Model: *NewModel(db, "artist")}
}

func (self *Artists) CreateTable() error {
	return self.db.Execute(`CREATE TABLE Artist
	( ID   INTEGER  NOT NULL PRIMARY KEY,
	  name TEXT     UNIQUE
	);`)
}

// Define scheme of artist entry.
type Artist struct {
	Id   int    `column:"ID" set:"0"`
	Name string `column:"name"`
}

// Albums returns a prepared Query to query the albums of the artist.
func (self *Artist) AlbumsQuery() *Query {
	q := NewQuery("album")
	return q.Where("artist_id =", self.Id)
}
