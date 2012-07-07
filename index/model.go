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
	"fmt"
	"reflect"
	"time"
)

// Defines one item, that is return by an model.
type Item struct {
	Index *Index
}

type Query map[string]interface{}
type Result map[string]interface{}

const (
	stStart = iota
	stWhere
	stLike
	stAll
	stOrder
	stOrderedDsc
	stLimit
	stOffset
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

	timer    time.Time
	Duration time.Duration

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

		// Give it a pointer to the index. This is a little bit ugly.
		if t.Field(i).Name == "Item" {
			v.Field(i).FieldByName("Index").Set(reflect.ValueOf(m.index))
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

// Query TODO: Documentation needed.
func (m *Model) Query(sql string, args ...interface{}) ([]Result, error) {
	if !m.index.txOpen {
		err := m.index.BeginTransaction()
		if err != nil {
			return nil, err
		}
		defer m.index.EndTransaction()
	}

	// Just for debugging
	fmt.Printf("QUERY: %s :: ", sql)
	fmt.Println(args...)

	// do the actual query
	rows, err := m.index.tx.Query(sql, args...)
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

// Execute just executes sql query in global transaction.
func (m *Model) Execute(sql string, args ...interface{}) error {
	if !m.index.txOpen {
		err := m.index.BeginTransaction()
		if err != nil {
			return err
		}
		defer m.index.EndTransaction()
	}

	// Just for debugging
	fmt.Printf("EXEC: %s :: ", sql)
	fmt.Println(args...)

	_, err := m.index.tx.Exec(sql, args...)
	return err
}

// Exec queries the database.
func (m *Model) Exec(dest interface{}) error {
	if m.st == stStart {
		return fmt.Errorf("Model is not in an executable state.")
	}

	if m.state.err != nil {
		return m.state.err
	}

	m.st = stStart
	m.state.err = nil

	//TIME
	m.timer = time.Now()

	res, err := m.Query(m.sql, m.state.args...)
	if err != nil {
		return err
	}

	m.state.args = make([]interface{}, 0)

	// writing result into structs given by the caller
	err = m.DecodeAll(res, dest)
	if err != nil {
		return err
	}

	m.Duration = time.Since(m.timer)

	return err
}

/*
 *
 *	HELPER FUNCTIONS
 *
 */

// Count returns the number of all database entries of this model.
func (m *Model) Count() (count int) {
	if !m.index.txOpen {
		err := m.index.BeginTransaction()
		if err != nil {
			return -1
		}
		defer m.index.EndTransaction()
	}

	err := m.index.tx.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM %s", m.Name())).Scan(&count)

	if err != nil {
		return -1
	}

	return count
}

// Letters returns string of first letters in the column named column.
func (m *Model) Letters(column string) (string, error) {
	if m.st == stStart {
		return "", fmt.Errorf("Cannot call Letters() on state %d.", m.st)
	}

	if m.state.err != nil {
		return "", m.state.err
	}

	query := fmt.Sprintf("SELECT DISTINCT SUBSTR(UPPER(%s),1,1) FROM (%s)",
		column, m.state.sql)

	m.st = stStart
	m.state.err = nil
	m.state.sql = ""

	res, err := m.Query(query, m.state.args...)
	if err != nil {
		return "", err
	}

	m.state.args = make([]interface{}, 0)

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

// All just returns all rows in table.
func (m *Model) All() *Model {
	if m.st != stStart {
		m.state.err = fmt.Errorf("Can't call All() in state %d.", m.st)
		return nil
	}

	m.st = stAll
	m.state.sql = fmt.Sprintf("SELECT * FROM %s", m.Name())

	return m
}

// Find returns row with id ID.
func (m *Model) Find(ID int) *Model {
	return m.WhereQ(Query{"ID": ID})
}

// Where returns all rows that fullfil the constrictions given in constrictions.
//
// Example:
// 
// 		Where("column1 = ? AND column2 = ?", astring, bstring)
func (m *Model) Where(constriction string, args ...interface{}) *Model {
	switch m.st {
	case stStart:
		m.state.sql = fmt.Sprintf("SELECT * FROM %s WHERE %s", m.Name(), constriction)
	case stWhere, stLike:
		m.state.sql += " AND " + constriction
	default:
		m.state.err = fmt.Errorf("Can't call Where() on state %d.", m.st)
		return nil
	}

	m.st = stWhere
	m.state.args = append(m.state.args, args...)

	return m
}

// WhereQ returns all rows that fullfil the constrictions given in
// constrictions.
//
// It is like the Where-method but takes a Query map as argument.
// Example:
//
// 		WhereQ(Query{column1: astring, column2: bstring})
func (m *Model) WhereQ(query Query) *Model {
	switch m.st {
	case stStart:
		m.state.sql = fmt.Sprintf("SELECT * FROM %s WHERE ", m.Name())
	case stWhere, stLike:
		m.state.sql += " AND "
	default:
		m.state.err = fmt.Errorf("Can't call Where() on state %d.", m.st)
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

	m.state.sql += where
	m.state.args = append(m.state.args, vals...)

	return m
}

// LikeQ returns all rows that fullfil the constrictions given in constrictions.
//
// It is like WhereQ, but the wildcard '%' is allowed in the query.
// Example:
//
// 		Like(Query{column1: "Some%", column2: "%string%"})
func (m *Model) LikeQ(query Query) *Model {
	switch m.st {
	case stStart:
		m.state.sql = fmt.Sprintf("SELECT * FROM %s WHERE ", m.Name())
	case stWhere, stLike:
		m.state.sql += " AND "
	default:
		m.state.err = fmt.Errorf("Can't call Where() on state %d.", m.st)
		return nil
	}

	m.st = stWhere

	var where string

	vals := make([]interface{}, len(query))

	c := 0
	for key, val := range query {
		vals[c] = val
		where += fmt.Sprintf("%s.%s LIKE ? AND ", m.Name(), key)
		c++
	}
	where = where[:len(where)-5]

	m.state.sql += where
	m.state.args = append(m.state.args, vals...)

	return m

}

// Limit limits the number of returned rows to number.
func (m *Model) Limit(number int) *Model {
	switch m.st {
	case stAll, stWhere, stLike, stOrder:
		m.state.sql += " LIMIT ?"
		m.state.args = append(m.state.args, number)
	default:
		m.state.err = fmt.Errorf("Cannot call Limit() on state %d.", m.st)
		return nil
	}

	m.st = stLimit

	return m
}

// Offset returns rows with an offset offset.
func (m *Model) Offset(offset int) *Model {
	switch m.st {
	case stAll, stWhere, stLike, stOrder, stLimit:
		m.state.sql += " OFFSET ?"
		m.state.args = append(m.state.args, offset)
	default:
		m.state.err = fmt.Errorf("Cannot call Offset() on state %d.", m.st)
		return nil
	}

	m.st = stOffset

	return m
}

// OrderedBy is ordering the result ascending by column.
func (m *Model) OrderBy(column string) *Model {
	switch m.st {
	case stAll, stWhere, stLike:
		m.state.sql += " ORDER BY UPPER(" + column + ")"
	default:
		m.state.err = fmt.Errorf("Cannot call OrderedBy() on state %d.", m.st)
		return nil
	}

	m.st = stOrder

	return m
}

/*
 *
 *	DEFINITIONS
 *
 */
// Error types

type ErrWrongType struct{}

func (e *ErrWrongType) Error() string { return "Wrong type." }
