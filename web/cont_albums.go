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

package web

import (
	"fmt"
	"musicrawler/lib/database"
	"musicrawler/lib/database/query"
	"musicrawler/lib/web/controller"
	"musicrawler/model"
	"net/http"
	"path/filepath"
	"strconv"
)

type trackLink struct {
	Track model.Track
	Path  string
}

type albumsSelectTmpl struct {
	Tracks []trackLink
}

// Controller to serve artists
type ControllerAlbums struct {
	controller.Controller
}

// Constructor.
func NewControllerAlbums(db *database.Database, route, filepath string) *ControllerAlbums {
	c := &ControllerAlbums{
		controller.Controller: *controller.NewController(db, route, filepath),
	}

	c.AddTemplate("select", "index", "album")

	return c
}

// Implementation of SelectHandler.
func (self *ControllerAlbums) Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Not implemented yet.")
}

func (self *ControllerAlbums) Select(w http.ResponseWriter, r *http.Request, selector string) {
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

	var album model.Album

	err = query.New(self.Db, "album").Find(id).Exec(&album)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tracks []model.Track

	q := album.TracksQuery(self.Db)
	q.Join("album", "id", "", "album_id")
	q.Join("artist", "id", "album", "artist_id")

	err = q.Order("tracknumber").Exec(&tracks)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var td albumsSelectTmpl
	td.Tracks = make([]trackLink, len(tracks))

	// prepare structure for template
	for i := 0; i < len(tracks); i++ {
		td.Tracks[i].Track = tracks[i]
		td.Tracks[i].Path = fmt.Sprintf("/%s/%d/%s",
			"content", tracks[i].Id, filepath.Base(tracks[i].Path))
	}

	// render the website
	self.RenderPage(
		w,
		"select",
		&controller.Page{Title: album.Name},
		td,
	)
}
