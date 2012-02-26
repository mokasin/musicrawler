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
	recv := make(chan string)

	go CrawlFiles(directory, []string{"mp3", "ogg"}, recv)

	l = list.New()

	for path := range recv {
		l.PushBack(path)
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

	filelist := getfilelist(dir)

	db_exists := fileexist(databasefn)

	fmt.Println("-> Open database:", databasefn)
	// open or create database
	index, err := NewDatabase(databasefn)
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

	// Add all found files into Database
	if _, err := index.Update(filelist); err != nil {
		fmt.Println("DATABASE ERROR:", err)
	}
}
