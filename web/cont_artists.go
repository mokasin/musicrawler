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
	"musicrawler/index"
	"net/http"
	"strings"
)

type artistData struct {
	Name string
	URL  string
}

// Controller to serve artists
type ControllerArtists struct {
	Controller

	Artists    []artistData
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

	c.Select(w, r, string(letters[0]))
}

// Select shows a list of tracks sorted by artists und paged with first letter of
// artist. Implementation of SelectHandler.
func (c *ControllerArtists) Select(w http.ResponseWriter, r *http.Request, selector string) {

	tokens := ParseURL(selector)

	if len(tokens[0]) > 1 {
		http.NotFound(w, r)
		return
	}
	switch len(tokens) {
	case 1:
		c.byFirstLetter(w, r, rune(tokens[0][0]))
	case 2:
		c.byName(w, r, tokens[1])
	default:
		http.NotFound(w, r)
		return

	}
}

func (c *ControllerArtists) generatePager(letters, active string) {
	// creating pager
	c.Pager = make([]pager, len(letters))

	for i := 0; i < len(letters); i++ {
		if string(letters[i]) == active {
			c.Pager[i].Active = true
		}
		c.Pager[i].Label = string(letters[i])
		c.Pager[i].Path = "/" + c.route + "/" + string(letters[i])
	}
}

// generateLinkToArtist takes an artist's name and spit out the link to it.
func (c *ControllerArtists) linkToArtist(name string) string {
	var letter string

	// Remove slashes from artist's name
	rep := strings.NewReplacer("/", "")

	name = rep.Replace(name)

	if len(name) != 0 {
		letter = strings.ToUpper(string(name[0])) + "/"
	}

	return "/" + c.route + "/" + letter + name
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

	c.Artists = make([]artistData, len(*artists))

	// prepare structure for template
	for i := 0; i < len(*artists); i++ {
		c.Artists[i].Name = (*artists)[i].Name
		c.Artists[i].URL = c.linkToArtist((*artists)[i].Name)
	}

	c.generatePager(letters, string(letter))
	c.Breadcrumb = Breadcrump(r.URL.Path)

	// render the website
	renderInPage(w, "index", c.Tmpl("artists"), c,
		"Artists starting with"+string(letter))
}

func (c *ControllerArtists) byName(w http.ResponseWriter, r *http.Request, name string) {
}
