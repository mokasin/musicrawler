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
	"container/list"
	"flag"
	"fmt"
)

const databasefn = "index.db"

var supportedFileTypes []string = []string{"mp3", "ogg"}

func getfilelist(directory string) (l *list.List) {
	recv := make(chan *FileInfo)

	go CrawlFiles(directory, supportedFileTypes, recv)

	l = list.New()

	for fi := range recv {
		l.PushBack(fi)
	}

	return l
}

func updateFiles(dir string, index *Index) {
	var added, updated int
	var status *UpdateStatus
	var result *UpdateResult

	statusChannel := make(chan *UpdateStatus)
	resultChannel := make(chan *UpdateResult)

	// get filelist
	filelist := getfilelist(dir)

	// Add all found files into Database
	go index.Update(filelist, statusChannel, resultChannel)

TRACKUPDATE:
	for {
		select {
		case status = <-statusChannel:
			if status.err != nil {
				fmt.Printf("INDEX ERROR (%s): %v\n", status.path, status.err)
			} else {
				switch status.action {
				case TRACK_UPDATE:
					updated++
				case TRACK_ADD:
					added++
				}
			}
		case result = <-resultChannel:
			break TRACKUPDATE
		}
	}

	if result.err != nil {
		fmt.Println("DATABASE ERROR:", result.err)
	}

	fmt.Printf("Added: %d\tUpdated: %d\tDeleted: %d\n", added, updated,
		result.deleted)

}

func main() {
	dir := "."
	flag.Parse()
	if flag.NArg() != 0 {
		dir = flag.Arg(0)
	}

	// open or create database
	fmt.Println("-> Open database:", databasefn)

	index, err := NewIndex(databasefn)
	if err != nil {
		fmt.Println("DATABASE ERROR:", err)
		return
	}
	defer index.Close()

	fmt.Println("-> Update files.")
	updateFiles(dir, index)
}
