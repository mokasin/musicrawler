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

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

type Index struct {
	Filename  string
	db        *sql.DB
	timestamp int64

	tx     *sql.Tx // global transaction
	txOpen bool    // flag true, when exists an open transaction

	Artists *Artists
	Albums  *Albums
	Tracks  *Tracks
}

// Creates a new Index struct and connects it to the database at filename.
// Needs to be closed with method Close()!
func NewIndex(filename string) (*Index, error) {
	_, err := os.Stat(filename)
	newdatabase := os.IsNotExist(err)

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	i := &Index{Filename: filename, db: db}

	// Make it nosync and disable the journal
	if _, err := i.db.Exec("PRAGMA synchronous=OFF"); err != nil {
		return nil, err
	}
	if _, err := i.db.Exec("PRAGMA journal_mode=OFF"); err != nil {
		return nil, err
	}

	// initializing members
	i.Artists = NewArtists(i)
	i.Albums = NewAlbums(i)
	i.Tracks = NewTracks(i)

	// If databsae file does not exist
	if newdatabase {
		if err := i.createDatabase(); err != nil {
			return nil, err
		}
	}

	return i, nil
}

func (i *Index) Timestamp() int64 {
	return i.timestamp
}

// Closes the opened database.
func (i *Index) Close() {
	i.db.Close()
}

// Creates the basic database structure.
func (i *Index) createDatabase() error {
	if err := i.BeginTransaction(); err != nil {
		return err
	}
	defer i.EndTransaction()

	if err := i.Artists.CreateDatabase(); err != nil {
		return err
	}

	if err := i.Albums.CreateDatabase(); err != nil {
		return err
	}

	if err := i.Tracks.CreateDatabase(); err != nil {
		return err
	}

	return nil
}

// BeginTransaction starts a new database transaction.
func (i *Index) BeginTransaction() (err error) {
	if i.txOpen {
		return &ErrExistingTransaction{}
	}

	i.tx, err = i.db.Begin()
	if err != nil {
		return err
	}

	i.txOpen = true

	return nil
}

// EndTransaction closes a opened database transaction.
func (i *Index) EndTransaction() error {
	if !i.txOpen {
		return &ErrNoOpenTransaction{}
	}

	i.txOpen = false

	return i.tx.Commit()
}

type ErrNoOpenTransaction struct{}

func (e *ErrNoOpenTransaction) Error() string { return "No open transaction." }

type ErrExistingTransaction struct{}

func (e *ErrExistingTransaction) Error() string {
	return "There is an existing transaction."
}
