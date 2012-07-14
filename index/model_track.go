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

import "fmt"

func CreateTrackTable(db *Database) error {
	return db.Execute(`CREATE TABLE Track
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

// Define scheme of track entry.
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

type JoinedTrack struct {
	Id          int    `column:"track:ID" set:"0"`
	Path        string `column:"track:path"`
	Title       string `column:"track:title"`
	Tracknumber int    `column:"track:tracknumber"`
	Year        int    `column:"track:year"`
	Length      int    `column:"track:length"`
	Genre       string `column:"track:genre"`
	Filemtime   int    `column:"track:filemtime"`
	DBMtime     int    `column:"track:dbmtime"`
	Artist      string `column:"artist:name"`
	Album       string `column:"album:name"`
}

func (self *Track) AlbumQuery(db *Database) *Query {
	return NewQuery(db, "album").Where("ID =", self.AlbumID)
}

// LengthString returns a nicely formatted string of the track's length.
func (self *Track) LengthString() string {
	return fmt.Sprintf("%d:%02d", self.Length/60, self.Length%60)
}

// LengthString returns a nicely formatted string of the track's length.
func (self *JoinedTrack) LengthString() string {
	return fmt.Sprintf("%d:%02d", self.Length/60, self.Length%60)
}
