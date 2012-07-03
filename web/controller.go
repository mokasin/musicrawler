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
)

type Controller struct {
	index *index.Index
	tmpl  *template.Template
	route string
}

// Constructor. Needs templates to register.
func NewController(index *index.Index, route string, templates ...string) *Controller {
	var tmpl *template.Template

	if len(templates) > 0 {
		for i := 0; i < len(templates); i++ {
			templates[i] = websitePath + "templates/" + templates[i] + ".tpl"
		}
		tmpl = template.Must(template.ParseFiles(templates...))
	}

	return &Controller{
		index: index,
		route: route,
		tmpl:  tmpl,
	}
}

// Parses and returns template with name name. At the first call, the parsed
// template is saved at c.tmpl
func (c *Controller) Tmpl(name string) *template.Template {
	t := c.tmpl.Lookup(name + ".html")
	if t == nil {
		t, _ = template.ParseFiles(websitePath + "templates/" + name + ".tpl")
	}

	return t
}
