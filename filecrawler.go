/*  Copyright 2012, mokasin
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package main

import "os"
import "path/filepath"

type FileInfo struct {
	Name  string
	Mtime int64
}

type walker struct {
	Dir       string
	Filetypes []string
	receiver  chan<- *FileInfo
}

// Sends path to w.receiver channel, if fileextension matches one of Filetypes.
func (w *walker) walkfunc(path string, info os.FileInfo, err error) error {
	for _, v := range w.Filetypes {
		if filepath.Ext(path) == "."+v {
			w.receiver <- &FileInfo{path, info.ModTime().Unix()}
			break
		}
	}

	return nil
}

// Sends all filepathes of type filetypes to the receiver channel. Is meant to
// be a goroutine.
func CrawlFiles(dir string, filetypes []string, receiver chan<- *FileInfo) {
	w := &walker{Dir: dir, Filetypes: filetypes, receiver: receiver}

	// have to use closure because argument as to be a function not a method
	filepath.Walk(dir,
		func(p string, i os.FileInfo, e error) error {
			return w.walkfunc(p, i, e)
		})

	close(receiver)
}
