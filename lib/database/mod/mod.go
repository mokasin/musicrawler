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

package mod

import (
	"database/sql"
	. "musicrawler/lib/database"
	"musicrawler/lib/database/encoding"
)

type Mod struct {
	db    *Database
	table string
}

func New(db *Database, table string) *Mod {
	return &Mod{db: db, table: table}
}

func (self *Mod) Insert(item interface{}) (sql.Result, error) {
	return self.insert(item, false)
}

func (self *Mod) InsertIgnore(item interface{}) (sql.Result, error) {
	return self.insert(item, true)
}

// Insert adds a new row the associated table from the given struct. If ignore
// is true it ignores duplication conflicts.
func (self *Mod) insert(item interface{}, ignore bool) (sql.Result, error) {
	entries, err := encoding.Encode(item)
	if err != nil {
		return nil, err
	}

	var sql, cols, qmarks string
	vals := make([]interface{}, len(entries))

	// prepare arguments
	for i := 0; i < len(entries); i++ {
		cols += entries[i].Column + ","
		vals[i] = entries[i].Value
		qmarks += "?,"
	}
	// remove last comma
	cols = cols[:len(cols)-1]
	qmarks = qmarks[:len(qmarks)-1]

	if ignore {
		sql = "INSERT OR IGNORE INTO " +
			self.table + "(" + cols + ") VALUES(" + qmarks + ")"
	} else {
		sql = "INSERT INTO " + self.table +
			"(" + cols + ") VALUES(" + qmarks + ")"
	}

	res, err := self.db.Execute(sql, vals...)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Update updates a entry with ID id in the associated table.
func (self *Mod) Update(id int, item interface{}) error {
	entries, err := encoding.Encode(item)
	if err != nil {
		return err
	}

	vals := make([]interface{}, len(entries))
	sql := "UPDATE " + self.table + " SET "

	// prepare arguments
	for i := 0; i < len(entries); i++ {
		sql += entries[i].Column + " = ?,"
		vals[i] = entries[i].Value
	}
	// remove last comma
	sql = sql[:len(sql)-1]

	sql += " WHERE ID = ?"

	vals = append(vals, id)

	_, err = self.db.Execute(sql, vals...)
	if err != nil {
		return err
	}

	return nil
}
