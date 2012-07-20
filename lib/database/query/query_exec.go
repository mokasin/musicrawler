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

package query

import (
	"fmt"
	. "musicrawler/lib/database/encoding"
)

// Exec queries database with query and writes results into dest. Dest must be a
// pointer to a slice of structs.
func (self *Query) Exec(dest interface{}) error {
	col, err := ExtractColumns(dest)
	if err != nil {
		return err
	}

	sql := self.columns(col...).toSQL()

	res, err := self.db.Query(sql.SQL, sql.Args...)
	if err != nil {
		return err
	}

	// writing result into structs given by the caller
	err = DecodeAll(res, dest)
	if err != nil {
		return err
	}

	return err
}

// Count returns the number of all database entries of this model.
func (self *Query) Count() (int, error) {
	sqlQuery := self.toSQL()

	sql := "SELECT COUNT(*) FROM (" + sqlQuery.SQL + ")"

	res, err := self.db.Query(sql, sqlQuery.Args...)
	if err != nil {
		return -1, err
	}

	if len(res) == 0 {
		return 0, nil
	}

	v, ok := res[0]["COUNT(*)"].(int)
	if !ok {
		return -1, fmt.Errorf("Result is no int.")
	}

	return v, nil
}

// Letters returns string of first letters in the column named column.
func (self *Query) Letters(column string) (string, error) {
	sqlQuery := self.toSQL()

	sql := "SELECT DISTINCT SUBSTR(UPPER(" + column + "),1,1) FROM (" +
		sqlQuery.SQL + ")"

	res, err := self.db.Query(sql, sqlQuery.Args...)
	if err != nil {
		return "", err
	}

	var s string
	for i := 0; i < len(res); i++ {
		v, ok := res[i][fmt.Sprintf("SUBSTR(UPPER(%s),1,1)", column)].(string)
		if !ok {
			return "", fmt.Errorf("Result is no string.")
		}

		s += v
	}

	return s, nil
}
