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

	Artists []artistData
	Pager   []pager
}

// Constructor.
func NewControllerArtists(index *index.Index) *ControllerArtists {
	return &ControllerArtists{Controller: *NewController(index)}
}

func (c *ControllerArtists) Index(w http.ResponseWriter, r *http.Request) {
	// get first letter of artists
	letters, err := c.index.Artists.Letters()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Select(w, r, string(letters[0]))
}

// Shows a list of trakcs sorted by artists und paged with first letter of
// artist.
func (c *ControllerArtists) Select(w http.ResponseWriter, r *http.Request, selector string) {
	// get first letter of artists
	letters, err := c.index.Artists.Letters()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(selector) > 1 {
		http.NotFound(w, r)
		return
	}

	if !strings.Contains(letters, selector) {
		http.NotFound(w, r)
		return
	}

	// populating data
	artists, err := c.index.Artists.ByFirstLetter(rune(selector[0]))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Artists = make([]artistData, len(*artists))

	rep := strings.NewReplacer("/", "")
	for i := 0; i < len(*artists); i++ {
		c.Artists[i].Name = (*artists)[i]

		// remove some chars from URL
		c.Artists[i].URL = rep.Replace((*artists)[i])
	}

	// creating pager
	c.Pager = make([]pager, len(letters))

	for i := 0; i < len(letters); i++ {
		if string(letters[i]) == selector {
			c.Pager[i].Active = true
		}
		c.Pager[i].Label = string(letters[i])
		c.Pager[i].Path = string(letters[i])
	}

	// render the website
	renderInPage(w, "index", c.Tmpl("artists"), c, &Page{Title: "musicrawler"})
}
