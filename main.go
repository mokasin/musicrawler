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

	trackInfoChannel := make(chan TrackInfo)
	statusChannel := make(chan *UpdateStatus)
	resultChannel := make(chan *UpdateResult)
	doneChannel := make(chan bool)

	timeStart := time.Now()

	// Output of crawler(s) connects to the input of index.Update() over
	// trackInfoChannel channel

	go func() {
		// signal is emitted, when index.Update() has cleaned up everything
		resultChannel <- index.Update(trackInfoChannel, statusChannel)
	}()

	//filecrawler := NewFileCrawler(dir, supportedFileTypes)
	//go filecrawler.Crawl(trackInfoChannel, doneChannel)

	tt := new(testCrawler)
	go tt.Crawl(trackInfoChannel, doneChannel)

	go func() {
		<-doneChannel
		close(trackInfoChannel)
	}()

	counter := 0
	for status := range statusChannel {
		counter++
		if status.err != nil {
			fmt.Printf("%d: %d, INDEX ERROR (%s): %v\n", counter,
				status.action, status.path, status.err)
		} else {
			if *verbosity {
				fmt.Printf("%6d: %d, %s\n", counter, status.action, status.path)
			}
			switch status.action {
			case TRACK_UPDATE:
				updated++
			case TRACK_ADD:
				added++
			}
		}
	}

	deltaTime := time.Since(timeStart).Seconds()

	result := <-resultChannel
	if result.err != nil {
		fmt.Println("DATABASE ERROR:", result.err)
	}

	fmt.Printf("Added: %d\tUpdated: %d\n", added, updated)
	fmt.Printf("Total: %.2f min. %.2f ms per track.\n", deltaTime/60,
		deltaTime/float64(added+updated)*1000)
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
	//defer index.Close()

	fmt.Println("-> Update files.")
	updateFiles(dir, index)
}
