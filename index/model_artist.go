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

func NewArtists(index *Index) *Artists {
	// feed it with index and table name
	return &Artists{Model: *NewModel(index, "artist")}
}

func (a *Artists) CreateDatabase() error {
	return a.Execute(`CREATE TABLE Artist
	( ID   INTEGER  NOT NULL PRIMARY KEY,
	  name TEXT     UNIQUE
	);`)
}

// Define scheme of artist entry.
type Artist struct {
	Item

	Id   int    `column:"ID" set:"0"`
	Name string `column:"name"`
}

func (a *Artist) Albums() *Albums {
	return a.Index.Albums.Where("artist_id = ?", a.Id)
}

func (a *Artists) Exec() (*[]Artist, error) {
	var ar []Artist
	err := a.Model.Exec(&ar)
	return &ar, err
}

// Wrappers for convinence and type safety.
func (a *Artists) All() *Artists {
	a.Model.All()
	return a
}

func (a *Artists) Find(ID int) *Artists {
	a.Model.Find(ID)
	return a
}

func (a *Artists) Where(query string, args ...interface{}) *Artists {
	a.Model.Where(query, args...)
	return a
}

func (a *Artists) WhereQ(query Query) *Artists {
	a.Model.WhereQ(query)
	return a
}

func (a *Artists) LikeQ(query Query) *Artists {
	a.Model.LikeQ(query)
	return a
}

func (a *Artists) Limit(number int) *Artists {
	a.Model.Limit(number)
	return a
}

func (a *Artists) Offset(offset int) *Artists {
	a.Model.Offset(offset)
	return a
}

func (a *Artists) OrderBy(column string) *Artists {
	a.Model.OrderBy(column)
	return a
}
