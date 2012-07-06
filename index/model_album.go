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
type Albums struct {
	Model
}

func NewAlbums(index *Index) *Albums {
	// feed it with index and table name
	return &Albums{Model: *NewModel(index, "album")}
}

// Define scheme of artist entry.
type Album struct {
	Id       int    `column:"ID" set:"0"`
	Name     string `column:"name"`
	ArtistID int    `column:"artist_id"`

	Index *Index
}

func (a *Album) Artist() (*Artist, error) {
	artists, err := a.Index.Artists.Find(a.ArtistID).Exec()

	if len(*artists) > 0 {
		return &((*artists)[0]), err
	}

	return nil, err
}

func (a *Album) Tracks() *Tracks {
	return a.Index.Tracks.Where("album_id = ?", a.Id)
}

func (a *Albums) Exec() (*[]Album, error) {
	var ar []Album
	err := a.Model.Exec(&ar)
	return &ar, err
}

// Wrappers for convinence.
func (a *Albums) All() *Albums {
	a.Model.All()
	return a
}

func (a *Albums) Find(ID int) *Albums {
	a.Model.Find(ID)
	return a
}

func (a *Albums) Where(query string, args ...interface{}) *Albums {
	a.Model.Where(query, args...)
	return a
}

func (a *Albums) WhereQ(query Query) *Albums {
	a.Model.WhereQ(query)
	return a
}

func (a *Albums) LikeQ(query Query) *Albums {
	a.Model.LikeQ(query)
	return a
}

func (a *Albums) Limit(number int) *Albums {
	a.Model.Limit(number)
	return a
}

func (a *Albums) OrderBy(column string) *Albums {
	a.Model.OrderBy(column)
	return a
}
