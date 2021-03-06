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
	. "github.com/mokasin/musicrawler/lib/database"
	"strings"
)

type join struct {
	OnTable, OnFieldName, OwnTable, OwnFieldName string
}

type where struct {
	Constriction string
	Value        interface{}
}

type wherein struct {
	FieldName string
	Values    []interface{}
}

type like struct {
	Constriction string
	Value        interface{}
}

type sortDirection int

const (
	ascending sortDirection = iota
	descending
)

var sortDirectionToSQL = map[sortDirection]string{
	ascending:  "ASC",
	descending: "DESC",
}

type order struct {
	FieldName string
	Direction sortDirection
}

// Query provides methods to build an SQL-Query.
type Query struct {
	table string
	db    *Database

	cols    []string
	join    []join
	where   []where
	wherein []wherein
	like    []like
	order   []order
	limit   uint
	offset  uint

	err error
}

// New creates a new Query for a specifig table.
func New(db *Database, table string) *Query {
	return &Query{db: db, table: table}
}

// sqlQuery represents a SQL query with arguments.
type sqlQuery struct {
	SQL  string
	Args []interface{}
}

// toSQL encodes the query into an SQL-Query.
func (self *Query) toSQL() *sqlQuery {
	var cols, join, where, wherein, order, limit, offset string
	sql := &sqlQuery{}

	// set columns
	if len(self.cols) == 0 {
		cols = self.table + ".*"
	} else {
		for i, v := range self.cols {
			cols += v + " AS \"" + strings.Replace(v, ".", ":", -1) + "\""
			if i < len(self.cols)-1 {
				cols += ","
			}
		}
	}

	// add join statement
	for _, v := range self.join {
		join += " JOIN " + v.OnTable + " ON " +
			v.OwnTable + "." + v.OwnFieldName + " = " +
			v.OnTable + "." + v.OnFieldName

	}

	// add constriction if available
	if len(self.where) > 0 || len(self.like) > 0 || len(self.wherein) > 0 {
		where = " WHERE"
	}

	// add boolean constrictions
	for i, v := range self.where {
		where += " " + v.Constriction + " ?"

		if i < len(self.where)-1 {
			where += " AND"
		}

		sql.Args = append(sql.Args, v.Value)
	}

	if len(self.where) > 0 && len(self.wherein) > 0 {
		where += " AND"
	}

	// add in set constriction
	for i, v := range self.wherein {
		wherein += " " + v.FieldName + " IN ("
		for i := 0; i < len(v.Values); i++ {
			wherein += "?,"
		}
		// trim the last comma
		wherein = wherein[:len(wherein)-1]
		wherein += ")"

		if i < len(self.where)-1 {
			wherein += " AND"
		}

		sql.Args = append(sql.Args, v.Values...)
	}

	if len(self.wherein) > 0 && len(self.like) > 0 {
		where += " AND"
	}

	// add wildcard constriction
	for i, v := range self.like {
		where += " " + v.Constriction + " LIKE ?"

		if i < len(self.like)-1 {
			where += " AND"
		}

		sql.Args = append(sql.Args, v.Value)
	}

	// add ordering statement
	if len(self.order) > 0 {
		order = " ORDER BY "
	}
	for i, v := range self.order {
		order += v.FieldName + " " + sortDirectionToSQL[v.Direction]

		if i < len(self.order)-1 {
			order += ","
		}
	}

	// add limit...
	if self.limit != 0 {
		limit = " LIMIT ?"
		sql.Args = append(sql.Args, self.limit)
	}

	// ...and offset
	if self.offset != 0 {
		offset = " OFFSET ?"
		sql.Args = append(sql.Args, self.offset)
	}

	// put everything together
	sql.SQL = "SELECT " + cols + " FROM " +
		self.table + join + where + wherein + order + limit + offset

	return sql
}

// columns returns a derivated Query that returns only the given cols. A '*'
// selects all available columns.
//
// cols must have the format:
//
// 		<table>.<column>
//
// Multiple calls overwrite the previous one.
func (self *Query) columns(cols ...string) *Query {
	self.cols = cols
	return self
}

// Join returns a derivated Query that joins onTable and ownTable with respect
// to the fields onFieldName and ownFieldname.
// If ownFieldname is an empty string "", self.table is used.
func (self *Query) Join(onTable, onFieldName, ownTable, ownFieldName string) *Query {
	if ownTable == "" {
		ownTable = self.table
	}

	self.join = append(self.join, join{
		OnTable:      onTable,
		OnFieldName:  onFieldName,
		OwnTable:     ownTable,
		OwnFieldName: ownFieldName,
	})

	return self
}

// Where returns a derivated Query with an applied constriction. The
// constriction must be a string of the form
//
// 		<fieldName> <operator>
//
// Multiple wheres are concatenation with an AND. The fieldName is compared to the
// value.
//
// Example:
//
// 		Where("ID >", 5)
//
func (self *Query) Where(constriction string, value interface{}) *Query {
	self.where = append(self.where, where{constriction, value})
	return self
}

// Find is just an alias for matching the ID.
func (self *Query) Find(ID int) *Query {
	return self.Where("ID =", ID)
}

// WhereIn returns a derivated Query with an applied constriction. The
// constriction must be a string of the form
//
// 		<fieldName>
//
// Multiple calls are concatenation with an AND. The fieldName is a set of
// values.
// Example:
//
// 		WhereIn("ID", 5, 7, 3)
//
func (self *Query) WhereIn(fieldname string, values ...interface{}) *Query {
	self.wherein = append(self.wherein, wherein{fieldname, values})
	return self
}

// Like returns a derivated Query with an applied constriction. The
// constriction must be a string of the form
//
// 		<fieldName> <operator>
//
// Multiple wheres are concatenation with an AND. The fieldName is compared to the
// value. Use % as a wildcard that matches a value of arbitrary length, and _ to
// match just a single character.
//
// Example:
//
// 		Like("name", "A%")
//
func (self *Query) Like(constriction string, value string) *Query {
	self.like = append(self.like, like{constriction, value})
	return self
}

// Order returns a derivated Query that the results are ordered by the given
// fieldName. If the fieldName is prefixed with a minus sign '-' the ordering is
// descending. Multiple orderings are applied in order of call.
func (self *Query) Order(fieldName string) *Query {
	var o order
	if strings.HasPrefix(fieldName, "-") {
		o = order{
			FieldName: fieldName[1:],
			Direction: descending,
		}
	} else {
		o = order{
			FieldName: fieldName,
			Direction: ascending,
		}
	}

	self.order = append(self.order, o)

	return self
}

// Limit returns a derivated Query that the number of results are limited to
// limit. Multiple calls just overrides the previous one.
func (self *Query) Limit(limit uint) *Query {
	self.limit = limit
	return self
}

// Offset returns a derivated Query that has an offset of how many results are
// are skipped. Multiple calls just override the previous one.
func (self *Query) Offset(offset uint) *Query {
	self.offset = offset
	return self
}
