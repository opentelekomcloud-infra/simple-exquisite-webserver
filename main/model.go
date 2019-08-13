package main

import (
	"database/sql"
)

type entity struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (e *entity) getEntity(db *sql.DB) error {
	return db.QueryRow("SELECT name FROM entitys WHERE id=$1", e.ID).Scan(&e.Name)
}

func (e *entity) updateEntity(db *sql.DB) error {
	_, err := db.Exec("UPDATE entitys SET name=$1 WHERE id=$2", e.Name, e.ID)
	return err
}

func (e *entity) deleteEntity(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM entitys WHERE id=$1", e.ID)
	return err
}

func (e *entity) createEntity(db *sql.DB) error {
	// postgres doesn't return the last inserted ID so this is the workaround
	err := db.QueryRow(
		"INSERT INTO entitys(name) VALUES($1) RETURNING id",
		e.Name).Scan(&e.ID)
	return err
}

func getEntitys(db *sql.DB, start, count int) ([]entity, error) {
	rows, err := db.Query("SELECT id, name FROM entitys LIMIT $1 OFFSET $2", count, start)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	entitys := []entity{}

	for rows.Next() {
		var e entity
		if err := rows.Scan(&e.ID, &e.Name); err != nil {
			return nil, err
		}
		entitys = append(entitys, e)
	}

	return entitys, nil
}
