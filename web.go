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
	"html/template"
	"log"
	"musicrawler/index"
	"net/http"
)

// FIXME: needs an absolute path so musicrawler can be run anywhere
const web = "web/"
const assets = web + "assets"

// Caching the templates
var templates = template.Must(template.ParseFiles(web+"templates/index.html",
	web+"templates/alltracks.html"))

// Basic page structure.
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

type HttpTrackServer struct {
	index *index.Index
	tc    tracksCache
}

// Constructor of HttpTrackServer. Needs an index.Index to work on.
func NewHttpTrackServer(i *index.Index) *HttpTrackServer {
	return &HttpTrackServer{index: i}
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
