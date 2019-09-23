package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
)

type Entity struct {
	Uuid string `json:"uuid"`
	Data string `json:"data"`
}

// CreatePostgreDBIfNotExist create new database on given PostgreSQL instance if given DB does not exist on server
func CreatePostgreDBIfNotExist(dbName string, host string, port int, username string, password string) error {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, "postgres")
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT * FROM pg_database WHERE datname=$1 LIMIT 1", dbName)
	if (rows == nil) && (err == nil) {
		_, err = db.Exec("CREATE DATABASE $1", dbName)
	}
	return err
}

// CreateTable if not exists
func CreateTable(db *sql.DB) {
	if db == nil {
		return
	}
	sqlTable := `
	CREATE TABLE IF NOT EXISTS entity(
		uuid TEXT NOT NULL PRIMARY KEY,
		data TEXT
	);
	`

	_, err := db.Exec(sqlTable)
	if err != nil {
		panic(err)
	}
}

func (e *Entity) getEntity(db *sql.DB) error {
	if db == nil {
		return FakeGet(e)
	}

	return db.QueryRow("SELECT data FROM entity WHERE uuid like ($1)", e.Uuid).Scan(&e.Data)
}

func (e *Entity) updateEntity(db *sql.DB) error {
	if db == nil {
		return FakeUpdate(e)
	}

	_, err := db.Exec("UPDATE entity SET data=$1 WHERE uuid=$2", e.Data, e.Uuid)
	return err
}

func (e *Entity) deleteEntity(db *sql.DB) error {
	if db == nil {
		return FakeDel(e)
	}

	_, err := db.Exec("DELETE FROM entity WHERE uuid=$1", e.Uuid)
	return err
}

func (e *Entity) createEntity(db *sql.DB) error {
	if db == nil {
		return FakeNew(e)
	}

	// postgres doesn't return the last inserted Uuid so this is the workaround
	_, err := db.Exec(
		"INSERT INTO entity(uuid, data) VALUES($1, $2)", e.Uuid, e.Data)
	return err
}

func isConnectionError(err error) bool {
	switch err.(type) {
	default:
		return false
	case *net.OpError:
		return true
	}
}

func getEntities(db *sql.DB, start, count int) ([]Entity, error) {
	if db == nil {
		return FakeList(count, start)
	}

	rows, err := db.Query("SELECT uuid, data FROM entity LIMIT $1 OFFSET $2", count, start)
	if err != nil {
		if isConnectionError(err) {
			return nil, errors.New("can't connect to database")
		}
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	var entities []Entity

	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.Uuid, &e.Data); err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}

	return entities, nil
}
