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

package index

import "fmt"

// Artists represents the model of tracks in database.
type Artists struct {
	index            *Index
	lettersCache     cache
	allCache         cache
	firstLetterCache map[rune]cache
}

// Constructor returns intstance of Artists.
func NewArtists(i *Index) *Artists {
	return &Artists{
		index:            i,
		firstLetterCache: make(map[rune]cache),
	}
}

// Returns a string array for an arbitrary query (if it begins with "SELECT %s..."i)
func (a *Artists) QueryName(query string, args ...interface{}) (*[]string, error) {
	tx, err := a.index.db.Begin()
	if err != nil {
		return nil, err
	}

	var count int
	err = tx.QueryRow(fmt.Sprintf(query, "COUNT(*)"), args...).Scan(&count)
	if err != nil {
		return nil, err
	}

	artists := make([]string, count)

	rows, err := tx.Query(fmt.Sprintf(query, "name"), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	var artist string

	c := 0
	for rows.Next() {
		if err = rows.Scan(&artist); err != nil {
			return nil, err
		}
		artists[c] = artist
		c++
	}
	return &artists, rows.Err()
}

const artists_sql_all = "SELECT %s FROM Artist;"

// All returns a point to an array of all artist names
func (a *Artists) All() (*[]string, error) {
	if a.allCache.data == nil || a.allCache.ctime != a.index.Timestamp() {
		var err error
		a.allCache.data, err = a.QueryName(artists_sql_all)
		if err != nil {
			return nil, err
		}
		a.allCache.ctime = a.index.Timestamp()
	}

	val, ok := a.allCache.data.(*[]string)
	if !ok {
		return a.QueryName(artists_sql_all)
	}

	return val, nil
}

const artists_sql_startingwith = "SELECT %s FROM Artist WHERE name LIKE ? || '%%' ORDER BY UPPER(name);"

// Returns an array of all artist names
func (a *Artists) ByName(name string) (*[]string, error) {
	return a.QueryName(artists_sql_startingwith, name)
}

// ByFirstLetter returns a pointer to an array containg all artists which names
// starting with letter.
func (a *Artists) ByFirstLetter(letter rune) (*[]string, error) {
	if a.firstLetterCache[letter].data == nil ||
		a.firstLetterCache[letter].ctime != a.index.Timestamp() {

		data, err := a.ByName(string(letter))
		if err != nil {
			return nil, err
		}

		a.firstLetterCache[letter] = cache{
			data:  data,
			ctime: a.index.Timestamp(),
		}
	}

	val, ok := a.firstLetterCache[letter].data.(*[]string)
	if !ok {
		return a.ByName(string(letter))
	}

	return val, nil
}

const artists_sql_firstletters = "SELECT DISTINCT %s FROM Artist ORDER BY UPPER(name)"

// queryLetters reads leading letters of artists name from database
func (a *Artists) queryLetters() (letters string, err error) {
	rows, err := a.index.db.Query(
		fmt.Sprintf(artists_sql_firstletters, "UPPER(SUBSTR(name, 1, 1))"))
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var l string
	c := 0
	for rows.Next() {
		if err = rows.Scan(&l); err != nil {
			return "", err
		}
		letters += l
		c++
	}

	return letters, rows.Err()
}

// Letters returns a string of the first letters of the artist names.
func (a *Artists) Letters() (string, error) {
	var err error
	if a.lettersCache.data == "" || a.lettersCache.ctime != a.index.Timestamp() {
		a.lettersCache.data, err = a.queryLetters()
		a.lettersCache.ctime = a.index.Timestamp()
	}

	val, ok := a.lettersCache.data.(string)
	if !ok {
		return a.queryLetters()
	}

	return val, err
}

// Returns a map of all artists for each starting letter
func (a *Artists) FirstLetterMap() (map[string][]string, error) {

	var artists *[]string
	var err error
	m := make(map[string][]string)

	letters, err := a.Letters()
	if err != nil {
		return nil, err
	}

	for c := 0; c < len(letters); c++ {
		artists, err = a.ByFirstLetter(rune(letters[c]))
		if err != nil {
			return nil, err
		}
		m[string(letters[c])] = *artists
	}

	return m, nil
}
