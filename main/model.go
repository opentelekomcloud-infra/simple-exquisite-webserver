package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/twinj/uuid"
	"log"
	"math/rand"
	"net"
	"regexp"
	"time"
	"unsafe"
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

const DataRandCS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ :;~`\\|/?.,<>{}()&*%$#@"

var src = rand.NewSource(time.Now().UnixNano())

func randomString(size int, prefix string) string {
	pLen := len(prefix)
	result := make([]byte, size)
	for i := 0; i < pLen; i++ {
		result[i] = prefix[i]
	}
	for i := pLen; i < size; i++ {
		result[i] = DataRandCS[src.Int63()%int64(len(DataRandCS))]
	}
	return *(*string)(unsafe.Pointer(&result)) // faster way to convert big slice to string
}

//CreateSomeEntities create `count` of random entities with given chars in data (20000 by default)
func CreateSomeEntities(count int, dataSize ...int) []Entity {
	size := 20000
	var data = make([]Entity, count)
	startTime := time.Now().Unix()
	if len(dataSize) > 0 {
		size = dataSize[0]
	}
	for i := 0; i < count; i++ {
		func() {
			data[i] = Entity{
				Uuid: uuid.NewV4().String(),
				Data: randomString(size, "RANDOM DATA: "),
			}
		}()
	}
	log.Print("Waiting for data to be generated...")
	log.Printf("Generated data in %vs", time.Now().Unix()-startTime)
	return data
}

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
