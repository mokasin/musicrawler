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
type Tracks struct {
	Model
}

func NewTracks(db *Database) *Tracks {
	// feed it with index and table name
	return &Tracks{Model: *NewModel(db, "track")}
}

func (self *Tracks) CreateTable() error {
	return self.db.Execute(`CREATE TABLE Track
	( ID          INTEGER NOT NULL PRIMARY KEY,
	  path        TEXT NOT NULL,
	  title       TEXT,
	  tracknumber INTEGER,
	  year        INTEGER,
	  length      INTEGER,
	  genre       TEXT,
	  album_id    INTEGER REFERENCES Album(ID) ON DELETE SET NULL,
	  filemtime	  INTEGER,
	  dbmtime     INTEGER
    );`)
}

// Define scheme of artist entry.
type Track struct {
	Id          int    `column:"ID" set:"0"`
	Path        string `column:"path"`
	Title       string `column:"title"`
	Tracknumber int    `column:"tracknumber"`
	Year        int    `column:"year"`
	Length      int    `column:"length"`
	Genre       string `column:"genre"`
	AlbumID     int    `column:"album_id"`
	Filemtime   int    `column:"filemtime"`
	DBMtime     int    `column:"dbmtime"`
}
