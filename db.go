// Package database is a library help for interact with database by model
package gomodel

import (
	"database/sql"

	"github.com/cosiner/gohper/bitset"
)

type (
	// Model represent a database model
	Model interface {
		Table() string
		// Vals store values of fields to given slice
		Vals(fields uint, vals []interface{})
		Ptrs(fields uint, ptrs []interface{})
	}

	// DB holds database connection, all typeinfos, and sql cache
	DB struct {
		// driver string
		*sql.DB
		types map[string]*TypeInfo
		*Cacher

		ModelCount int
	}
)

var (
	FieldCount = bitset.BitCountUint
)

// Open create a database manager and connect to database server
func Open(driver, dsn string, maxIdle, maxOpen int) (*DB, error) {
	db := NewDB()
	err := db.Connect(driver, dsn, maxIdle, maxOpen)

	return db, err
}

// New create a new db
func NewDB() *DB {
	return &DB{
		types:      make(map[string]*TypeInfo),
		ModelCount: 10,
	}
}

// Connect connect to database server
func (db *DB) Connect(driver, dsn string, maxIdle, maxOpen int) error {
	db_, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}

	db_.SetMaxIdleConns(maxIdle)
	db_.SetMaxOpenConns(maxOpen)
	db.DB = db_
	db.Cacher = NewCacher(Types, db)

	return nil
}

// registerType save type info of model
func (db *DB) registerType(v Model, table string) *TypeInfo {
	ti := parseTypeInfo(v, db)
	db.types[table] = ti

	return ti
}

// TypeInfo return type information of given model
// if type info not exist, it will parseTypeInfo it and save type info
func (db *DB) TypeInfo(v Model) *TypeInfo {
	table := v.Table()
	if ti, has := db.types[table]; has {
		return ti
	}

	return db.registerType(v, table)
}

func FieldVals(fields uint, v Model) []interface{} {
	args := make([]interface{}, FieldCount(fields))
	v.Vals(fields, args)

	return args
}

func FieldPtrs(fields uint, v Model) []interface{} {
	ptrs := make([]interface{}, FieldCount(fields))
	v.Ptrs(fields, ptrs)

	return ptrs
}

func (db *DB) Insert(v Model, fields uint, needId bool) (int64, error) {
	return db.ArgsInsert(v, fields, needId, FieldVals(fields, v)...)
}

func (db *DB) ArgsInsert(v Model, fields uint, needId bool, args ...interface{}) (int64, error) {
	stmt, err := db.TypeInfo(v).InsertStmt(fields)

	return StmtExec(stmt, err, needId, args...)
}

func (db *DB) Update(v Model, fields, whereFields uint) (int64, error) {
	c1, c2 := FieldCount(fields), FieldCount(whereFields)
	args := make([]interface{}, c1+c2)
	v.Vals(fields, args)
	v.Vals(whereFields, args[c1:])

	return db.ArgsUpdate(v, fields, whereFields, args...)
}

func (db *DB) ArgsUpdate(v Model, fields, whereFields uint, args ...interface{}) (int64, error) {
	stmt, err := db.TypeInfo(v).UpdateStmt(fields, whereFields)

	return StmtExec(stmt, err, false, args...)
}

func (db *DB) Delete(v Model, whereFields uint) (int64, error) {
	return db.ArgsDelete(v, whereFields, FieldVals(whereFields, v)...)
}

func (db *DB) ArgsDelete(v Model, whereFields uint, args ...interface{}) (int64, error) {
	stmt, err := db.TypeInfo(v).DeleteStmt(whereFields)

	return StmtExec(stmt, err, false, args...)
}

// One select one row from database
func (db *DB) One(v Model, fields, whereFields uint) error {
	stmt, err := db.TypeInfo(v).SelectOneStmt(fields, whereFields)
	scanner, rows := StmtQuery(stmt, err, FieldVals(whereFields, v)...)

	return scanner.One(rows, FieldPtrs(fields, v)...)
}

func (db *DB) Limit(s Store, v Model, fields, whereFields uint, start, count int) error {
	c := FieldCount(whereFields)
	args := make([]interface{}, c+2)
	v.Vals(whereFields, args)
	args[c], args[c+1] = start, count

	return db.ArgsLimit(s, v, fields, whereFields, args...)
}

func (db *DB) ArgsLimit(s Store, v Model, fields, whereFields uint, args ...interface{}) error {
	stmt, err := db.TypeInfo(v).SelectLimitStmt(fields, whereFields)
	scanner, rows := StmtQuery(stmt, err, args...)

	return scanner.Limit(rows, s, args[len(args)-1].(int))
}

func (db *DB) All(s Store, v Model, fields, whereFields uint) error {
	return db.ArgsAll(s, v, fields, whereFields, FieldVals(whereFields, v)...)
}

// ArgsAll select all rows, the last two argument must be "start" and "count"
func (db *DB) ArgsAll(s Store, v Model, fields, whereFields uint, args ...interface{}) error {
	stmt, err := db.TypeInfo(v).SelectAllStmt(fields, whereFields)
	scanner, rows := StmtQuery(stmt, err, args...)

	return scanner.All(rows, s, db.ModelCount)
}

// Count return count of rows for model, arguments was extracted from Model
func (db *DB) Count(v Model, whereFields uint) (count int64, err error) {
	return db.ArgsCount(v, whereFields, FieldVals(whereFields, v)...)
}

//Args Count return count of rows for model use custome arguments
func (db *DB) ArgsCount(v Model, whereFields uint,
	args ...interface{}) (count int64, err error) {
	ti := db.TypeInfo(v)

	stmt, err := ti.CountStmt(whereFields)
	scanner, rows := StmtQuery(stmt, err, args...)

	err = scanner.One(rows, &count)

	return
}

// ExecUpdate execute a update operation, return resolved result
func (db *DB) ExecUpdate(s string, needId bool, args ...interface{}) (int64, error) {
	res, err := db.Exec(s, args...)

	return ResolveResult(res, err, needId)
}

// StmtExec execute stmt with given arguments and resolve the result if error is nil
func StmtExec(stmt *sql.Stmt, err error, needId bool, args ...interface{}) (int64, error) {
	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(args...)

	return ResolveResult(res, err, needId)
}

// StmtQuery execute the query stmt, error stored in Scanner
func StmtQuery(stmt *sql.Stmt, err error, args ...interface{}) (Scanner, *sql.Rows) {
	if err != nil {
		return Scanner{err}, nil
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return Scanner{err}, nil
	}

	return normalScanner, rows
}

// ResolveResult resolve sql result, if need id, return last insert id
// else return affected rows count
func ResolveResult(res sql.Result, err error, needId bool) (int64, error) {
	if err != nil {
		return 0, err
	}

	if needId {
		return res.LastInsertId()
	} else {
		return res.RowsAffected()
	}
}
