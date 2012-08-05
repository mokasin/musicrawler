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

// The tmpl package provides methods to manage Go template files and data that
// is used in the template. By convention template files have the file ending
// .tpl.
//
// To use it, simple add some templates to a named group. They are associated.
// Those templates can access data, that is added by the AddDataToTemplate
// method.
//
// Refer to text/template documentation for further help.
package tmpl

import (
	"fmt"
	"html/template"
	"net/http"
)

// Basic page structure.
type Page struct {
	Title    string
	BackLink string
}

// Template manages templates and makes them accessible through a name.
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

// AddTemplate adds associated templates to the named template, parses and saves
// them. The file ending .tpl must not supplied as it is added automatically.
func (self *Template) AddTemplate(name string, templates ...string) {
	if len(templates) > 0 {
		for i := 0; i < len(templates); i++ {
			templates[i] = self.templateFilepath + "templates/" + templates[i] + ".tpl"
		}
		self.templates[name] = template.Must(template.ParseFiles(templates...))
	}

	self.templateData[name] = make(map[string]interface{})
}

// AddDataToTemplate adds data to template which is accessible by data_id.
func (self *Template) AddDataToTemplate(template, data_id string, data interface{}) error {
	_, ok := self.templateData[template]
	if !ok {
		return fmt.Errorf("There is no template named '%s' registered.", template)
	}

	self.templateData[template][data_id] = data

	return nil
}

// Write template with name tmpl to w.
//
// A address to a Page struct should be supplied. In it some general
// information for the template engine is saved.
func (self *Template) RenderPage(w http.ResponseWriter, tmpl string, p *Page) {
	self.AddDataToTemplate(tmpl, "Page", p)

	err := self.templates[tmpl].Execute(w, self.templateData[tmpl])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
