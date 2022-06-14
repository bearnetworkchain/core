package cache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

var ErrorNotFound = errors.New("使用提供的鍵未找到任何值")

// 存儲是為了傳遞並由 New 函數使用（它提供命名空間和類型安全）
type Storage struct {
	storagePath string
}

// 緩存是一個命名空間和類型安全的鍵值存儲
type Cache[T any] struct {
	storage   Storage
	namespace string
}

// NewStorage sets up the storage needed for later cache usage
// path is the full path (including filename) to the database file to ues
// It does not need to be closed as this happens automatically in each call to the cache
func NewStorage(path string) (Storage, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return Storage{}, err
	}

	return Storage{path}, nil
}

// New creates a namespaced and typesafe key-value Cache
func New[T any](storage Storage, namespace string) Cache[T] {
	return Cache[T]{
		storage:   storage,
		namespace: namespace,
	}
}

// Key creates a single composite key from a list of keyParts
func Key(keyParts ...string) string {
	return strings.Join(keyParts, "")
}

// Clear deletes all namespaces and cached values
func (s Storage) Clear() error {
	db, err := openDb(s.storagePath)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			return tx.DeleteBucket(name)
		})
	})
}

// Put sets key to value within the namespace
// If the key already exists, it will be overwritten
func (c Cache[T]) Put(key string, value T) error {
	db, err := openDb(c.storage.storagePath)
	if err != nil {
		return err
	}
	defer db.Close()

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(value); err != nil {
		return err
	}
	result := buf.Bytes()

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(c.namespace))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), result)
	})
}

// Get fetches the value of key within the namespace.
// If no value exists, it will return found == false
func (c Cache[T]) Get(key string) (val T, err error) {
	db, err := openDb(c.storage.storagePath)
	if err != nil {
		return
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.namespace))
		if b == nil {
			return ErrorNotFound
		}
		c := b.Cursor()
		if k, v := c.Seek([]byte(key)); bytes.Equal(k, []byte(key)) {
			if v == nil {
				return ErrorNotFound
			}

			var decodedVal T
			d := gob.NewDecoder(bytes.NewReader(v))
			if err := d.Decode(&decodedVal); err != nil {
				return err
			}

			val = decodedVal
		} else {
			return ErrorNotFound
		}

		return nil
	})

	return val, err
}

// Delete removes a value for key within the namespace
func (c Cache[T]) Delete(key string) error {
	db, err := openDb(c.storage.storagePath)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.namespace))
		if b == nil {
			return nil
		}

		return b.Delete([]byte(key))
	})
}

func openDb(path string) (*bolt.DB, error) {
	return bolt.Open(path, 0640, &bolt.Options{Timeout: 1 * time.Minute})
}
