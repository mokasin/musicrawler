/*  Copyright 2012, mokasin
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with the program. If not, see <http://www.gnu.org/licenses/>.
 */

package controller

import (
	"code.google.com/p/gorilla/mux"
	"github.com/mokasin/musicrawler/lib/database/query"
	"github.com/mokasin/musicrawler/lib/web/controller"
	"github.com/mokasin/musicrawler/lib/web/env"
	"github.com/mokasin/musicrawler/lib/web/tmpl"
	"github.com/mokasin/musicrawler/model/album"
	"github.com/mokasin/musicrawler/model/track"
	"net/http"
	"path/filepath"
	"strconv"
)

// Controller to serve artists
type ControllerAlbum struct {
	controller.Controller
}

// Constructor.
func NewAlbum(env *env.Environment) *ControllerAlbum {
	c := &ControllerAlbum{
		controller.Controller: *controller.NewController(env),
	}

	c.Tmpl.AddTemplate("album_index", "index", "albums")
	c.Tmpl.AddTemplate("album_show", "index", "album")

	return c
}

// Implementation of SelectHandler.
func (self *ControllerAlbum) Index(w http.ResponseWriter, r *http.Request) {
	var albums []album.Album

	// retreive all albums
	err := query.New(self.Env.Db, "album").Exec(&albums)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// prepare data for template
	for i := 0; i < len(albums); i++ {
		url, err := self.URL("album", controller.Pairs{"id": albums[i].Id})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		albums[i].Link = url
	}

	self.Tmpl.AddDataToTemplate("album_index", "Albums", &albums)

	// render the website
	self.Tmpl.RenderPage(
		w,
		"album_index",
		&tmpl.Page{Title: "Albums"},
	)
}

func (self *ControllerAlbum) Show(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := self.Env.Db.BeginTransaction(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer self.Env.Db.EndTransaction()

	// retreive album by id
	var album album.Album

	err = query.New(self.Env.Db, "album").Find(id).Exec(&album)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// retreive tracks of album
	var tracks []track.Track

	q := album.TracksQuery(self.Env.Db)
	q.Join("album", "id", "", "album_id")
	q.Join("artist", "id", "album", "artist_id")

	err = q.Order("tracknumber").Exec(&tracks)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// prepare data for template
	for i := 0; i < len(tracks); i++ {
		url, err := self.URL("content", controller.Pairs{"id": tracks[i].Id, "filename": filepath.Base(tracks[i].Path)})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks[i].Link = url
	}

	self.Tmpl.AddDataToTemplate("album_show", "Album", &album)
	self.Tmpl.AddDataToTemplate("album_show", "Tracks", &tracks)

	backlink, _ := self.URL("artist", controller.Pairs{"id": album.ArtistID})

	// render the website
	self.Tmpl.RenderPage(
		w,
		"album_show",
		&tmpl.Page{Title: album.Name, BackLink: backlink},
	)
}
