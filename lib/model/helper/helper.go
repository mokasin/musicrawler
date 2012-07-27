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

package helper

import (
	"net/url"
	"strings"
)

type activelink struct {
	Label  string
	Link   string
	Active bool
}

type Pager []activelink

func NewPager(baseurl string, labels []string, active string) Pager {
	// creating pager
	pager := make(Pager, len(labels))

	for i := 0; i < len(labels); i++ {
		if labels[i] == active {
			pager[i].Active = true
		}
		pager[i].Label = labels[i]
		v := url.Values{}
		v.Add("page", labels[i])
		pager[i].Link = "/" + strings.TrimLeft(baseurl, "/") + "?" + v.Encode()
	}

	return pager
}
