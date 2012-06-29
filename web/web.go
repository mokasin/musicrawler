/*  Copyright 2012, mokasin
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package web

import (
	"musicrawler/index"
	"net/http"
	"time"
)

// FIXME to ungeneric
const websitePath = "website/"
const websiteAssetsPath = websitePath + "assets/"

var statusChannel chan<- *Status

type Status struct {
	Msg       string
	Err       error
	Timestamp time.Time
}

// msg sends Status to status channel.
func msg(msg string, err error) {
	statusChannel <- &Status{
		Msg:       msg,
		Err:       err,
		Timestamp: time.Now(),
	}
}

// Manages a HTTP server to serve audio files saved in database. 
type HTTPTrackServer struct {
	index  *index.Index
	router *Router
}

// Constructor of HTTPTrackServer. Needs an index.Index to work on.
func NewHTTPTrackServer(i *index.Index, stat chan<- *Status) *HTTPTrackServer {
	// set global variable
	statusChannel = stat

	return &HTTPTrackServer{index: i, router: NewRouter()}
}

// Starts http server on port 8080 and set routes.
func (h *HTTPTrackServer) StartListing() {
	msg("Start listening.", nil)

	// Adding routes

	h.router.AddRoute("artists", NewControllerArtists(h.index))
	h.router.AddRoute("content", NewControllerContent(h.index))

	// Just serve the assets.
	http.Handle("/assets/",
		http.StripPrefix("/assets/", http.FileServer(http.Dir(websiteAssetsPath))))

	// let the router handle the rest
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		h.router.routeHandler(w, req)
	})

	// and start the server.
	err := http.ListenAndServe(":8080", nil)

	msg("", err)
}
