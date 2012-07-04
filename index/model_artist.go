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
	Id   int    `name:"ID" set:"0"`
	Name string `name:"name"`
}

// Wrappers for convinence.
func (a *Artists) All() (*[]Artist, error) {
	var ar []Artist
	err := a.Model.All(&ar)
	return &ar, err
}

func (a *Artists) Find(ID int) (*Artist, error) {
	var ar Artist
	err := a.Model.Find(&ar, ID)
	return &ar, err
}

func (a *Artists) Where(query Query, limit int) (*[]Artist, error) {
	var ar []Artist
	err := a.Model.Where(&ar, query, limit)
	return &ar, err
}
