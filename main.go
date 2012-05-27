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
	"os"
)

const databasefn = "./index.db"

func getfilelist(directory string) (l *list.List) {
	recv := make(chan *FileInfo)

	go CrawlFiles(directory, []string{"mp3", "ogg"}, recv)

	l = list.New()

	for fi := range recv {
		l.PushBack(fi)
	}

	return l
}

// Retruns true, if file filename exists.
func fileexist(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func main() {
	flag.Parse()

	dir := "."

	if flag.NArg() != 0 {
		dir = flag.Arg(0)
	}

	fmt.Println("-> Digg filesystem.")

	filelist := getfilelist(dir)

	db_exists := fileexist(databasefn)

	fmt.Println("-> Open database:", databasefn)

	// open or create database
	index, err := NewIndex(databasefn)
	if err != nil {
		fmt.Println("DATABASE ERROR:", err)
		return
	}
	defer index.Close()

	// if database file doesn't exist, create new databse scheme
	if !db_exists {
		fmt.Println("-> Create database structure.")
		err = index.CreateDatabase()

		if err != nil {
			fmt.Println("DATABASE ERROR:", err)
			return
		}
	}

	fmt.Println("-> Update files.")

	var statusMsg [3]string
	statusMsg[TRACK_NOUPDATE] = "NUP"
	statusMsg[TRACK_UPDATE] = "UPD"
	statusMsg[TRACK_ADD] = "ADD"

	var added, updated int
	var status *UpdateStatus
	var result *UpdateResult

	statusChannel := make(chan *UpdateStatus)
	resultChannel := make(chan *UpdateResult)

	// Add all found files into Database
	go index.Update(filelist, statusChannel, resultChannel)

TRACKUPDATE:
	for {
		select {
		case status = <-statusChannel:
			if status.err != nil {
				fmt.Println("DATABASE ERROR (", status.path, "):", status.err)
			} else {
				switch status.action {
				case TRACK_UPDATE:
					updated++
				case TRACK_ADD:
					added++
				}
				//fmt.Println(statusMsg[status.action], status.path)
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
