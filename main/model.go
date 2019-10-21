package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
)

// CreatePostgreDBIfNotExist create new database on given PostgreSQL instance if given DB does not exist on server
func CreatePostgreDBIfNotExist(dbName string, host string, port int, username string, password string) error {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, "postgres")
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}
	rowSize := 0
	err = db.QueryRow("SELECT COUNT(*) FROM pg_database WHERE datname = $1", dbName).Scan(&rowSize)
	if err != nil {
		panic(err)
	}
	if rowSize == 0 {
		log.Printf("Database %s does not exist, DB to be created", dbName)
		pattern := "^[a-zA-Z_]+$"
		seemsOk, _ := regexp.MatchString(pattern, dbName)
		if !seemsOk {
			return fmt.Errorf("invalid database name: %s. Database name must match %s pattern", dbName, pattern)
		}
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)) // This is insecure, but it won't work other way
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

	_, err := db.Exec("UPDATE entity SET data = $1 WHERE uuid = $2", e.Data, e.Uuid)
	return err
}

func (e *Entity) deleteEntity(db *sql.DB) error {
	if db == nil {
		return FakeDel(e)
	}

	_, err := db.Exec("DELETE FROM entity WHERE uuid = $1", e.Uuid)
	return err
}

func (e *Entity) createEntity(db *sql.DB) error {
	if db == nil {
		return FakeNew(e)
	}

	// postgres doesn't return the last inserted Uuid so this is the workaround
	_, err := db.Exec(
		"INSERT INTO entity(uuid, data) VALUES ($1, $2)", e.Uuid, e.Data)
	return err
}

//AddEntities — add multiple entities in single transaction
func AddEntities(db *sql.DB, entities []Entity) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	count := len(entities)
	baseQ, _ := tx.Prepare("INSERT INTO entity (uuid, data) VALUES ($1, $2)")
	for i := 0; i < count; i++ {
		entity := entities[i]
		_, err = baseQ.Exec(entity.Data, entity.Uuid)
		if err != nil {
			_ = tx.Rollback()
			log.Fatalf("Unable to insert data %v into entities table", entity)
			return err
		}
	}
	return tx.Commit()
}

func isConnectionError(err error) bool {
	switch err.(type) {
	default:
		return false
	case *net.OpError:
		return true
	}
}

func getEntities(db *sql.DB, count int, filter string) ([]Entity, error) {
	if db == nil {
		return FakeList(count, 0)
	}

	rows, err := db.Query(`
		SELECT uuid, data
			FROM entity
			WHERE data LIKE $2
			LIMIT $1`,
		count, filter)
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
