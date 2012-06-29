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
	index  *index.Index
	tmpl   *template.Template
	status chan<- *Status
}

// Constructor.
func NewController(index *index.Index) *Controller {
	return &Controller{index: index}
}

// Parses and returns template with name name. At the first call, the parsed
// template is saved at c.tmpl
func (c *Controller) Tmpl(name string) *template.Template {
	if c.tmpl == nil {
		c.tmpl = template.Must(
			template.ParseFiles(websitePath + "templates/" + name + ".html"))
	}
	return c.tmpl.Lookup(name + ".html")
}
