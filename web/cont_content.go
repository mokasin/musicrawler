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
	"musicrawler/index"
	"net/http"
)

type ControllerContent struct {
	Controller
}

func NewControllerContent(db *index.Database, route string) *ControllerContent {
	return &ControllerContent{Controller: *NewController(db, route)}
}

// Serving a audio file that has an entry in the database.
func (self *ControllerContent) Select(w http.ResponseWriter, r *http.Request, path string) {
	var tracks []index.Track

	valid := false

	q := index.NewQuery(self.db, "track")
	err := q.Exec(&tracks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	for _, val := range tracks {
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

	msg(fmt.Sprintf("Serving \"%s\" to %s", path, r.RemoteAddr), nil)
}
