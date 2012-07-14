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
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

var (
	ErrNoOpenTransaction   = errors.New("No open transaction.")
	ErrExistingTransaction = errors.New("There is an existing transaction.")
)

type Database struct {
	Filename  string
	db        *sql.DB
	timestamp int64

	tx     *sql.Tx // global transaction
	txOpen bool    // flag true, when exists an open transaction
}

// Creates a new Database struct and connects it to the database at filename.
// Needs to be closed with method Close()!
func NewDatabase(filename string) (*Database, error) {
	_, err := os.Stat(filename)
	newdatabase := os.IsNotExist(err)

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	datab := &Database{Filename: filename, db: db}

	// Make it nosync and disable the journal
	if _, err := db.Exec("PRAGMA synchronous=OFF"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA journal_mode=OFF"); err != nil {
		return nil, err
	}

	// If databsae file does not exist
	if newdatabase {
		if err := datab.createDatabase(); err != nil {
			return nil, err
		}
	}

	return datab, nil
}

func (self *Database) Timestamp() int64 {
	return self.timestamp
}

// Closes the opened database.
func (self *Database) Close() {
	self.db.Close()
}

// Creates the basic database structure.
func (self *Database) createDatabase() error {
	if err := self.BeginTransaction(); err != nil {
		return err
	}
	defer self.EndTransaction()

	if err := CreateArtistTable(self); err != nil {
		return err
	}

	if err := CreateArtistTable(self); err != nil {
		return err
	}

	if err := CreateTrackTable(self); err != nil {
		return err
	}

	return nil
}

// BeginTransaction starts a new database transaction.
func (self *Database) BeginTransaction() (err error) {
	if self.txOpen {
		return ErrExistingTransaction
	}

	self.tx, err = self.db.Begin()
	if err != nil {
		return err
	}

	self.txOpen = true

	return nil
}

// EndTransaction closes a opened database transaction.
func (self *Database) EndTransaction() error {
	if !self.txOpen {
		return ErrNoOpenTransaction
	}

	self.txOpen = false

	return self.tx.Commit()
}

// Execute just executes sql query in global transaction.
func (self *Database) Execute(sql string, args ...interface{}) error {
	if !self.txOpen {
		err := self.BeginTransaction()
		if err != nil {
			return err
		}
		defer self.EndTransaction()
	}

	_, err := self.tx.Exec(sql, args...)
	return err
}
