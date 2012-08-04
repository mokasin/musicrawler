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

package controller

import (
	"fmt"
	"musicrawler/lib/web/env"
	"musicrawler/lib/web/tmpl"
	"net/url"
)

// Base controller type. A controller needs access to the enviroment (database,
// router, ...) and has to manage templates.
type Controller struct {
	Env  *env.Environment
	Tmpl *tmpl.Template
}

// Constructor. An environment must be provided so the controller has acces to
// database, router, ...
func NewController(env *env.Environment) *Controller {
	return &Controller{
		Env:  env,
		Tmpl: tmpl.New(env.TmplPath),
	}
}

// URL returns an url to the to a route. The first argument must be the name of
// the route. Then pairs of arguments can be given.
//
// See gorilla/mux URL-Method for further help.
func (self *Controller) URL(s ...string) (*url.URL, error) {

	r := self.Env.Router.Get(s[0])

	if r == nil {
		return nil, fmt.Errorf("No such route with name %s.", s[0])
	}

	url, err := r.URL(s[1:]...)

	if err != nil {
		return nil, err
	}

	return url, nil
}
