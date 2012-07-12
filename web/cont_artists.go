/*  Copyright 2012, mokasin
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  c program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with c program. If not, see <http://www.gnu.org/licenses/>.
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
func NewControllerArtists(index *index.Index, route string) *ControllerArtists {
	return &ControllerArtists{
		Controller: *NewController(index, route, "artists"),
	}
}

// Implementation of SelectHandler.
func (c *ControllerArtists) Index(w http.ResponseWriter, r *http.Request) {

	// get first letter of artists
	letters, err := c.index.Artists.All().OrderBy("name").Letters("name")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := r.URL.Query().Get("page")

	// just go to the first page by default
	if page == "" {
		c.byFirstLetter(w, r, rune(letters[0]))
		return
	}

	// No request should contain more than 1 letter.
	if len(page) != 1 {
		http.NotFound(w, r)
		return
	}

	c.byFirstLetter(w, r, rune(page[0]))
}

func (c *ControllerArtists) Select(w http.ResponseWriter, r *http.Request, selector string) {
	id, err := strconv.Atoi(selector)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	artists, err := c.index.Artists.Find(id).Exec()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(*artists) == 0 {
		http.NotFound(w, r)
		return
	}

	artist := (*artists)[0]
	albums, err := artist.Albums().OrderBy("name").Exec()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Albums = make([]nameURL, len(*albums))

	// prepare structure for template
	for i := 0; i < len(*albums); i++ {
		c.Albums[i].Name = (*albums)[i].Name
		c.Albums[i].URL = "#"
	}

	c.Breadcrumb = Breadcrump(r.URL.Path)

	// render the website
	renderInPage(w, "index", c.Tmpl("artist"), c, artist.Name)
}

func (c *ControllerArtists) generatePager(letters, active string) {
	// creating pager
	c.Pager = make([]pager, len(letters))

	for i := 0; i < len(letters); i++ {
		if string(letters[i]) == active {
			c.Pager[i].Active = true
		}
		c.Pager[i].Label = string(letters[i])
		v := url.Values{}
		v.Add("page", string(letters[i]))
		c.Pager[i].Path = "/" + c.route + "?" + v.Encode()
	}
}

// firstLetter shows a list of all artists whom's name starting with letter.
func (c *ControllerArtists) byFirstLetter(w http.ResponseWriter, r *http.Request, letter rune) {

	// get first letter of artists
	letters, err := c.index.Artists.All().OrderBy("name").Letters("name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !strings.ContainsRune(letters, letter) {
		http.NotFound(w, r)
		return
	}

	// populating data
	artists, err := c.index.Artists.LikeQ(
		index.Query{"name": string(letter) + "%"},
	).Exec()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Artists = make([]nameURL, len(*artists))

	// prepare structure for template
	for i := 0; i < len(*artists); i++ {
		c.Artists[i].Name = (*artists)[i].Name
		c.Artists[i].URL = fmt.Sprintf("/%s/%d", c.route, (*artists)[i].Id)
	}

	c.generatePager(letters, string(letter))
	c.Breadcrumb = Breadcrump(r.URL.Path)

	// render the website
	renderInPage(w, "index", c.Tmpl("artists"), c,
		"Artists starting with"+string(letter))
}
