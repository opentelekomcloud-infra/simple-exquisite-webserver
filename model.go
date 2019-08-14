package main

import (
	"database/sql"
)

type entity struct {
	ID   int    `json:"id"`
	Data string `json:"Data"`
}

func (e *entity) getEntity(db *sql.DB) error {
	return db.QueryRow("SELECT Data FROM entitys WHERE id=$1", e.ID).Scan(&e.Data)
}

func (e *entity) updateEntity(db *sql.DB) error {
	_, err := db.Exec("UPDATE entitys SET Data=$1 WHERE id=$2", e.Data, e.ID)
	return err
}

func (e *entity) deleteEntity(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM entitys WHERE id=$1", e.ID)
	return err
}

func (e *entity) createEntity(db *sql.DB) error {
	// postgres doesn't return the last inserted ID so this is the workaround
	err := db.QueryRow(
		"INSERT INTO entitys(Data) VALUES($1) RETURNING id",
		e.Data).Scan(&e.ID)
	return err
}

func getEntitys(db *sql.DB, start, count int) ([]entity, error) {
	rows, err := db.Query("SELECT id, Data FROM entitys LIMIT $1 OFFSET $2", count, start)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	entitys := []entity{}

	for rows.Next() {
		var e entity
		if err := rows.Scan(&e.ID, &e.Data); err != nil {
			return nil, err
		}
		entitys = append(entitys, e)
	}

	return entitys, nil
}
