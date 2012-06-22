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
	"musicrawler/filecrawler"
	"musicrawler/index"
	"time"
)

var supportedFileTypes []string = []string{"mp3", "ogg"}

func updateTracks() {
	var added, updated int

	statusChannel := make(chan *index.UpdateStatus, 100)
	resultChannel := make(chan error)

	timeStart := time.Now()

	go sourceList.Update(statusChannel, resultChannel)

	counter := 0
	for status := range statusChannel {
		counter++
		if status.Err != nil {
			fmt.Printf("%d: %d, INDEX ERROR (%s): %v\n", counter,
				status.Action, status.Path, status.Err)
		} else {
			if *verbosity {
				fmt.Printf("%6d: %d, %s\n", counter, status.Action, status.Path)
			}
			switch status.Action {
			case index.TRACK_UPDATE:
				updated++
			case index.TRACK_ADD:
				added++
			}
		}
	}

	r := <-resultChannel
	if r != nil {
		fmt.Printf("ERROR: %v", r)
	}
	deltaTime := time.Since(timeStart).Seconds()

	fmt.Printf("Added: %d\tUpdated: %d\n", added, updated)
	fmt.Printf("Total: %.4f min. %.2f ms per track.\n", deltaTime/60,
		deltaTime/float64(added+updated)*1000)
}

var verbosity = flag.Bool("v", false, "be verbose")
var sourceList *SourceList

func main() {
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

	fmt.Println("-> Open database:", dbFileName)

	// open or create database
	index, err := index.NewIndex(dbFileName)
	if err != nil {
		fmt.Println("DATABASE ERROR:", err)
		return
	}
	defer index.Close()

	sourceList = NewSourceList(index)

	if flag.NArg() == 0 {
		sourceList.Add(filecrawler.New(".", supportedFileTypes))
		fmt.Println("-> Crawling directory: ./")
	} else {
		for i := 0; i < flag.NArg(); i++ {
			sourceList.Add(filecrawler.New(flag.Arg(i), supportedFileTypes))
			fmt.Println("-> Crawling directory:", flag.Arg(i))
		}
	}

	fmt.Println("-> Update files.")
	updateTracks()

	fmt.Println("-> Starting webserver...\n")

	httptrackserver := NewHttpTrackServer(index)

	if err := httptrackserver.StartListing(); err != nil {
		fmt.Println("ERROR:", err)
	}
}
