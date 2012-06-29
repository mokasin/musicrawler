/*  Copyright 2012, mokasin
 *
 *  This program is free software: you can redistribute it and/or modify
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

package web

import (
	"fmt"
	"musicrawler/index"
	"net/http"
	"time"
)

// FIXME: needs an absolute path so musicrawler can be run anywhere
const websitePath = "website/"
const websiteAssetsPath = websitePath + "assets/"

type Status struct {
	Msg       string
	Err       error
	Timestamp time.Time
}

type HttpTrackServer struct {
	index  *index.Index
	status chan<- *Status
}

// Constructor of HttpTrackServer. Needs an index.Index to work on.
func NewHttpTrackServer(i *index.Index, status chan<- *Status) *HttpTrackServer {
	return &HttpTrackServer{index: i, status: status}
}

// msg sends Status to status channel.
func (hts *HttpTrackServer) msg(msg string, err error) {
	hts.status <- &Status{
		Msg:       msg,
		Err:       err,
		Timestamp: time.Now(),
	}
}

// Serving a track file.
func (hts *HttpTrackServer) handlerFileContent(w http.ResponseWriter, r *http.Request) {
	// validate path against database
	valid := false
	path := r.URL.Path[8:]
	tracks, err := hts.index.Tracks.All()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	for _, val := range *tracks {
		if path == val.Path {
			valid = true
			break
		}
	}

	if !valid {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, path)

	hts.msg(fmt.Sprintf("Serving \"%s\" to %s", path, r.RemoteAddr), nil)
}

// Starts http server on port 8080 and set routes.
func (hts *HttpTrackServer) StartListing() {
	c_allTracks := NewControllerAllTracks(hts.index)

	// methods are no expression -> closure
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c_allTracks.Handler(w, r)
	})

	http.Handle("/assets/",
		http.StripPrefix("/assets/", http.FileServer(http.Dir(websiteAssetsPath))))

	http.HandleFunc("/content/", func(w http.ResponseWriter, r *http.Request) {
		hts.handlerFileContent(w, r)
	})

	err := http.ListenAndServe(":8080", nil)
	hts.msg("", err)
}
