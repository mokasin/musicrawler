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

type artistLink struct {
	Artist index.Artist
	Path   string
}

type artistsIndexTmpl struct {
	Pager   []activelink
	Artists []artistLink
}

type albumLink struct {
	Album index.Album
	Path  string
}

type artistsSelectTmpl struct {
	Breadcrumb []activelink
	Albums     []albumLink
}

// Controller to serve artists
type ControllerArtists struct {
	Controller
}

// Constructor.
func NewControllerArtists(db *index.Database, route string) *ControllerArtists {
	c := &ControllerArtists{
		Controller: *NewController(db, route),
	}

	c.AddTemplate("index", "index", "artists")
	c.AddTemplate("select", "index", "artist")

	return c
}

// Implementation of SelectHandler.
func (self *ControllerArtists) Index(w http.ResponseWriter, r *http.Request) {

	// get first letter of artists
	q := index.NewQuery(self.db, "artist").Order("name")

	letters, err := q.Letters("name")

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

	if err := self.db.BeginTransaction(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer self.db.EndTransaction()

	var artist index.Artist
	q := index.NewQuery(self.db, "artist").Find(id)
	err = q.Exec(&artist)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	q = artist.AlbumsQuery(self.db).Order("name")

	var albums []index.Album

	err = q.Exec(&albums)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var td artistsSelectTmpl
	td.Albums = make([]albumLink, len(albums))

	// prepare structure for template
	for i := 0; i < len(albums); i++ {
		td.Albums[i].Album = albums[i]
		//TODO don't hard code pathes
		td.Albums[i].Path = fmt.Sprintf("/%s/%d", "album", albums[i].Id)
	}

	td.Breadcrumb = Breadcrump(r.URL.Path)

	// render the website
	self.renderPage(
		w,
		"select",
		&Page{Title: artist.Name},
		td,
	)
}

func (self *ControllerArtists) generatePager(letters, active string) []activelink {
	// creating pager
	pager := make([]activelink, len(letters))

	for i := 0; i < len(letters); i++ {
		if string(letters[i]) == active {
			pager[i].Active = true
		}
		pager[i].Label = string(letters[i])
		v := url.Values{}
		v.Add("page", string(letters[i]))
		pager[i].Path = "/" + self.route + "?" + v.Encode()
	}

	return pager
}

// firstLetter shows a list of all artists whom's name starting with letter.
func (self *ControllerArtists) byFirstLetter(w http.ResponseWriter, r *http.Request, letter rune) {
	if err := self.db.BeginTransaction(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer self.db.EndTransaction()

	// get first letter of artists
	q := index.NewQuery(self.db, "artist").Order("name")
	letters, err := q.Letters("name")

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

	q = index.NewQuery(self.db, "artist").Like("name", string(letter)+"%")

	err = q.Exec(&artists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var td artistsIndexTmpl
	td.Artists = make([]artistLink, len(artists))

	// prepare structure for template
	for i := 0; i < len(artists); i++ {
		td.Artists[i].Artist = artists[i]
		td.Artists[i].Path = fmt.Sprintf("/%s/%d", self.route, artists[i].Id)
	}

	td.Pager = self.generatePager(letters, string(letter))

	// render the website
	self.renderPage(
		w,
		"index",
		&Page{Title: "Artists starting with" + string(letter)},
		td,
	)
}