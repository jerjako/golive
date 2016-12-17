package data

import (
	"encoding/json"
	"fmt"

	"github.com/alexjomin/golive/model"
	"github.com/boltdb/bolt"
)

type boltSource struct {
	db     *bolt.DB
	bucket []byte
}

// NewBoltSource returns an bolt based data source
func NewBoltSource(path string) (Source, error) {

	db, err := bolt.Open(path, 0600, nil)

	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("deliveries"))

		if err != nil {
			if err != bolt.ErrBucketExists {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &boltSource{
		db:     db,
		bucket: []byte("deliveries"),
	}, nil

}

func (s *boltSource) Get(key string) (interface{}, error) {

	var v []byte

	_ = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)

		if b == nil {
			fmt.Println("bucket is nil")
			return nil
		}
		v = b.Get([]byte(key))
		return nil
	})

	if v == nil {
		return nil, nil
	}

	return v, nil

}

func (s *boltSource) Delete(key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucket))
		return b.Delete([]byte(key))
	})
}

func (s *boltSource) GetAll() ([]model.Delivery, error) {

	p := []model.Delivery{}

	err := s.db.View(func(tx *bolt.Tx) error {

		c := tx.Bucket([]byte(s.bucket)).Cursor()

		i := 0

		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			var d model.Delivery
			json.Unmarshal(v, &d)
			p = append(p, d)
			i++
		}

		return nil

	})

	return p, err

}

func (s *boltSource) Insert(key string, value interface{}) error {

	buf, err := json.Marshal(value)

	if err != nil {
		return err
	}

	_ = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		err := b.Put([]byte(key), buf)
		return err
	})

	return nil
}
