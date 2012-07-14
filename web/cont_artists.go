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
	"net/url"
	"strconv"
	"strings"
)

type nameURL struct {
	Name string
	URL  string
}

// Controller to serve artists
type ControllerArtists struct {
	Controller

	Artists    []nameURL
	Albums     []nameURL
	Pager      []pager
	Breadcrumb []pager
}

// Constructor.
func NewControllerArtists(db *index.Database, route string) *ControllerArtists {
	return &ControllerArtists{
		Controller: *NewController(db, route, "artists", "artist"),
	}
}

// Implementation of SelectHandler.
func (self *ControllerArtists) Index(w http.ResponseWriter, r *http.Request) {

	// get first letter of artists
	q := index.NewQuery("artist").Order("name")

	letters, err := self.db.Artists.Letters(q, "name")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := r.URL.Query().Get("page")

	// just go to the first page by default
	if page == "" {
		self.byFirstLetter(w, r, rune(letters[0]))
		return
	}

	// No request should contain more than 1 letter.
	if len(page) != 1 {
		http.NotFound(w, r)
		return
	}

	self.byFirstLetter(w, r, rune(page[0]))
}

func (self *ControllerArtists) Select(w http.ResponseWriter, r *http.Request, selector string) {
	id, err := strconv.Atoi(selector)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var artists []index.Artist
	q := index.NewQuery("artist").Find(id)

	err = self.db.Artists.Exec(q, &artists)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(artists) == 0 {
		http.NotFound(w, r)
		return
	}

	artist := artists[0]

	q = artist.AlbumsQuery().Order("name")

	var albums []index.Album

	err = self.db.Albums.Exec(q, &albums)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	self.Albums = make([]nameURL, len(albums))

	// prepare structure for template
	for i := 0; i < len(albums); i++ {
		self.Albums[i].Name = albums[i].Name
		self.Albums[i].URL = "#"
	}

	self.Breadcrumb = Breadcrump(r.URL.Path)

	// render the website
	renderInPage(w, "index", self.Tmpl("artist"), self, artist.Name)
}

func (self *ControllerArtists) generatePager(letters, active string) {
	// creating pager
	self.Pager = make([]pager, len(letters))

	for i := 0; i < len(letters); i++ {
		if string(letters[i]) == active {
			self.Pager[i].Active = true
		}
		self.Pager[i].Label = string(letters[i])
		v := url.Values{}
		v.Add("page", string(letters[i]))
		self.Pager[i].Path = "/" + self.route + "?" + v.Encode()
	}
}

// firstLetter shows a list of all artists whom's name starting with letter.
func (self *ControllerArtists) byFirstLetter(w http.ResponseWriter, r *http.Request, letter rune) {
	// get first letter of artists
	q := index.NewQuery("artist").Order("name")
	letters, err := self.db.Artists.Letters(q, "name")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !strings.ContainsRune(letters, letter) {
		http.NotFound(w, r)
		return
	}

	// populating data
	var artists []index.Artist

	q = index.NewQuery("artist").Like("name", string(letter)+"%")

	err = self.db.Artists.Exec(q, &artists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	self.Artists = make([]nameURL, len(artists))

	// prepare structure for template
	for i := 0; i < len(artists); i++ {
		self.Artists[i].Name = artists[i].Name
		self.Artists[i].URL = fmt.Sprintf("/%s/%d", self.route, artists[i].Id)
	}

	self.generatePager(letters, string(letter))
	self.Breadcrumb = Breadcrump(r.URL.Path)

	// render the website
	renderInPage(w, "index", self.Tmpl("artists"), self,
		"Artists starting with"+string(letter))
}
