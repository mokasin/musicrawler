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

package controller

import (
	"fmt"
	"musicrawler/lib/database"
	"musicrawler/lib/database/query"
	"musicrawler/lib/model/helper"
	"musicrawler/lib/web/controller"
	"musicrawler/lib/web/router"
	"musicrawler/model/album"
	"musicrawler/model/artist"
	"net/http"
	"strconv"
	"strings"
)

// Controller to serve artists
type ControllerArtist struct {
	controller.Controller
}

// Constructor.
func NewArtist(db *database.Database, router *router.Router, filepath string) *ControllerArtist {
	c := &ControllerArtist{
		controller.Controller: *controller.NewController(db, router, filepath),
	}

	c.AddTemplate("index", "index", "artists")
	c.AddTemplate("show", "index", "artist")

	return c
}

// Implementation of SelectHandler.
func (self *ControllerArtist) Index(w http.ResponseWriter, r *http.Request) {
	al, nal, err := artist.FirstLetters(self.Db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	letters := al

	// use zero to represent non alphabetic letters
	if len(nal) != 0 {
		letters = "0" + al
	}

	page := r.URL.Query().Get("page")

	switch {
	case page == "":
		// just go to the first page by default
		page = string(al[0])
	case len(page) != 1:
		// No request should contain more than 1 letter.
		http.NotFound(w, r)
		return
	}

	// url validation
	if !strings.ContainsAny(letters, page) {
		http.NotFound(w, r)
		return
	}

	// populating data
	var artists []artist.Artist

	if string(page) == "0" {
		err = artist.NonAlphaArtists(self.Db).Exec(&artists)
	} else {
		err = query.New(self.Db, "artist").
			Like("name", page+"%").Exec(&artists)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// prepare structure for template
	for i := 0; i < len(artists); i++ {
		artists[i].Link = fmt.Sprintf("%s/%d", self.Route(), artists[i].Id)
	}

	pager := helper.NewPager(self.Route(), strings.Split(letters, ""), page)

	self.AddDataToTemplate("index", "Artists", artists)
	self.AddDataToTemplate("index", "Pager", pager)

	// render the website
	self.RenderPage(
		w,
		"index",
		&controller.Page{Title: "Artists starting with " + string(page)},
	)
}

func (self *ControllerArtist) Show(w http.ResponseWriter, r *http.Request, selector string) {
	id, err := strconv.Atoi(selector)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := self.Db.BeginTransaction(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer self.Db.EndTransaction()

	var artist artist.Artist
	err = query.New(self.Db, "artist").Find(id).Exec(&artist)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err == query.ErrNoResults {
		http.NotFound(w, r)
		return
	}

	var albums []album.Album

	err = artist.AlbumsQuery(self.Db).Order("name").Exec(&albums)
	if err != nil && err != query.ErrNoResults {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := 0; i < len(albums); i++ {
		albums[i].Link = fmt.Sprintf("%s/%d",
			self.Router.GetRouteOf("album"), albums[i].Id)
	}

	self.AddDataToTemplate("show", "Arist", &artist)
	self.AddDataToTemplate("show", "Albums", &albums)

	// render the website
	self.RenderPage(
		w,
		"show",
		&controller.Page{Title: artist.Name},
	)
}
