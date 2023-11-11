package storage_test

import (
	"log"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yakimka/goshort/storage"
)

func TestSQLiteURLStorage(t *testing.T) {
	t.Run("can't get not existing url", func(t *testing.T) {
		store := createInitedSQLiteURLStorage()

		_, err := store.Get("unknown")

		if err == nil {
			t.Error("Expected error, but got nil")
		}
	})
	t.Run("can set and get url", func(t *testing.T) {
		store := createInitedSQLiteURLStorage()

		store.Set("test", "http://example.com/test")
		url, err := store.Get("test")

		if err != nil {
			t.Error("Expected nil, but got", err)
		}
		if url != "http://example.com/test" {
			t.Error("Expected http://example.com/test, but got", url)
		}
	})
	t.Run("can't set url with existing id", func(t *testing.T) {
		store := createInitedSQLiteURLStorage()
		store.Set("test", "http://example.com/test")

		err := store.Set("test", "http://example.com/try_to_replace")

		if err == nil {
			t.Error("Expected error, but got nil")
		}
		if err != storage.ErrUniqueConstraint {
			t.Error("Expected ErrUniqueConstraint, but got", err)
		}
	})
}

func createInitedSQLiteURLStorage() *storage.SQLiteURLStorage {
	storage, err := storage.NewSQLiteURLStorage(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	storage.Init()
	return storage
}
