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

import (
	"gotaglib"
	"os"
	"path/filepath"
)

type FileInfo struct {
	filename string
	mtime    int64
}

func (fi *FileInfo) Path() string {
	return fi.filename
}

func (fi *FileInfo) Mtime() int64 {
	return fi.mtime
}

func (fi *FileInfo) Tags() (*TrackTags, error) {
	tag, err := gotaglib.NewTaggedFile(fi.filename)
	if err != nil {
		return nil, err
	}

	return &TrackTags{
		Path:    tag.Filename,
		Title:   tag.Title,
		Artist:  tag.Artist,
		Album:   tag.Album,
		Comment: tag.Comment,
		Genre:   tag.Genre,
		Year:    tag.Year,
		Track:   tag.Track,
		Bitrate: tag.Bitrate,
		Length:  tag.Length,
	}, nil
}

type FileWalker struct {
	Dir       string
	Filetypes []string
}

func NewFileCrawler(dir string, filetypes []string) *FileWalker {
	return &FileWalker{Dir: dir, Filetypes: filetypes}
}

// Sends TrackInfo if filetype matches one of w.Filetypes
func (w *FileWalker) walkfunc(receiver chan<- TrackInfo, path string,
	info os.FileInfo, err error) error {
	for _, v := range w.Filetypes {
		if filepath.Ext(path) == "."+v {
			receiver <- &FileInfo{filename: path, mtime: info.ModTime().Unix()}
			break
		}
	}

	return nil
}

// Sends all filepathes of type filetypes to the receiver channel. Is meant to
// be a goroutine.
func (w *FileWalker) Crawl(tracks chan<- TrackInfo, done chan<- bool) {
	// have to use closure because argument as to be a function not a method
	filepath.Walk(w.Dir,
		func(p string, i os.FileInfo, e error) error {
			return w.walkfunc(tracks, p, i, e)
		})

	done <- true
}
