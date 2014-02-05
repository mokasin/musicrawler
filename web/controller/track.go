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
	//"code.google.com/p/gorilla/mux"
	//"musicrawler/lib/database/query"
	"musicrawler/lib/web/controller"
	"musicrawler/lib/web/env"
	"musicrawler/lib/web/tmpl"
	//"musicrawler/model/artist"
	"net/http"
)

// Controller to serve artists
type ControllerTrack struct {
	controller.Controller
}

// Constructor.
func NewTrack(env *env.Environment) *ControllerTrack {
	c := &ControllerTrack{
		controller.Controller: *controller.NewController(env),
	}

	c.Tmpl.AddTemplate("tracks", "index", "tracks")

	return c
}

func (self *ControllerTrack) Index(w http.ResponseWriter, r *http.Request) {
	// render the website
	self.Tmpl.RenderPage(
		w,
		"tracks",
		&tmpl.Page{},
	)
}
