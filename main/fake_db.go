package main

import (
	"database/sql"
	"database/sql/driver"
)

type FakeDBDriver struct {
	conn *fakeDBConn
}

func NewFakeDBDriver() FakeDBDriver {
	return FakeDBDriver{newFakeDbConn()}
}

type fakeDBConn struct {
	tx   *fakeTx
	stmt *fakeStmt
}

type fakeTx struct {
}

type fakeStmt struct {
}

type fakeResult struct {
}

type fakeRows struct {
}

func (f fakeRows) Columns() []string {
	return []string{}
}

func (f fakeRows) Close() error {
	return nil
}

func (f fakeRows) Next(dest []driver.Value) error {
	return nil
}

func (f fakeResult) LastInsertId() (int64, error) {
	return 1, nil
}

func (f fakeResult) RowsAffected() (int64, error) {
	return 1, nil
}

func (f fakeStmt) Close() error {
	return nil
}

func (f fakeStmt) NumInput() int {
	return 1
}

func (f fakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	return fakeResult{}, nil
}

func (f fakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	return fakeRows{}, nil
}

func (f fakeTx) Commit() error {
	return nil
}

func (f fakeTx) Rollback() error {
	return nil
}

func (f fakeDBConn) Prepare(query string) (driver.Stmt, error) {
	return f.stmt, nil
}

func (f fakeDBConn) Close() error {
	return nil
}

func (f fakeDBConn) Begin() (driver.Tx, error) {
	return f.tx, nil
}

func newFakeDbConn() *fakeDBConn {
	return &fakeDBConn{new(fakeTx), new(fakeStmt)}
}

func (f FakeDBDriver) Open(name string) (driver.Conn, error) {
	if f.conn == nil {
		f.conn = newFakeDbConn()
	}
	return f.conn, nil
}

func init() {
	driver := NewFakeDBDriver()
	sql.Register("fake", driver)
}
