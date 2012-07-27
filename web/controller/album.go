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
	"fmt"
	"musicrawler/lib/database"
	"musicrawler/lib/database/query"
	"musicrawler/lib/web/controller"
	"musicrawler/lib/web/router"
	"musicrawler/model/album"
	"musicrawler/model/track"
	"net/http"
	"path/filepath"
	"strconv"
)

// Controller to serve artists
type ControllerAlbum struct {
	controller.Controller
}

// Constructor.
func NewAlbum(db *database.Database, router *router.Router, filepath string) *ControllerAlbum {
	c := &ControllerAlbum{
		controller.Controller: *controller.NewController(db, router, filepath),
	}

	c.AddTemplate("index", "index", "albums")
	c.AddTemplate("show", "index", "album")

	return c
}

// Implementation of SelectHandler.
func (self *ControllerAlbum) Index(w http.ResponseWriter, r *http.Request) {
	var albums []album.Album

	err := query.New(self.Db, "album").Exec(&albums)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// prepare structure for template
	for i := 0; i < len(albums); i++ {
		albums[i].Link = fmt.Sprintf("%s/%d", self.Router.GetRouteOf("album"), albums[i].Id)
	}

	self.AddDataToTemplate("index", "Albums", &albums)

	// render the website
	self.RenderPage(
		w,
		"index",
		&controller.Page{Title: "Albums"},
	)
}

func (self *ControllerAlbum) Show(w http.ResponseWriter, r *http.Request, selector string) {
	id, err := strconv.Atoi(selector)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := self.Db.BeginTransaction(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer self.Db.EndTransaction()

	var album album.Album

	err = query.New(self.Db, "album").Find(id).Exec(&album)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tracks []track.Track

	q := album.TracksQuery(self.Db)
	q.Join("album", "id", "", "album_id")
	q.Join("artist", "id", "album", "artist_id")

	err = q.Order("tracknumber").Exec(&tracks)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// prepare structure for template
	for i := 0; i < len(tracks); i++ {
		tracks[i].Link = fmt.Sprintf(
			"%s/%d/%s",
			self.Router.GetRouteOf("content"),
			tracks[i].Id,
			filepath.Base(tracks[i].Path),
		)
	}

	self.AddDataToTemplate("show", "Album", &album)
	self.AddDataToTemplate("show", "Tracks", &tracks)

	// render the website
	self.RenderPage(
		w,
		"show",
		&controller.Page{Title: album.Name},
	)
}
