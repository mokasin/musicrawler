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
	"html/template"
	"musicrawler/index"
	"net/http"
)

// Basic page structure.
type Page struct {
	Title string
}

type Controller struct {
	db        *index.Database
	templates map[string]*template.Template
	route     string
}

// Constructor. Needs templates to register.
func NewController(db *index.Database, route string) *Controller {
	return &Controller{
		db:        db,
		route:     route,
		templates: make(map[string]*template.Template),
	}
}

// AddTemplate adds associated templates to the template cache.
func (self *Controller) AddTemplate(name string, templates ...string) {
	if len(templates) > 0 {
		for i := 0; i < len(templates); i++ {
			templates[i] = websitePath + "templates/" + templates[i] + ".tpl"
		}
		self.templates[name] = template.Must(template.ParseFiles(templates...))
	}
}

// Write template with name tmpl to w.
func (self *Controller) renderPage(w http.ResponseWriter, tmpl string, p *Page, data interface{}) {
	m := map[string]interface{}{"page": p, "content": data}

	err := self.templates[tmpl].Execute(w, m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
