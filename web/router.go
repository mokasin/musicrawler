/*  Copyright 2012, mokasin
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  c program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with c program. If not, see <http://www.gnu.org/licenses/>.
 */

package web

import (
	"net/http"
	"strings"
)

// Handler to list all resource.
type IndexHandler interface {
	Index(http.ResponseWriter, *http.Request)
}

// Handler to list resources that get selected by selector.
type SelectHandler interface {
	Select(http.ResponseWriter, *http.Request, string)
}

// ParseURL splits in subparts seperated by "/".
func ParseURL(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}

type Router struct {
	routes map[string]interface{}
}

func NewRouter() *Router {
	return &Router{routes: make(map[string]interface{})}
}

func NotImplemented(w http.ResponseWriter) {
	http.Error(w, "501 Not Implemented", http.StatusNotImplemented)
}

// routeHandler calls appropriate methods of controller depending on the path.
func (r *Router) routeHandler(w http.ResponseWriter, req *http.Request) {
	var resource, selector string

	path := strings.Trim(req.URL.Path, "/")
	// extract resource part and selector
	pos := strings.Index(path, "/")

	if pos == -1 {
		resource = path
	} else {
		resource = path[:pos]
		selector = path[pos+1:]
	}

	route, ok := r.routes[resource]
	if !ok {
		http.NotFound(w, req)
		return
	}

	if req.Method != "GET" {
		NotImplemented(w)
	}

	// call right method
	if len(selector) == 0 {
		handler, ok := route.(IndexHandler)
		if !ok {
			NotImplemented(w)
		}

		handler.Index(w, req)
	} else {
		handler, ok := route.(SelectHandler)
		if !ok {
			NotImplemented(w)
		}

		handler.Select(w, req, selector)
	}
}

// AddRoute registers a new route from a resource specified in an path to a
// controller.
func (r *Router) AddRoute(resource string, controller interface{}) {
	r.routes[resource] = controller

	http.HandleFunc(
		"/"+resource+"/",
		func(w http.ResponseWriter, req *http.Request) {
			r.routeHandler(w, req)
		},
	)
}
