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
	"html/template"
	"musicrawler/index"
	"musicrawler/source"
	"net/http"
	//	"strconv"
)

// Model of a simple pager.
type pager struct {
	Label  string
	Path   string
	Active bool
}

// Controller to serve all tracks
type controllerAllTracks struct {
	index  *index.Index
	tmpl   *template.Template
	Tracks []source.TrackTags
	Pager  []pager
}

// Constructor.
func NewControllerAllTracks(index *index.Index) *controllerAllTracks {
	return &controllerAllTracks{index: index}
}

// Parses and returns template with name name.
func (c *controllerAllTracks) Tmpl(name string) *template.Template {
	if c.tmpl == nil {
		c.tmpl = template.Must(
			template.ParseFiles(websitePath + "templates/" + name + ".html"))
	}
	return c.tmpl.Lookup(name + ".html")
}

// Shows unsorted list of all tracks in database.
func (c *controllerAllTracks) Handler(w http.ResponseWriter, r *http.Request) {
	// Only show tha) many tracks on one page
	const shownTracks = 100

	l, err := c.index.Tracks.All()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var pagestring string
	var pagenum int

	// parse URL for pagenumber and validate
	if _, err := fmt.Sscanf(r.RequestURI, "/%s", &pagestring); err != nil {
		pagenum = 0
		pagestring = "A"
	} else {
		//		pagenum, err = strconv.Atoi(pagestring)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}
	if pagenum < 0 || pagenum > len(*l)/shownTracks {
		http.NotFound(w, r)
		return
	}

	// populating data
	var artistmap map[string][]string
	//  var err error
	var tracks *[]source.TrackTags
	artistmap, err = c.index.Artists.FirstLetterMap()
	c.Tracks = make([]source.TrackTags, 0)

	if len(artistmap[pagestring]) > 0 {
		for i := 0; i < len(artistmap[pagestring]); i++ {
			tracks, err = c.index.Tracks.ByArtist(artistmap[pagestring][i])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			c.Tracks = append(c.Tracks, *tracks...)
		}
	}

	letters, err := c.index.Artists.Letters()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Pager = make([]pager, len(letters))
	for i := 0; i < len(letters); i++ {
		if string(letters[i]) == pagestring {
			c.Pager[i].Active = true
		}
		c.Pager[i].Label = string(letters[i])
		c.Pager[i].Path = string(letters[i])
	}

	renderInPage(w, "index", c.Tmpl("alltracks"), c, &Page{Title: "musicrawler"})
}
