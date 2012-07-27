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
	"musicrawler/lib/database"
	"musicrawler/lib/database/query"
	"musicrawler/lib/web/controller"
	"musicrawler/lib/web/router"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type ControllerContent struct {
	controller.Controller
}

type trackPathId struct {
	Id   int    `column:"ID"`
	Path string `column:"path"`
}

func NewContent(db *database.Database, router *router.Router, filepath string) *ControllerContent {
	return &ControllerContent{
		controller.Controller: *controller.NewController(db, router, filepath),
	}
}

// Serving a audio file that has an entry in the database.
func (self *ControllerContent) Show(w http.ResponseWriter, r *http.Request, selector string) {
	// Remove .mp3/.ogg
	base := filepath.Base(selector)
	if len(base) != 0 {
		selector = selector[:strings.LastIndex(selector, "/")]
	}

	id, err := strconv.Atoi(selector)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var track trackPathId

	err = query.New(self.Db, "track").Find(id).Exec(&track)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.ServeFile(w, r, track.Path)
}
