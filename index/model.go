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
	"fmt"
	"reflect"
)

type Query map[string]interface{}
type Result map[string]interface{}

const (
	stStart = iota
	stWhere
	stLike
	stAll
	stOrdered
	stOrderedDsc
	stLimit
	stExecuted
)

type state struct {
	sql  string
	args []interface{}
	st   int
	err  error
}

// Model describes the basis of all models, i.e. database representations.
// TODO Documentation
type Model struct {
	index *Index
	name  string // name of table

	tx     *sql.Tx
	txOpen bool // flag true, when exists an open transaction

	state
}

// Constructor of Model. Needs a database index to work on it, a name of the
// model name and a fscan function, to read out the data from the database.
func NewModel(index *Index, name string) *Model {
	return &Model{
		index: index,
		name:  name,
	}
}

// Name returns the name of the model.
func (m *Model) Name() string {
	return m.name
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
func (m *Model) Encode(src interface{}) (Result, error) {
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
		if t.Field(i).Tag.Get("set") != "0" {
			res[t.Field(i).Tag.Get("name")] = v.Field(i).Interface()
		}
	}

	return res, nil
}

// Decode reads a map of type Result and a structure like
// 		"field name" => <value>
// and spits out a struct to dest.
func (m *Model) Decode(src Result, dest interface{}) error {
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

		// read out struct's tag to get the column name
		val, ok := src[t.Field(i).Tag.Get("name")]
		if !ok {
			return fmt.Errorf("No column named '%s' connected to "+
				"'%s.%s %v' found in database.",
				t.Field(i).Tag.Get("name"),
				t.Name(), t.Field(i).Name, v.Field(i).Kind(),
			)
		}

		dest = reflect.New(reflect.TypeOf(dest))

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
func (m *Model) DecodeAll(src []Result, dest interface{}) error {
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
		return m.Decode(src[0], v.Interface())
	}

	// Making slice big enough
	t := reflect.TypeOf(dest)
	v.Elem().Set(reflect.MakeSlice(t.Elem(), len(src), len(src)))

	// Feed Decode method with it
	for i := 0; i < v.Elem().Len(); i++ {
		err := m.Decode(src[i], v.Elem().Index(i).Addr().Interface())
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

// BeginTransaction starts a new database transaction.
func (m *Model) BeginTransaction() (err error) {
	if m.txOpen {
		return &ErrExistingTransaction{}
	}

	m.tx, err = m.index.db.Begin()
	if err != nil {
		return err
	}

	m.txOpen = true

	return nil
}

// EndTransaction closes a opened database transaction.
func (m *Model) EndTransaction() error {
	if !m.txOpen {
		return &ErrNoOpenTransaction{}
	}

	m.txOpen = false

	return m.tx.Commit()
}

// Query TODO: Documentation needed.
func (m *Model) Query(dest interface{}, sql string, args ...interface{}) error {
	if !m.txOpen {
		err := m.BeginTransaction()
		if err != nil {
			return err
		}
		defer m.EndTransaction()
	}

	fmt.Printf("QUERY: %s :: ", fmt.Sprintf(sql, "*"))
	fmt.Println(args...)

	// get the count of rows in the result
	var count int
	err := m.tx.QueryRow(fmt.Sprintf(sql, "COUNT(*)"), args...).Scan(&count)
	if err != nil {
		return err
	}

	// do the actual query
	rows, err := m.tx.Query(fmt.Sprintf(sql, "*"), args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// find out about the columns in the database
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// prepare result
	result := make([]Result, count)
	for i := 0; i < count; i++ {
		result[i] = make(Result)
	}

	// stupid trick, because rows.Scan will not take []interface as args
	col_vals := make([]interface{}, len(columns))
	col_args := make([]interface{}, len(columns))

	// initialize col_args
	for i := 0; i < len(columns); i++ {
		col_args[i] = &col_vals[i]
	}

	// read out columns and save them in a Result map
	c := 0
	for rows.Next() {
		if err := rows.Scan(col_args...); err != nil {
			return err
		}

		for i := 0; i < len(columns); i++ {
			result[c][columns[i]] = col_vals[i]
		}

		c++
	}

	// writing result into structs given by the caller
	err = m.DecodeAll(result, dest)
	if err != nil {
		return err
	}

	return rows.Err()
}

// Execute just executes sql query in global transaction.
func (m *Model) Execute(sql string, args ...interface{}) error {
	if !m.txOpen {
		err := m.BeginTransaction()
		if err != nil {
			return err
		}
		defer m.EndTransaction()
	}

	fmt.Printf("EXEC: %s :: ", sql)
	fmt.Println(args...)

	_, err := m.tx.Exec(sql, args...)
	return err
}

// Exec queries the database.
func (m *Model) Exec(dest interface{}) error {
	if m.st == stStart {
		return fmt.Errorf("Model is not in executable st.")
	}
	m.st = stStart
	m.state.err = m.Query(dest, m.sql, m.args...)

	return m.state.err
}

/*
*
*	HELPER FUNCTIONS
*
 */

// Count returns the number of all database entries of this model.
func (m *Model) Count() (count int) {
	if !m.txOpen {
		err := m.BeginTransaction()
		if err != nil {
			return -1
		}
		defer m.EndTransaction()
	}

	err := m.tx.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM %s", m.Name())).Scan(&count)

	if err != nil {
		return -1
	}

	return count
}

// Create creates a new database instance (row) of model.
func (m *Model) Create(item interface{}) error {

	hitem, err := m.Encode(item)
	if err != nil {
		return err
	}

	vals := make([]interface{}, len(hitem))
	query := "INSERT INTO " + m.Name() + "("

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

	return m.Execute(query, vals...)
}

// All TODO: Documentation needed.
func (m *Model) All() *Model {
	if m.st != stStart {
		m.state.err = fmt.Errorf("Can't call All() in that st.")
		return nil
	}

	m.st = stAll
	m.state.sql = fmt.Sprintf("SELECT %%s FROM %s", m.Name())

	return m
}

// Find TODO: Documentation needed.
func (m *Model) Find(ID int) *Model {
	return m.Where(Query{"ID": ID})
}

// Where TODO: Documentation needed.
func (m *Model) Where(query Query) *Model {
	if m.st != stStart {
		m.state.err = fmt.Errorf("Can't call Where() in that st.")
		return nil
	}

	m.st = stWhere

	var where string

	vals := make([]interface{}, len(query))

	c := 0
	for key, val := range query {
		vals[c] = val
		where += fmt.Sprintf("%s.%s = ? AND ", m.Name(), key)
		c++
	}
	where = where[:len(where)-5]

	m.state.sql = fmt.Sprintf("SELECT %%s FROM %s WHERE %s", m.Name(), where)
	m.state.args = vals

	return m
}

// Where TODO: Documentation needed.
func (m *Model) Like(query Query) *Model {
	if m.st != stStart {
		m.state.err = fmt.Errorf("Can't call Where() on state %d.", m.st)
		return nil
	}

	m.st = stLike

	var where string

	vals := make([]interface{}, len(query))

	c := 0
	for key, val := range query {
		vals[c] = val
		where += fmt.Sprintf("%s.%s LIKE ? AND ", m.Name(), key)
		c++
	}
	where = where[:len(where)-5]

	m.state.sql = fmt.Sprintf("SELECT %%s FROM %s WHERE %s", m.Name(), where)
	m.state.args = vals

	return m
}

func (m *Model) Limit(number int) *Model {
	switch m.st {
	case stAll, stWhere, stLike:
		m.state.sql += " LIMIT ?"
		m.state.args = append(m.state.args, number)
	default:
		m.state.err = fmt.Errorf("Cannot call Limit() on state %d.", m.st)
		return nil
	}

	return m
}

/*
*
*	DEFINITIONS
*
 */
// Error types

type ErrNoOpenTransaction struct{}

func (e *ErrNoOpenTransaction) Error() string { return "No open transaction." }

type ErrExistingTransaction struct{}

func (e *ErrExistingTransaction) Error() string {
	return "There is an existing transaction."
}

type ErrWrongType struct{}

func (e *ErrWrongType) Error() string { return "Wrong type." }
