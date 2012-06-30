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

import "strings"

// Model of a simple pager.
type pager struct {
	Label  string
	Path   string
	Active bool
}

func Breadcrump(path string) (r []pager) {
	tokens := ParseURL(path)

	r = make([]pager, len(tokens))
	for i := 0; i < len(tokens); i++ {
		r[i].Label = tokens[i]
		r[i].Path = "/" + strings.Join(tokens[:i+1], "/")
	}
	r[len(r)-1].Active = true

	return r
}
