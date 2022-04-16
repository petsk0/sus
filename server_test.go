package sus

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func TestHandleGet(t *testing.T) {
	defer os.Remove("test.db")
	db, err := bolt.Open("test.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("urls"))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewShortener("test", 5, db)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	s.HandleGet(w, r)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("wrong http status code: got %d but wanted %d", w.Code, http.StatusOK)
	}
}

func TestHandlePost(t *testing.T) {
	defer os.Remove("test.db")
	db, err := bolt.Open("test.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("urls"))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewShortener("test", 5, db)

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	s.HandlePost(w, r)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("wrong http status code: got %d but wanted %d", w.Code, http.StatusOK)
	}
}

func TestHandleRedirect(t *testing.T) {
	defer os.Remove("test.db")
	db, err := bolt.Open("test.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("urls"))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	key := "aaaaa"
	val := "https://example.com"
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		err := b.Put([]byte(key), []byte(val))
		return err
	})

	s := NewShortener("test", 5, db)

	r := httptest.NewRequest(http.MethodGet, "/aaaaa", nil)
	w := httptest.NewRecorder()
	s.HandleRedirect(w, r)

	if status := w.Code; status != http.StatusFound {
		t.Errorf("wrong http status code: got %d but wanted %d", w.Code, http.StatusFound)
	}

	r = httptest.NewRequest(http.MethodGet, "/a23B5", nil)
	w = httptest.NewRecorder()
	s.HandleRedirect(w, r)

	if status := w.Code; status != http.StatusNotFound {
		t.Errorf("wrong http status code: got %d but wanted %d", w.Code, http.StatusNotFound)
	}
}
