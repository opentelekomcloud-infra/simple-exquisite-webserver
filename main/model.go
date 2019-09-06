package main

import (
	"database/sql"
	"errors"
	"net"
)

type entity struct {
	ID   string `json:"uuid"`
	Data string `json:"data"`
}

// CreateTable if not exists
func CreateTable(db *sql.DB) {
	sqlTable := `
	CREATE TABLE IF NOT EXISTS entities(
		Id TEXT NOT NULL PRIMARY KEY,
		Data TEXT
	);
	`

	_, err := db.Exec(sqlTable)
	if err != nil {
		panic(err)
	}
}

func (e *entity) getEntity(db *sql.DB) error {
	return db.QueryRow("SELECT Data FROM entities WHERE Id like ($1)", e.ID).Scan(&e.Data)
}

func (e *entity) updateEntity(db *sql.DB) error {
	_, err := db.Exec("UPDATE entities SET Data=$1 WHERE Id=$2", e.Data, e.ID)
	return err
}

func (e *entity) deleteEntity(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM entities WHERE Id=$1", e.ID)
	return err
}

func (e *entity) createEntity(db *sql.DB) error {
	// postgres doesn't return the last inserted ID so this is the workaround
	_, err := db.Exec(
		"INSERT INTO entities(Id, Data) VALUES($1, $2)", e.ID, e.Data)
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
	rows, err := db.Query("SELECT Id, Data FROM entities LIMIT $1 OFFSET $2", count, start)
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
		if err := rows.Scan(&e.ID, &e.Data); err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}

	return entities, nil
}
