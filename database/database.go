package database

import (
	"encoding/gob"
	"os"
	"sync"
)

type Database struct {
	data map[string]any
	lock sync.RWMutex
}

func NewDatabase() *Database {
	return &Database{
		data: make(map[string]any),
	}
}

func (db *Database) Set(key string, value any) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.data[key] = value
}

func (db *Database) Get(key string) (any, bool) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	value, ok := db.data[key]

	return value, ok
}

func (db *Database) Delete(key string) {
	db.lock.Lock()
	defer db.lock.Unlock()

	delete(db.data, key)
}

func (db *Database) Clear() {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.data = make(map[string]any)
}

func (db *Database) Persist(filename string) error {
	db.lock.RLock()
	defer db.lock.RUnlock()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if errClose := file.Close(); errClose != nil {
			panic(errClose)
		}
	}()

	encoder := gob.NewEncoder(file)
	if err = encoder.Encode(db.data); err != nil {
		return err
	}

	return nil
}

func (db *Database) Load(filename string) error {
	db.lock.RLock()
	defer db.lock.RUnlock()

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		if errClose := file.Close(); errClose != nil {
			panic(errClose)
		}
	}()

	decoder := gob.NewDecoder(file)
	if err = decoder.Decode(&db.data); err != nil {
		return err
	}

	return nil
}

func (db *Database) DeleteFile(filename string) error {
	//TODO: delete persist file

	return nil
}
