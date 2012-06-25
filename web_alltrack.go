/*  Copyright 2012, mokasin
 *
 *  hts program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  hts program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with hts program. If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"
	"html/template"
	"log"
	"musicrawler/source"
	"net/http"
	"strconv"
)

type tracksCache struct {
	cache *[]source.TrackTags
	ctime int64
}

func (hts *HttpTrackServer) tracksCache() *[]source.TrackTags {
	if hts.index.Timestamp() > hts.tc.ctime || hts.index.Timestamp() == 0 {
		var err error
		hts.tc.cache, err = hts.index.GetAllTracks()
		if err != nil {
			log.Println("ERROR:", err)
		}
		hts.tc.ctime = hts.index.Timestamp()
	}

	return hts.tc.cache
}

type pager struct {
	Label  string
	Path   string
	Active bool
}

type tracks struct {
	Tracks []source.TrackTags
	Pager  []pager
}

// Quick and dirty handler to serve all tracks in the database. Works just for
// files.
func (hts *HttpTrackServer) handlerAllTracks(w http.ResponseWriter, r *http.Request) {
	// Only show that many tracks on one page
	const shownTracks = 100

	l := hts.tracksCache()

	var pagestring string
	var pagenum int

	// parse URL for pagenumber and validate
	if _, err := fmt.Sscanf(r.RequestURI, "/%s", &pagestring); err != nil {
		pagenum = 0
	} else {
		pagenum, err = strconv.Atoi(pagestring)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}
	if pagenum < 0 || pagenum > len(*l)/shownTracks {
		http.NotFound(w, r)
		return
	}

	if pagenum < 0 {
		pagenum = 0
	} else if pagenum > len(*l)/shownTracks+1 {
		pagenum = len(*l) / shownTracks
	}

	// slicing the right tracks
	min, max := pagenum*100, pagenum*100+shownTracks-1
	if max >= len(*l) {
		max = len(*l) - 1
	}

	// populating data
	t := &tracks{
		Tracks: (*l)[min:max],
		Pager:  make([]pager, len(*l)/shownTracks+1),
	}
	for i := 0; i < len(t.Pager); i++ {
		if i == pagenum {
			t.Pager[i].Active = true
		}
		t.Pager[i].Label = strconv.Itoa(i)
		t.Pager[i].Path = strconv.Itoa(i)
	}

	// FIXME: Error handling
	// render alltracks to string to nest it into index
	body, _ := renderToString(templates.Lookup("alltracks.html"), t)
	p := &page{
		Title: "musicrawler",
		Body:  template.HTML(body),
	}

	renderTemplate(w, "index", p)
}
