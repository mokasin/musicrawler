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

package database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

var (
	ErrNoOpenTransaction   = errors.New("No open transaction.")
	ErrExistingTransaction = errors.New("There is an existing transaction.")
	ErrDatabaseExists      = errors.New("Can't create new database. A database already exists.")
)

type CreateTableFunc func(db *Database) error

// A Result is a mapping from column name to its value.
type Result map[string]interface{}

type Database struct {
	Filename  string
	db        *sql.DB
	timestamp int64

	tx       *sql.Tx // global transaction
	txOpen   bool    // flag true, when exists an open transaction
	fctables []CreateTableFunc

	newDB bool
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

	datab := &Database{Filename: filename, db: db, newDB: newdatabase}

	// Make it nosync and disable the journal
	if _, err := db.Exec("PRAGMA synchronous=OFF"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA journal_mode=OFF"); err != nil {
		return nil, err
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
func (self *Database) CreateDatabase() error {
	if !self.newDB {
		return ErrDatabaseExists
	}

	if len(self.fctables) == 0 {
		return fmt.Errorf("Can't create database. " +
			"No functions to create the tables are registered.")
	}

	if err := self.BeginTransaction(); err != nil {
		return err
	}
	defer self.EndTransaction()

	for _, t := range self.fctables {
		err := t(self)
		if err != nil {
			return err
		}
	}

	return nil
}

// Add registers a new function to create a table.
func (self *Database) Register(m CreateTableFunc) {
	self.fctables = append(self.fctables, m)
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

	// Updating timestamp
	self.timestamp = time.Now().Unix()

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
func (self *Database) Execute(sql string, args ...interface{}) (res sql.Result, err error) {
	if !self.txOpen {
		err = self.BeginTransaction()
		if err != nil {
			return nil, err
		}
		defer self.EndTransaction()
	}

	res, err = self.tx.Exec(sql, args...)
	return res, err
}

// QueryDB queries the database with a given SQL-string and arguments args and
// returns the result as a map from column name to its value.
func (self *Database) Query(sql string, args ...interface{}) ([]Result, error) {
	if !self.txOpen {
		err := self.BeginTransaction()
		if err != nil {
			return nil, err
		}
		defer self.EndTransaction()
	}

	// do the actual query
	rows, err := self.tx.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// find out about the columns in the database
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// prepare result
	var result []Result

	// stupid trick, because rows.Scan will not take []interface as args
	col_vals := make([]interface{}, len(columns))
	col_args := make([]interface{}, len(columns))

	// initialize col_args
	for i := 0; i < len(columns); i++ {
		col_args[i] = &col_vals[i]
	}

	// read out columns and save them in a Result map
	for rows.Next() {
		if err := rows.Scan(col_args...); err != nil {
			return nil, err
		}

		res := make(Result)

		for i := 0; i < len(columns); i++ {
			res[columns[i]] = col_vals[i]
		}

		result = append(result, res)
	}

	return result, rows.Err()
}
