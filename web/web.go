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
	"musicrawler/lib/database"
	"musicrawler/lib/web/env"
	"musicrawler/web/controller"
	"net"
	"net/http"
	"time"
)

// FIXME to ungeneric
const websitePath = "website/"
const assetsPath = websitePath + "assets/"

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
type Webserver struct {
	listener net.Listener
	addr     string
	env      *env.Environment

	cartist  *controller.ControllerArtist
	calbum   *controller.ControllerAlbum
	ccontent *controller.ControllerContent
	ctrack   *controller.ControllerTrack
}

// Constructor of Webserver. Needs an db.db to work on.
func New(db *database.Database, stat chan<- *Status, addr string) *Webserver {
	// set global variable
	statusChannel = stat

	env := env.New(db, websitePath)

	w := &Webserver{
		addr: addr,
		env:  env,

		cartist:  controller.NewArtist(env),
		calbum:   controller.NewAlbum(env),
		ccontent: controller.NewContent(env),
		ctrack:   controller.NewTrack(env),
	}

	w.establishRoutes()

	return w
}

// establishRoutes sets up routes of HTTP server.
func (self *Webserver) establishRoutes() {
	self.env.Router.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			self.ctrack.Index(w, r)
		}).Methods("GET")

	/* artist */

	self.env.Router.HandleFunc("/artist",
		func(w http.ResponseWriter, r *http.Request) {
			self.cartist.Index(w, r)
		}).Methods("GET").Name("artist_base")

	self.env.Router.HandleFunc("/artist/{id:[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			self.cartist.Show(w, r)
		}).Methods("GET").Name("artist")

	self.env.Router.HandleFunc("/artist.json",
		func(w http.ResponseWriter, r *http.Request) {
			self.cartist.IndexJSON(w, r)
		}).Methods("GET")

	self.env.Router.HandleFunc("/artist/{id:[0-9]+}.json",
		func(w http.ResponseWriter, r *http.Request) {
			self.cartist.ShowJSON(w, r)
		}).Methods("GET")

	self.env.Router.HandleFunc("/artist/{id:[0-9]+}/albums.json",
		func(w http.ResponseWriter, r *http.Request) {
			self.cartist.AlbumsJSON(w, r)
		}).Methods("GET")

	/* album */

	self.env.Router.HandleFunc("/album",
		func(w http.ResponseWriter, r *http.Request) {
			self.calbum.Index(w, r)
		}).Methods("GET").Name("album_base")

	self.env.Router.HandleFunc("/album/{id:[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			self.calbum.Show(w, r)
		}).Methods("GET").Name("album")

	self.env.Router.HandleFunc("/album.json",
		func(w http.ResponseWriter, r *http.Request) {
			self.calbum.IndexJSON(w, r)
		}).Methods("GET")

	self.env.Router.HandleFunc("/album/{id:[0-9]+}.json",
		func(w http.ResponseWriter, r *http.Request) {
			self.calbum.ShowJSON(w, r)
		}).Methods("GET")

	self.env.Router.HandleFunc("/album/{id:[0-9]+}/tracks.json",
		func(w http.ResponseWriter, r *http.Request) {
			self.calbum.TracksJSON(w, r)
		}).Methods("GET")

	/* content */

	self.env.Router.HandleFunc("/content/{id:[0-9]+}/{filename}",
		func(w http.ResponseWriter, r *http.Request) {
			self.ccontent.Show(w, r)
		}).Methods("GET").Name("content")

	/* assets */

	// Just serve the assets.
	http.Handle("/assets/",
		http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsPath))))

	// let the router handle the rest
	http.Handle("/", self.env.Router)
}

// Start starts http server that listens on self.addr.
func (self *Webserver) Start() {
	l, err := net.Listen("tcp", self.addr)
	if err != nil {
		msg("", err)
	}

	self.listener = l

	err = http.Serve(self.listener, nil)
	msg("", err)
}

// Stop stops current HTTP server.
func (self *Webserver) Stop() {
	self.listener.Close()
}
