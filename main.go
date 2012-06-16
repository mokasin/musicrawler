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
	//"runtime/pprof"
	"time"
)

var supportedFileTypes []string = []string{"mp3", "ogg"}

func updateFiles(dir string, index *Index) {
	var added, updated int
	var status *UpdateStatus
	var result *UpdateResult

	trackInfoChannel := make(chan TrackInfo, 100)
	statusChannel := make(chan *UpdateStatus, 1000)
	resultChannel := make(chan *UpdateResult)
	doneChannel := make(chan bool)

	filecrawler := NewFileCrawler(dir, supportedFileTypes)

	// Plug output of CrawlFiles into index.Update over fileInfoChannel
	go index.Update(trackInfoChannel, statusChannel, resultChannel)
	go filecrawler.Crawl(trackInfoChannel, doneChannel)

	timeStart := time.Now()

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
				if *verbosity {
					fmt.Printf("%d: %d, %s\n", counter, status.action, status.path)
				}
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

	deltaTime := time.Since(timeStart).Seconds()

	fmt.Printf("Added: %d\tUpdated: %d\n", added, updated)
	fmt.Printf("Total: %.2fmin. %.2f sec per track\n", deltaTime/60,
		deltaTime/float64(added+updated))

}

var verbosity = flag.Bool("v", false, "be verbose")

func main() {
	var dir string = "."
	var dbFileName string
	//var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&dbFileName, "database", "index.db", "path to database")

	flag.Parse()

	////PROFILER START
	//if *cpuprofile != "" {
	//	f, err := os.Create(*cpuprofile)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	pprof.StartCPUProfile(f)
	//	defer pprof.StopCPUProfile()
	//}
	////PROFILER END

	if flag.NArg() != 0 {
		dir = flag.Arg(0)
	}

	// open or create database
	fmt.Println("-> Open database:", dbFileName)

	index, err := NewIndex(dbFileName)
	if err != nil {
		fmt.Println("DATABASE ERROR:", err)
		return
	}
	defer index.Close()

	fmt.Println("-> Update files.")
	updateFiles(dir, index)
}
