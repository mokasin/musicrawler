/*  Copyright 2012, mokasin
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with the program. If not, see <http://www.gnu.org/licenses/>.
 */

package web

import (
	"fmt"
	"musicrawler/index"
	"net/http"
	"strconv"
)

// Controller to serve artists
type ControllerAlbums struct {
	Controller
}

// Constructor.
func NewControllerAlbums(db *index.Database, route string) *ControllerAlbums {
	c := &ControllerAlbums{
		Controller: *NewController(db, route),
	}

	c.AddTemplate("select", "index", "album")

	return c
}

// Implementation of SelectHandler.
func (self *ControllerAlbums) Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Not implemented yet.")
}

func (self *ControllerAlbums) Select(w http.ResponseWriter, r *http.Request, selector string) {
	id, err := strconv.Atoi(selector)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := self.db.BeginTransaction(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer self.db.EndTransaction()

	var album index.Album
	q := index.NewQuery(self.db, "album").Find(id)

	err = q.Exec(&album)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render the website
	self.renderPage(
		w,
		"select",
		&Page{Title: album.Name},
		nil,
	)
}
