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
	"strings"
)

type where struct {
	Constriction string
	Value        interface{}
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

	where  []where
	like   []like
	order  []order
	limit  uint
	offset uint

	err error
}

// NewQuery creates a new Query for a specifig table.
func NewQuery(table string) *Query {
	return &Query{table: table}
}

// sqlQuery represents a SQL query with arguments.
type sqlQuery struct {
	SQL  string
	Args []interface{}
}

// toSQL encodes the query into an SQL-Query.
func (q *Query) toSQL() *sqlQuery {
	var where, order, limit, offset string
	sql := &sqlQuery{}

	if len(q.where) > 0 || len(q.like) > 0 {
		where = " WHERE"
	}

	for i, v := range q.where {
		where += " " + v.Constriction + " ?"

		if i < len(q.where)-1 {
			where += " AND"
		}

		sql.Args = append(sql.Args, v.Value)
	}

	if len(q.where) > 0 && len(q.like) > 0 {
		where += " AND"
	}

	for i, v := range q.like {
		where += " " + v.Constriction + " LIKE ?"

		if i < len(q.like)-1 {
			where += " AND"
		}

		sql.Args = append(sql.Args, v.Value)
	}

	if len(q.order) > 0 {
		order = " ORDER BY "
	}
	for i, v := range q.order {
		order += v.FieldName + " " + sortDirectionToSQL[v.Direction]

		if i < len(q.order)-1 {
			order += ","
		}
	}

	if q.limit != 0 {
		limit = " LIMIT ?"
		sql.Args = append(sql.Args, q.limit)
	}

	if q.offset != 0 {
		offset = " OFFSET ?"
		sql.Args = append(sql.Args, q.offset)
	}

	sql.SQL = "SELECT * FROM " + q.table + where + order + limit + offset

	return sql
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
func (q *Query) Where(constriction string, value interface{}) *Query {
	q.where = append(q.where, where{constriction, value})
	return q
}

// Find is just an alias for matching the ID.
func (q *Query) Find(ID int) *Query {
	return q.Where("ID =", ID)
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
func (q *Query) Like(constriction string, value string) *Query {
	q.like = append(q.like, like{constriction, value})
	return q
}

// Order returns a derivated Query that the results are ordered by the given
// fieldName. If the fieldName is prefixed with a minus sign '-' the ordering is
// descending. Multiple orderings are applied in order of call.
func (q *Query) Order(fieldName string) *Query {
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

	q.order = append(q.order, o)

	return q
}

// Limit returns a derivated Query that the number of results are limited to
// limit. Multiple calls just overrides the previous one.
func (q *Query) Limit(limit uint) *Query {
	q.limit = limit
	return q
}

// Offset returns a derivated Query that has an offset of how many results are
// are skipped. Multiple calls just override the previous one.
func (q *Query) Offset(offset uint) *Query {
	q.offset = offset
	return q
}
