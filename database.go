package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"
)

var (
	errNotAJsonFile        = errors.New("the provided file is not a json file")
	errTableNotFound       = errors.New("table does not exist")
	errColumnNotFound      = errors.New("table does not exist")
	errRecordAlreadyExists = errors.New("record already exists")
	errRecordNotFound      = errors.New("record does not exist")
)

type table string
type row map[string]any
type database struct {
	sync.Mutex
	Tables map[table][]row `json:"tables"`
}

func loadDB(path string) (*database, error) {
	// check if the file is a json file
	s := strings.Split(path, ".")
	if s[len(s)-1] != "json" {
		return nil, errNotAJsonFile
	}
	// read from file
	content, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not read database: '%s'", err.Error())
	}

	// decode json
	database := &database{
		Tables: make(map[table][]row),
	}
	if err = json.NewDecoder(content).Decode(&database.Tables); err != nil {
		return nil, fmt.Errorf("could not decode json: '%s'", err.Error())
	}

	return database, nil
}

func (db *database) tableExists(name table) bool {
	_, ok := db.Tables[name]
	return ok
}

func (db *database) getTable(name table) ([]row, error) {
	if !db.tableExists(name) {
		return nil, errTableNotFound
	}

	return db.Tables[name], nil
}

func (db *database) DeleteRowById(name table, id float64) error {
	if !db.tableExists(name) {
		return errTableNotFound
	}

	db.Lock()
	defer db.Unlock()

	for i, row := range db.Tables[name] {
		val, ok := row["id"]
		if !ok {
			return errColumnNotFound
		}

		if val == id {
			db.Tables[name] = append(db.Tables[name][:i], db.Tables[name][i+1:]...)
			return nil
		}
	}

	return errRecordNotFound
}

func (db *database) AddRow(name table, body row) (row, error) {
	// check if a table exists
	if !db.tableExists(name) {
		return nil, errTableNotFound
	}

	db.Lock()
	defer db.Unlock()

	db.Tables[name] = append(db.Tables[name], body)
	l := len(db.Tables[name])
	db.Tables[name][l-1]["id"] = db.Tables[name][l-2]["id"].(float64) + 1
	return db.GetRowById(name, db.Tables[name][l-1]["id"].(float64))
}

func (db *database) EditRowById(name table, id float64, body row) (row, error) {
	if !db.tableExists(name) {
		return nil, errTableNotFound
	}

	db.Lock()
	defer db.Unlock()

	for i, row := range db.Tables[name] {
		// check if key exist in row
		val, ok := row["id"]
		if !ok {
			return nil, errColumnNotFound
		}

		if val == id {
			db.Tables[name][i] = body
			return db.Tables[name][i], nil
		}
	}

	return nil, errRecordNotFound
}

func (db *database) GetRowById(name table, id float64) (row, error) {
	if !db.tableExists(name) {
		return nil, errTableNotFound
	}

	for i, row := range db.Tables[name] {
		// check if key exist in row
		val, ok := row["id"]
		if !ok {
			return nil, errColumnNotFound
		}

		if val == id {
			return db.Tables[name][i], nil
		}
	}

	return nil, errRecordNotFound
}

func quickSort[T cmp.Ordered](rows []row, col string, descendingOrder bool) {
	sort.Slice(rows, func(i, j int) bool {
		if descendingOrder {
			return rows[i][col].(T) > rows[j][col].(T)
		}
		return rows[i][col].(T) < rows[j][col].(T)
	})
}

func searchRecords[T comparable](data []row, col string, v T) ([]row, error) {
	result := make([]row, 0)

	for _, row := range data {
		val, ok := row[col]
		if !ok {
			return result, errColumnNotFound
		}

		if val == v {
			result = append(result, row)
		}
	}

	return slices.Clone(result), nil
}
