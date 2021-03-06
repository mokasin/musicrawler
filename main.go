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
	"github.com/mokasin/musicrawler/lib/database"
	"github.com/mokasin/musicrawler/lib/source/filecrawler"
	"github.com/mokasin/musicrawler/model/album"
	"github.com/mokasin/musicrawler/model/artist"
	"github.com/mokasin/musicrawler/model/track"
	"github.com/mokasin/musicrawler/web"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"
)

var supportedFileTypes []string = []string{"mp3", "ogg"}

func updateTracks() {
	var added, updated, errors int

	actionMsg := []string{"-", "M", "A"}

	statusChannel := make(chan *UpdateStatus, 100)
	resultChannel := make(chan *UpdateResult)

	timeStart := time.Now()

	go sourceList.Update(statusChannel, resultChannel)

	counter := 0
	for status := range statusChannel {
		counter++
		if status.Err != nil {
			if *vverbosity {
				fmt.Printf("%6d: %s, INDEX ERROR (%s): %v\n", counter,
					actionMsg[status.Action], status.Path, status.Err)
			}
			errors++
		} else {
			if *vverbosity {
				fmt.Printf("%6d: %s, %s\n", counter,
					actionMsg[status.Action], status.Path)
			}

			switch status.Action {
			case TRACK_UPDATE:
				updated++
			case TRACK_ADD:
				added++
			}
		}
	}

	r := <-resultChannel
	if r.Err != nil {
		fmt.Printf("ERROR: %v", r)
	}
	deltaTime := time.Since(timeStart).Seconds()

	fmt.Printf("   Added: %d\tUpdated: %d\tDeleted: %d\tErrors: %d\n",
		added, updated, r.Deleted, errors)
	fmt.Printf("   Total: %.4f min. %.2f ms per track.\n", deltaTime/60,
		deltaTime/float64(added+updated)*1000)
}

var version string
var verbosity = flag.Bool("v", false, "be verbose")
var vverbosity = flag.Bool("vv", false, "be very verbose")
var sourceList *SourceList

func main() {
	var dbFileName string
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&dbFileName, "database", "index.db", "path to database")
	updateFlag := flag.Bool("u", true, "update database")
	flag.Parse()

	//PROFILER START
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	//PROFILER END

	fmt.Printf("musicrawler v. %s\n", version)
	fmt.Println("-> Open database:", dbFileName)

	// open or create database
	mydb, err := database.NewDatabase(dbFileName)
	if err != nil {
		fmt.Println("DATABASE ERROR:", err)
		return
	}
	defer mydb.Close()

	// Create database tables
	mydb.Register(artist.CreateArtistTable)
	mydb.Register(album.CreateAlbumTable)
	mydb.Register(track.CreateTrackTable)

	err = mydb.CreateDatabase()
	if err != nil && err != database.ErrDatabaseExists {
		fmt.Println(err)
	}

	sourceList = NewSourceList(mydb)

	if *updateFlag {
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
	}

	fmt.Println("-> Starting webserver...\n")

	status := make(chan *web.Status, 1000)

	w := web.New(mydb, status, ":8080")
	go w.Start()

	fmt.Println("   ...Listening on :8080")

	// React on SIGINT
	c := make(chan os.Signal, 1)
	stopc := make(chan bool)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			switch sig {
			case syscall.SIGINT:
				fmt.Println("Stopping server.")
				w.Stop()
				stopc <- true
			}
		}
	}()

	for {
		select {
		case msg := <-status:
			if *verbosity {
				if msg.Err != nil {
					fmt.Printf("%v: SERVER ERROR: %v\n", msg.Timestamp, msg.Err)
					break
				} else {
					fmt.Printf("%v: %s\n", msg.Timestamp, msg.Msg)
				}
			}
		case <-stopc:
			return
		}
	}
}
