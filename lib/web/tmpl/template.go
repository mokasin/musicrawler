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

package tmpl

import (
	"html/template"
	"net/http"
)

// Basic page structure.
type Page struct {
	Title string
}

type Template struct {
	templates        map[string]*template.Template
	templateData     map[string]map[string]interface{}
	templateFilepath string
}

// Constructor. Needs templates to register.
func New(templateFilepath string) *Template {
	return &Template{
		templateFilepath: templateFilepath,
		templates:        make(map[string]*template.Template),
		templateData:     make(map[string]map[string]interface{}),
	}
}

// AddTemplate adds associated templates to the template cache.
func (self *Template) AddTemplate(name string, templates ...string) {
	if len(templates) > 0 {
		for i := 0; i < len(templates); i++ {
			templates[i] = self.templateFilepath + "templates/" + templates[i] + ".tpl"
		}
		self.templates[name] = template.Must(template.ParseFiles(templates...))
	}

	self.templateData[name] = make(map[string]interface{})
}

func (self *Template) AddDataToTemplate(template, data_id string, data interface{}) {
	self.templateData[template][data_id] = data
}

// Write template with name tmpl to w.
func (self *Template) RenderPage(w http.ResponseWriter, tmpl string, p *Page) {
	self.AddDataToTemplate(tmpl, "Page", p)

	err := self.templates[tmpl].Execute(w, self.templateData[tmpl])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
