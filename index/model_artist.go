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

import (
	"fmt"
)

type Artists struct {
	Model
}

func NewArtists(index *Index) *Artists {
	return &Artists{Model: *NewModel(index, "artist")}
}

type Artist struct {
	Id   int    `name:"ID" create:"0"`
	Name string `name:"name"`
}

// just for fun
func (m *Artist) String() string {
	return fmt.Sprintf("Artist{id -> %03d, name -> %s}", m.Id, m.Name)
}

// All the following is just about type conversion. This sucks.
func (a *Artists) toArtists(src []Result, err error) ([]*Artist, error) {
	if err != nil {
		return nil, err
	}

	dest := make([]*Artist, len(src))

	for i := 0; i < len(dest); i++ {
		dest[i] = new(Artist)
		err = a.Decode(src[i], dest[i])
		if err != nil {
			return nil, err
		}
	}

	return dest, nil
}

func (a *Artists) toArtist(src Result, err error) (*Artist, error) {
	if err != nil {
		return nil, err
	}

	dest := new(Artist)
	err = a.Decode(src, dest)
	if err != nil {
		return nil, err
	}

	return dest, nil
}

func (a *Artists) All() ([]*Artist, error) {
	return a.toArtists(a.Model.All())
}

func (a *Artists) Find(ID int) (*Artist, error) {
	return a.toArtist(a.Model.Find(ID))
}

func (a *Artists) Where(query Query) ([]*Artist, error) {
	return a.toArtists(a.Model.Where(query))
}
