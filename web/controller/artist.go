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
	"code.google.com/p/gorilla/mux"
	"encoding/json"
	"musicrawler/lib/database/query"
	"musicrawler/lib/model/helper"
	"musicrawler/lib/web/controller"
	"musicrawler/lib/web/env"
	"musicrawler/lib/web/tmpl"
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
func NewArtist(env *env.Environment) *ControllerArtist {
	c := &ControllerArtist{
		controller.Controller: *controller.NewController(env),
	}

	c.Tmpl.AddTemplate("artist_index", "index", "artists")
	c.Tmpl.AddTemplate("artist_show", "index", "artist")

	return c
}

// Implementation of SelectHandler.
func (self *ControllerArtist) Index(w http.ResponseWriter, r *http.Request) {
	al, nal, err := artist.FirstLetters(self.Env.Db)
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
		page = string(letters[0])
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
		err = artist.NonAlphaArtists(self.Env.Db).Exec(&artists)
	} else {
		err = query.New(self.Env.Db, "artist").
			Like("name", page+"%").Exec(&artists)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// prepare data for template
	for i := 0; i < len(artists); i++ {
		url, err := self.URL("artist", controller.Pairs{"id": artists[i].Id})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		artists[i].Link = url
	}

	url, err := self.URL("artist_base", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager := helper.NewPager(url, strings.Split(letters, ""), page)

	self.Tmpl.AddDataToTemplate("artist_index", "Artists", artists)
	self.Tmpl.AddDataToTemplate("artist_index", "Pager", pager)

	// render the website
	self.Tmpl.RenderPage(
		w,
		"artist_index",
		&tmpl.Page{Title: "Artists starting with " + string(page)},
	)
}

func (self *ControllerArtist) IndexJSON(w http.ResponseWriter, r *http.Request) {
	var artists []artist.Artist

	err := query.New(self.Env.Db, "artist").Exec(&artists)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(artists)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func (self *ControllerArtist) Show(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := self.Env.Db.BeginTransaction(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer self.Env.Db.EndTransaction()

	// retreive artist by id
	var artist artist.Artist

	err = query.New(self.Env.Db, "artist").Find(id).Exec(&artist)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// retreive albums of artist
	var albums []album.Album

	err = artist.AlbumsQuery(self.Env.Db).Order("name").Exec(&albums)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// prepare data for template
	for i := 0; i < len(albums); i++ {
		url, err := self.URL("album", controller.Pairs{"id": albums[i].Id})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		albums[i].Link = url
	}

	self.Tmpl.AddDataToTemplate("artist_show", "Artist", &artist)
	self.Tmpl.AddDataToTemplate("artist_show", "Albums", &albums)

	backlink, _ := self.URL("artist_base", nil)

	// render the website
	self.Tmpl.RenderPage(
		w,
		"artist_show",
		&tmpl.Page{Title: artist.Name, BackLink: backlink},
	)
}

func (self *ControllerArtist) ShowJSON(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// retreive artist by id
	var artist artist.Artist

	err = query.New(self.Env.Db, "artist").Find(id).Exec(&artist)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(artist)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func (self *ControllerArtist) AlbumsJSON(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// retreive artist by id
	var artist artist.Artist

	err = query.New(self.Env.Db, "artist").Find(id).Exec(&artist)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// retreive albums of artist
	var albums []album.Album

	err = artist.AlbumsQuery(self.Env.Db).Order("name").Exec(&albums)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(albums)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(b)
}
