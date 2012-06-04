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
	//"net/http"
	//_ "net/http/pprof"
	//"time"
	"log"
	"os"
	"runtime/pprof"
)

const databasefn = "index.db"

var supportedFileTypes []string = []string{"mp3", "ogg"}

///////
//type testCrawler string
//
//func (t *testCrawler) Crawl(tracks chan<- TrackInfo, done chan<- bool) {
//	for i := int64(0); i < 30000; i++ {
//		tracks <- &FileInfo{filename: "/home/mokasin/Music/test.mp3", mtime: i}
//	}
//	done <- true
//}
//
///////

func updateFiles(dir string, index *Index) {
	var added, updated int
	var status *UpdateStatus
	var result *UpdateResult

	trackInfoChannel := make(chan TrackInfo, 20)
	statusChannel := make(chan *UpdateStatus, 10)
	resultChannel := make(chan *UpdateResult)
	doneChannel := make(chan bool)

	filecrawler := NewFileCrawler(dir, supportedFileTypes)

	// Plug output of CrawlFiles into index.Update over fileInfoChannel
	go index.Update(trackInfoChannel, statusChannel, resultChannel)
	go filecrawler.Crawl(trackInfoChannel, doneChannel)

	//tt := new(testCrawler)
	//go tt.Crawl(trackInfoChannel, doneChannel)

	counter := 0
TRACKUPDATE:
	for {
		select {
		case status = <-statusChannel:
			counter++
			if status.err != nil {
				fmt.Printf("%d: %d, INDEX ERROR (%s): %v\n", counter,
					status.action, status.path, status.err)
			} else {
				fmt.Printf("%d: %d, %s\n", counter, status.action, status.path)
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
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

	dir := "."
	flag.Parse()

	//PROFILER START

	//go http.ListenAndServe(":12345", nil)

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	//PROFILER END

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

	// endless loop, so pprof server isn't killed
	//for {
	//	time.Sleep(5 * time.Second)
	//}

}
