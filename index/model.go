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
	"errors"
	"fmt"
	"reflect"
	"time"
)

var ErrWrongType error = errors.New("Wrong type.")

type Result map[string]interface{}

// Model describes the basis of all models, i.e. database representations.
// TODO Documentation
type Model struct {
	db   *Database
	name string // name of table

	timer    time.Time
	Duration time.Duration
}

// Constructor of Model. Needs a database db to work on it, a name of the
// model name and a fscan function, to read out the data from the database.
func NewModel(db *Database, name string) *Model {
	return &Model{
		db:   db,
		name: name,
	}
}

// Name returns the name of the model.
func (self *Model) Name() string {
	return self.name
}

/*
 *
 *	ENCODING AND DECODING
 *	MAP <-> STRUCT
 *
 */

// Encode eats a pointer to a struct src and converts all exported fields into a
// map
// 		"field name" => <values>
func (self *Model) Encode(src interface{}) (Result, error) {
	v := reflect.ValueOf(src)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("src must be a pointer to struct.")
	}

	v = v.Elem()
	t := v.Type()

	res := make(map[string]interface{})

	for i := 0; i < v.NumField(); i++ {
		// is the field exported?
		if t.Field(i).PkgPath != "" {
			continue
		}

		// check struct's tag if value should be set (!= "0")
		// and just use fields with tag
		if t.Field(i).Tag.Get("set") != "0" &&
			t.Field(i).Tag.Get("column") != "" {
			res[t.Field(i).Tag.Get("column")] = v.Field(i).Interface()
		}
	}

	return res, nil
}

// Decode reads a map of type Result and a structure like
// 		"field name" => <value>
// and spits out a struct to dest.
func (self *Model) Decode(src Result, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to struct.")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		// is the field exported?
		if t.Field(i).PkgPath != "" {
			continue
		}

		if t.Field(i).Tag.Get("column") == "" {
			continue
		}

		// read out struct's tag to get the column name
		val, ok := src[t.Field(i).Tag.Get("column")]
		if !ok {
			return fmt.Errorf("No column named '%s' connected to "+
				"'%s.%s %v' found in database.",
				t.Field(i).Tag.Get("name"),
				t.Name(), t.Field(i).Name, v.Field(i).Kind(),
			)
		}

		// do type assertion
		switch v.Field(i).Kind() {
		case reflect.Int:
			val, ok := val.(int64)
			if !ok {
				return fmt.Errorf("Cannot do assertion to 'int' on field "+
					"'%s.%s %v'.",
					t.Name(), t.Field(i).Name, v.Field(i).Kind())
			}

			v.Field(i).SetInt(int64(val))

		case reflect.String:
			val, ok := val.(string)
			if !ok {
				return fmt.Errorf("Cannot do assertion to 'string' on "+
					"'%s.%s %v').",
					t.Name(), t.Field(i).Name, v.Field(i).Kind())
			}

			v.Field(i).SetString(val)
		default:
			return fmt.Errorf("Type '%v' of '%s.%s' is not supported "+
				"right now.",
				v.Field(i).Kind(), t.Name(), t.Field(i).Name)
		}

	}

	return nil
}

// DecodeAll does what Decode does but with a couple of results.
func (self *Model) DecodeAll(src []Result, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || (v.Elem().Kind() != reflect.Slice &&
		v.Elem().Kind() != reflect.Struct) {
		return fmt.Errorf("dest must be a pointer to a slice or a struct.")
	}

	if v.Elem().Kind() == reflect.Struct {
		if len(src) != 1 {
			return fmt.Errorf("Can't write data from database to struct. " +
				"Dimension mismatch.")
		}
		return self.Decode(src[0], v.Interface())
	}

	// Making slice big enough
	t := reflect.TypeOf(dest)
	v.Elem().Set(reflect.MakeSlice(t.Elem(), len(src), len(src)))

	// Feed Decode method with it
	for i := 0; i < v.Elem().Len(); i++ {
		err := self.Decode(src[i], v.Elem().Index(i).Addr().Interface())
		if err != nil {
			return err
		}
	}

	return nil
}

/*
 *
 *	BASIC DATABASE ACCESS
 *
 */

// QueryDB TODO: Documentation needed.
func (self *Model) QueryDB(sql string, args ...interface{}) ([]Result, error) {
	if !self.db.txOpen {
		err := self.db.BeginTransaction()
		if err != nil {
			return nil, err
		}
		defer self.db.EndTransaction()
	}

	// do the actual query
	rows, err := self.db.tx.Query(sql, args...)
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

// Exec queries database with query and writes results into dest. Dest must be a
// pointer to a slice of structs.
func (self *Model) Exec(query *Query, dest interface{}) error {
	//TIME
	self.timer = time.Now()

	sql := query.toSQL()

	res, err := self.QueryDB(sql.SQL, sql.Args...)
	if err != nil {
		return err
	}

	// writing result into structs given by the caller
	err = self.DecodeAll(res, dest)
	if err != nil {
		return err
	}

	self.Duration = time.Since(self.timer)

	return err
}

// Create creates a new database instance (row) of model.
func (self *Model) Create(item interface{}) error {
	hitem, err := self.Encode(item)
	if err != nil {
		return err
	}

	vals := make([]interface{}, len(hitem))
	query := "INSERT INTO " + self.Name() + "("

	c := 0
	var tmp string

	for k, v := range hitem {
		query += k + ","
		vals[c] = v
		tmp += "?,"
		c++
	}

	query = query[:len(query)-1]
	query += ") VALUES(" + tmp[:len(tmp)-1] + ")"

	return self.db.Execute(query, vals...)
}

/*
 *
 *	HELPER FUNCTIONS
 *
 */

// Count returns the number of all database entries of this model.
func (self *Model) Count() (count int) {
	if !self.db.txOpen {
		err := self.db.BeginTransaction()
		if err != nil {
			return -1
		}
		defer self.db.EndTransaction()
	}

	err := self.db.tx.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM %s", self.Name())).Scan(&count)

	if err != nil {
		return -1
	}

	return count
}

// Letters returns string of first letters in the column named column.
func (self *Model) Letters(query *Query, column string) (string, error) {

	sqlQuery := query.toSQL()

	sql := fmt.Sprintf("SELECT DISTINCT SUBSTR(UPPER(%s),1,1) FROM (%s)",
		column, sqlQuery.SQL)

	res, err := self.QueryDB(sql, sqlQuery.Args...)
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
