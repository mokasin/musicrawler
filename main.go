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
	"flag"
	"fmt"
)

const databasefn = "index.db"

var supportedFileTypes []string = []string{"mp3", "ogg"}

func updateFiles(dir string, index *Index) {
	var filecrawler TrackSource
	var added, updated int
	var status *UpdateStatus
	var result *UpdateResult

	trackInfoChannel := make(chan TrackInfo, 20)
	statusChannel := make(chan *UpdateStatus, 10)
	resultChannel := make(chan *UpdateResult)
	doneChannel := make(chan bool)

	filecrawler = NewFileCrawler(dir, supportedFileTypes)

	// Plug output of CrawlFiles into index.Update over fileInfoChannel
	go index.Update(trackInfoChannel, statusChannel, resultChannel)
	go filecrawler.Crawl(trackInfoChannel, doneChannel)

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
		case <-doneChannel:
			close(trackInfoChannel)
		case result = <-resultChannel:
			break TRACKUPDATE
		}
	}

	if result.err != nil {
		fmt.Println("DATABASE ERROR:", result.err)
	}

	fmt.Printf("Added: %d\tUpdated: %d\n", added, updated)

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
