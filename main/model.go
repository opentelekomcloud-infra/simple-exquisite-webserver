package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
)

type entity struct {
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

func (e *entity) getEntity(db *sql.DB) error {
	return db.QueryRow("SELECT data FROM entity WHERE id like ($1)", e.Uuid).Scan(&e.Data)
}

func (e *entity) updateEntity(db *sql.DB) error {
	_, err := db.Exec("UPDATE entity SET data=$1 WHERE id=$2", e.Data, e.Uuid)
	return err
}

func (e *entity) deleteEntity(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM entity WHERE id=$1", e.Uuid)
	return err
}

func (e *entity) createEntity(db *sql.DB) error {
	// postgres doesn't return the last inserted Uuid so this is the workaround
	_, err := db.Exec(
		"INSERT INTO entity(id, data) VALUES($1, $2)", e.Uuid, e.Data)
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

func getEntities(db *sql.DB, start, count int) ([]entity, error) {
	rows, err := db.Query("SELECT id, data FROM entity LIMIT $1 OFFSET $2", count, start)
	if err != nil {
		if isConnectionError(err) {
			return nil, errors.New("can't connect to database")
		}
		return nil, err
	}

	defer rows.Close()

	var entities []entity

	for rows.Next() {
		var e entity
		if err := rows.Scan(&e.Uuid, &e.Data); err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}

	return entities, nil
}
