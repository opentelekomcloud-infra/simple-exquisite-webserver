package main

import (
	"database/sql"
	"regexp"
	"strings"
)

var FakeDataStorage = make(map[string]Entity)

func FakeGet(e *Entity) error {
	ent, ok := FakeDataStorage[e.Uuid]
	if !ok {
		return sql.ErrNoRows
	}
	e.Data = ent.Data
	return nil
}

func FakeList(length int, filter string) ([]Entity, error) {
	storLength := len(FakeDataStorage)
	values := make([]Entity, 0, storLength)

	reFilter := strings.Replace(filter, "%", ".*", -1)
	if reFilter == "" {
		reFilter = ".*"
	}
	for _, val := range FakeDataStorage {
		match, _ := regexp.MatchString(reFilter, val.Data)
		if match {
			values = append(values, val)
		}
	}
	if length > storLength {
		length = storLength
	}
	vLength := len(values)
	if length > vLength {
		length = vLength
	}
	return values[:length], nil
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
