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

// Define scheme of artist entry.
type Artist struct {
	Id   int    `name:"ID" set:"0"`
	Name string `name:"name"`
}

func (a *Artists) Exec() (*[]Artist, error) {
	var ar []Artist
	err := a.Model.Exec(&ar)
	return &ar, err
}

// Wrappers for convinence.
func (a *Artists) All() *Artists {
	a.Model.All()
	return a
}

func (a *Artists) Find(ID int) *Artists {
	a.Model.Find(ID)
	return a
}

func (a *Artists) Where(query Query) *Artists {
	a.Model.Where(query)
	return a
}

func (a *Artists) Like(query Query) *Artists {
	a.Model.Like(query)
	return a
}

func (a *Artists) Limit(number int) *Artists {
	a.Model.Limit(number)
	return a
}

func (a *Artists) OrderBy(column string) *Artists {
	a.Model.OrderBy(column)
	return a
}
