package database

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/coreos/bbolt"
)

var (
	highestIDBucket = []byte("highest_id")
	cookiesBucket   = []byte("cookies")
)

type Database bolt.DB

func NewDatabase() (*Database, error) {
	db, err := bolt.Open("tumblr.db", 0644, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{
			highestIDBucket,
			cookiesBucket,
		} {
			_, err := tx.CreateBucketIfNotExists(name)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return (*Database)(db), nil
}

func (s *Database) Close() error {
	return s.get().Close()
}

func (s *Database) GetHighestID(blogName string) (int64, error) {
	var highestID int64

	err := s.get().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(highestIDBucket).Get([]byte(blogName))
		if b == nil {
			return nil
		}

		var err error
		highestID, err = strconv.ParseInt(string(b), 10, 64)
		return err
	})
	if err != nil {
		return 0, err
	}

	return highestID, nil
}

func (s *Database) SetHighestID(blogName string, highestID int64) error {
	return s.get().Update(func(tx *bolt.Tx) error {
		s := strconv.FormatInt(highestID, 10)
		return tx.Bucket(highestIDBucket).Put([]byte(blogName), []byte(s))
	})
}

var (
	WwwTumblrComCookiesKey = []byte("www.tumblr.com")
)

func (s *Database) GetCookies(key []byte) (cookies []*http.Cookie) {
	err := s.get().View(func(tx *bolt.Tx) error {
		data := tx.Bucket(cookiesBucket).Get(key)
		if len(data) == 0 {
			return nil
		}

		return json.Unmarshal(data, cookies)
	})
	if err != nil {
		log.Panic(err)
	}
	return
}

func (s *Database) SetCookies(key []byte, cookies []*http.Cookie) {
	err := s.get().View(func(tx *bolt.Tx) error {
		data, err := json.Marshal(cookies)
		if err != nil {
			return err
		}

		return tx.Bucket(cookiesBucket).Put(key, data)
	})
	if err != nil {
		log.Panic(err)
	}
	return
}

func (s *Database) get() *bolt.DB {
	return (*bolt.DB)(s)
}
