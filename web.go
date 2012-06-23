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
	"musicrawler/index"
	"musicrawler/source"
	"net/http"
)

// FIXME: needs an absolute path so musicrawler can be run anywhere
const web = "web/"
const assets = web + "assets"

var templates = template.Must(template.ParseFiles(web + "templates/index.html"))

type HttpTrackServer struct {
	index *index.Index
}

// Baisc page structure.
type page struct {
	Title string
	Body  template.HTML
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Quick and dirty handler to serve all tracks in the database. Works just for
// files.
func (hts *HttpTrackServer) handlerAllTracks(w http.ResponseWriter, r *http.Request) {
	p := &page{
		Title: "musicrawler",
	}

	l, err := hts.index.GetAllTracks()
	if err != nil {
		fmt.Println("ERROR:", err)
	}

	// Bad style. Don'mix look with code! But for now…
	body := "<table class=\"table table-condensed\">"
	body += "<thead><tr><th></th><th>Artist</th><th>Title</th><th>Album</th></thead>"

	// Doesn't work yet for mpeg due licensing problems.
	//const audio = "<audio controls=\"controls}\"><source src=\"%s\" type=\"audio/mpeg\" />Not supported.</audio> "
	const audio = "<a href=\"content%s\" title=\"Play\" class=\"sm2_button\">%s</a>"

	for e := l.Front(); e != nil; e = e.Next() {
		t, ok := e.Value.(source.TrackTags)
		if ok {
			body += fmt.Sprintf(
				"<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
				fmt.Sprintf(audio, t.Path, t.Title),
				t.Artist,
				"<a href=\"content"+t.Path+"\">"+t.Title+"</a>",
				t.Album,
			)
		}
	}

	body += "</table>"
	p.Body = template.HTML(body)

	renderTemplate(w, "index", p)
}

// Serving a (mp3)file.
func (hts *HttpTrackServer) handlerFileContent(w http.ResponseWriter, r *http.Request) {
	if *verbosity {
		log.Printf("Serving %s to %s", r.URL.Path[8:], r.RemoteAddr)
	}
	http.ServeFile(w, r, r.URL.Path[8:])
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
