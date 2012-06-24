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
	"bytes"
	"fmt"
	"html/template"
	"log"
	"musicrawler/index"
	"musicrawler/source"
	"net/http"
	"strconv"
)

// FIXME: needs an absolute path so musicrawler can be run anywhere
const web = "web/"
const assets = web + "assets"

var templates = template.Must(template.ParseFiles(web+"templates/index.html",
	web+"templates/alltracks.html"))

type tracksCache struct {
	cache *[]source.TrackTags
	ctime int64
}

type HttpTrackServer struct {
	index *index.Index
	tc    tracksCache
}

// Baisc page structure.
type page struct {
	Title string
	Body  template.HTML
}

func renderToString(template *template.Template, data interface{}) (string, error) {
	var buffer bytes.Buffer
	if err := template.Execute(&buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	p := &page{
		Title: "musicrawler",
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

	//var prevnext string
	//if pagenum == 0 {
	//	prevnext = "Previous | <a href=\"" + strconv.Itoa(pagenum+1) + "\" >Next </a>"
	//} else if pagenum == len(*l)/shownTracks {
	//	prevnext = "<a href=\"" + strconv.Itoa(pagenum-1) + "\" > Previous</a> | Next"
	//} else {
	//	prevnext = "<a href=\"" + strconv.Itoa(pagenum-1) + "\" > Previous</a> | <a href=\"" + strconv.Itoa(pagenum+1) + "\" >Next </a>"
	//}

	// render alltracks to string to nest it into index
	body, _ := renderToString(templates.Lookup("alltracks.html"), t)
	p.Body = template.HTML(body)

	renderTemplate(w, "index", p)
}

// Serving a (mp3)file.
func (hts *HttpTrackServer) handlerFileContent(w http.ResponseWriter, r *http.Request) {
	// validate path against database
	valid := false
	path := r.URL.Path[8:]
	for _, val := range *hts.tc.cache {
		if path == val.Path {
			valid = true
			break
		}
	}

	if !valid {
		http.NotFound(w, r)
		return
	}

	if *verbosity {
		log.Printf("Serving %s to %s", path, r.RemoteAddr)
	}
	http.ServeFile(w, r, path)
}

// Constructor of HttpTrackServer. Needs an index.Index to work on.
func NewHttpTrackServer(i *index.Index) *HttpTrackServer {
	return &HttpTrackServer{index: i}
}

// Starts http server on port 8080
func (hts *HttpTrackServer) StartListing() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hts.handlerAllTracks(w, r)
	})

	http.Handle("/assets/",
		http.StripPrefix("/assets/", http.FileServer(http.Dir(assets))))

	http.HandleFunc("/content/", func(w http.ResponseWriter, r *http.Request) {
		hts.handlerFileContent(w, r)
	})

	return http.ListenAndServe(":8080", nil)
}
