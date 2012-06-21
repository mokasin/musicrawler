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
	"musicrawler/index"
	"musicrawler/source"
)

// Struct to manage different track sources
type SourceList struct {
	sources *list.List
	index   *index.Index
}

// Constructor of Sources.
func NewSourceList(i *index.Index) *SourceList {
	return &SourceList{
		sources: list.New(),
		index:   i,
	}
}

// Add a source to source list.
func (s *SourceList) Add(source source.TrackSource) {
	s.sources.PushBack(source)
}

// Remove element from source list.
func (s *SourceList) Remove(e *list.Element) {
	s.sources.Remove(e)
}

func (s *SourceList) Update(statusChannel chan *index.UpdateStatus,
	result chan error) {

	trackInfoChannel := make(chan source.TrackInfo, 100)
	updateResultChannel := make(chan *index.UpdateResult)
	doneChannel := make(chan bool)

	// Output of crawler(s) connects to the input of index.Update() over
	// trackInfoChannel channel
	go s.index.Update(trackInfoChannel, statusChannel, updateResultChannel)

	running := 0

	for e := s.sources.Front(); e != nil; e = e.Next() {
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

	// wait for index.Update to finish
	r := <-updateResultChannel
	result <- r.Err
}
