/*  Copyright 2012, mokasin
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR t PARTICULAR PURPOSE. See the
 *  GNU General Public License for more details.
 *
 *  You should have received t copy of the GNU General Public License
 *  along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package index

// Define artist model.
type Tracks struct {
	Model
}

func NewTracks(index *Index) *Tracks {
	// feed it with index and table name
	return &Tracks{Model: *NewModel(index, "track")}
}

func (t *Tracks) CreateDatabase() error {
	return t.Execute(`CREATE TABLE Track
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
	Item

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

func (t *Track) Album() (*Album, error) {
	albums, err := t.Index.Albums.Find(t.AlbumID).Exec()

	if len(*albums) > 0 {
		return &((*albums)[0]), err
	}

	return nil, err
}

func (t *Track) Artist() (*Artist, error) {
	ar, err := t.Album()
	if err != nil {
		return nil, err
	}
	return ar.Artist()
}

func (t *Tracks) Exec() (*[]Track, error) {
	var ar []Track
	err := t.Model.Exec(&ar)
	return &ar, err
}

// Wrappers for convinence.
func (t *Tracks) All() *Tracks {
	t.Model.All()
	return t
}

func (t *Tracks) Find(ID int) *Tracks {
	t.Model.Find(ID)
	return t
}

func (t *Tracks) Where(query string, args ...interface{}) *Tracks {
	t.Model.Where(query, args...)
	return t
}

func (t *Tracks) WhereQ(query Query) *Tracks {
	t.Model.WhereQ(query)
	return t
}

func (t *Tracks) LikeQ(query Query) *Tracks {
	t.Model.LikeQ(query)
	return t
}

func (t *Tracks) Limit(number int) *Tracks {
	t.Model.Limit(number)
	return t
}

func (t *Tracks) Offset(offset int) *Tracks {
	t.Model.Offset(offset)
	return t
}

func (t *Tracks) OrderBy(column string) *Tracks {
	t.Model.OrderBy(column)
	return t
}
