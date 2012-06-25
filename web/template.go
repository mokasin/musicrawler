/*  Copyright 2012, mokasin
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  hts program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with hts program. If not, see <http://www.gnu.org/licenses/>.
 */

package web

import (
	"bytes"
	"html/template"
	"net/http"
)

// Caching the main template
var pageTemplates = template.Must(
	template.ParseFiles(websitePath + "templates/index.html"))

// Basic page structure.
type Page struct {
	Title string
	Body  template.HTML
}

func renderToString(tmpl *template.Template, data interface{}) (string, error) {
	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func renderInPage(w http.ResponseWriter, tmpl string, child *template.Template,
	childData interface{}, p *Page) {
	body, err := renderToString(child, childData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		p.Body = template.HTML(body)
	}

	renderPage(w, tmpl, p)
}

func renderPage(w http.ResponseWriter, tmpl string, p *Page) {
	err := pageTemplates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
