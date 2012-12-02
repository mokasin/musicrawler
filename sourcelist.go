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
	"github.com/mokasin/musicrawler/lib/database"
	"github.com/mokasin/musicrawler/lib/source"
)

// Struct to manage different track sources
type SourceList struct {
	sources *list.List
	db      *database.Database
}

// Constructor of Sources.
func NewSourceList(db *database.Database) *SourceList {
	return &SourceList{
		sources: list.New(),
		db:      db,
	}
}

// Add a source to source list.
func (self *SourceList) Add(source source.TrackSource) {
	self.sources.PushBack(source)
}

// Remove element from source list.
func (self *SourceList) Remove(e *list.Element) {
	self.sources.Remove(e)
}

func (self *SourceList) Update(statusChannel chan *UpdateStatus,
	result chan *UpdateResult) {

	trackInfoChannel := make(chan source.TrackInfo, 100)
	updateResultChannel := make(chan *UpdateResult)
	doneChannel := make(chan bool)

	// Output of crawler(self) connects to the input of database.Update() over
	// trackInfoChannel channel
	go UpdateDatabase(self.db, trackInfoChannel, statusChannel, updateResultChannel)

	running := 0

	for e := self.sources.Front(); e != nil; e = e.Next() {
		if ts, ok := e.Value.(source.TrackSource); ok {
			running++
			go ts.Crawl(trackInfoChannel, doneChannel)
		}
	}

	// wait until every source crawling has finished
	go func() {
		for running > 0 {
			running--
			<-doneChannel
		}
		close(trackInfoChannel)
	}()

	// wait for database.Update to finish
	r := <-updateResultChannel
	result <- r
}
