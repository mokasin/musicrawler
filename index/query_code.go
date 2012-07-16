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
	"strings"
)

var ErrWrongType error = errors.New("Wrong type.")

type Result map[string]interface{}

// Encode eats a pointer to a struct src and converts all exported fields into a
// map
// 		"field name" => <values>
func (self *Query) Encode(src interface{}) (Result, error) {
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
func (self *Query) Decode(src Result, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to struct.")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
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
			return fmt.Errorf("No column named '%s' found in query result. "+
				"Struct field '%s.%s %v' cannot be written.\n"+
				"Query result: %v",
				t.Field(i).Tag.Get("column"),
				t.Name(), t.Field(i).Name, v.Field(i).Kind(),
				src,
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

// DecodeAll decodes a slice of results. If there is only one result, dest can
// be a pointer to a struct. If there are more result, dest must be a pointer to
// a slice of structs.
func (self *Query) DecodeAll(src []Result, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr ||
		(v.Elem().Kind() != reflect.Slice &&
			v.Elem().Kind() != reflect.Struct) {
		return fmt.Errorf("dest must be a pointer to a slice or a struct.")
	}

	// if just one struct is given, it is unnecessary to return a slice
	if v.Elem().Kind() == reflect.Struct {
		if len(src) != 1 {
			return fmt.Errorf("Can't write data from database to single " +
				"struct. Got multiple results.")
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

// ExtractColumns extracts the column names from a struct's tags
// and returns them as a slice. str must be a pointer to a struct.
// 
// The tag must have the form
//
// 		column:"table:columname"
//
// 'table' is optional.
func (self *Query) ExtractColumns(str interface{}) (columns []string, err error) {
	v := reflect.ValueOf(str)

	if v.Kind() != reflect.Ptr ||
		(v.Elem().Kind() != reflect.Slice &&
			v.Elem().Kind() != reflect.Struct) {
		return nil, fmt.Errorf("str must be a pointer to a slice or a struct.")
	}

	v = v.Elem()
	t := v.Type()

	// if just one struct is given, it is unnecessary to return a slice
	if v.Kind() == reflect.Slice {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("column")

		// ignore fields with empty tag
		if tag == "" {
			continue
		}

		tag = strings.Replace(tag, ":", ".", 1)
		columns = append(columns, tag)
	}

	return columns, nil
}
