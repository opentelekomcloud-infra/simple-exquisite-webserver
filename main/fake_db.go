package main

import (
	"database/sql"
)

var FakeDataStorage map[string]Entity

func FakeGet(e *Entity) error {
	ent, ok := FakeDataStorage[e.Uuid]
	if !ok {
		return sql.ErrNoRows
	}
	e.Data = ent.Data
	return nil
}

func FakeList(length int, offset int) ([]Entity, error) {
	storLength := len(FakeDataStorage)
	values := make([]Entity, storLength)
	i := 0

	if max := storLength - offset; length > max {
		length = max
	}
	if offset > storLength {
		offset = storLength
	}

	for _, val := range FakeDataStorage {
		values[i] = val
		i++
	}
	return values[offset : offset+length], nil
}

func FakeNew(e *Entity) error {
	FakeDataStorage[e.Uuid] = *e
	return nil
}

func FakeDel(e *Entity) error {
	ent, ok := FakeDataStorage[e.Uuid]
	if !ok {
		return sql.ErrNoRows
	}
	delete(FakeDataStorage, ent.Uuid)
	return nil
}

func FakeUpdate(e *Entity) error {
	_, ok := FakeDataStorage[e.Uuid]
	if !ok {
		return sql.ErrNoRows
	}
	FakeDataStorage[e.Uuid] = *e
	return nil
}
