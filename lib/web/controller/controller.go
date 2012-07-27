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
	"html/template"
	"musicrawler/lib/database"
	"musicrawler/lib/web/router"
	"net/http"
)

// Basic page structure.
type Page struct {
	Title string
}

type Controller struct {
	Db     *database.Database
	route  string
	Router *router.Router

	templates    map[string]*template.Template
	templateData map[string]map[string]interface{}
	filepath     string
}

// Constructor. Needs templates to register.
func NewController(db *database.Database, router *router.Router, filepath string) *Controller {
	return &Controller{
		Db:           db,
		Router:       router,
		filepath:     filepath,
		templates:    make(map[string]*template.Template),
		templateData: make(map[string]map[string]interface{}),
	}
}

// AddTemplate adds associated templates to the template cache.
func (self *Controller) AddTemplate(name string, templates ...string) {
	if len(templates) > 0 {
		for i := 0; i < len(templates); i++ {
			templates[i] = self.filepath + "templates/" + templates[i] + ".tpl"
		}
		self.templates[name] = template.Must(template.ParseFiles(templates...))
	}

	self.templateData[name] = make(map[string]interface{})
}

func (self *Controller) AddDataToTemplate(template, data_id string, data interface{}) {
	self.templateData[template][data_id] = data
}

// Write template with name tmpl to w.
func (self *Controller) RenderPage(w http.ResponseWriter, tmpl string, p *Page) {
	self.AddDataToTemplate(tmpl, "Page", p)

	err := self.templates[tmpl].Execute(w, self.templateData[tmpl])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (self *Controller) Route() string {
	return self.route
}

func (self *Controller) SetRoute(route string) {
	self.route = route
}
