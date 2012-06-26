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

package source

import "fmt"

// Metadata for a track
type TrackTags struct {
	Path    string
	Title   string
	Artist  string
	Album   string
	Comment string
	Genre   string
	Year    uint
	Track   uint
	Bitrate uint
	Length  uint
}

func (tt *TrackTags) LengthString() string {
	return fmt.Sprintf("%d:%02d", tt.Length/60, tt.Length%60)
}

// Basic information about a track.
type TrackInfo interface {
	Path() string
	Mtime() int64
	Tags() (*TrackTags, error)
}

// Abstract interface for sources of tracks. To implement the interface a method
// Crawl has to be defined, that sends the tracks of the source over the tracks
// channel.
type TrackSource interface {
	Crawl(tracks chan<- TrackInfo, done chan<- bool)
}
