package localkv

import (
	"github.com/tidwall/buntdb"
	"os"
	"path"
)

// LocalKV Structure to hold the db client
type LocalKV struct {
	db     *buntdb.DB
	dbPath string
}

// NewLocalKV constructor for the db client
func NewLocalKV(databasePath *string) (*LocalKV, error) {
	dbPath := path.Join(*databasePath, "kv.db")
	db, err := buntdb.Open(dbPath)
	if err != nil {
		return nil, err
	}

	return &LocalKV{
		db:     db,
		dbPath: dbPath,
	}, nil
}

// Close closes the db
func (l *LocalKV) Close() error {
	return l.db.Close()
}

// Get gets a value from the db
func (l *LocalKV) Get(key string) (string, error) {
	var val string

	err := l.db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(key)
		if err != nil {
			return err
		}

		val = v
		return nil
	})

	return val, err
}

// Set sets a value in the db
func (l *LocalKV) Set(key, value string) error {
	return l.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, value, nil)

		return err
	})
}

// RemoveDB removes db file
func (l *LocalKV) RemoveDB() error {
	return os.Remove(l.dbPath)
}
