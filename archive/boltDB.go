package archive

import (
	"binary"
	"bytes"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/pagodabox/golang-hatchet"
)

type (
	BoltArchive struct {
		db            *bolt.DB
		maxBucketSize uint32
	}
)

func (archive *BoltArchive) Slice(name string, offset, limit uint32, level int) ([]logtap.Messages, uint32, error) {
	var messages []logtap.Messages
	var nextIdx uint64
	err := archive.db.View(func(tx *bolt.Tx) error {
		messages = make([]logtap.Messages, 0)
		bucket := tx.Bucket([]byte(name))
		c := bucket.Cursor()
		k, _ := c.First()
		if k == nil {
			return
		}

		// I need to skip to the correct id
		c.Seek(offset)

		for k, v := c.First(); k != nil && limit > 0; k, v = c.Next() {
			msg := logtap.Message{}
			if err := json.Unmarshal(v, &msg); err != nil {
				return err
			}

			err = binary.Read(k, binary.BigEndian, &nextIdx)
			if err != nil {
				return err
			}

			if level < msg.Level {
				limit--
				append(messages, msg)
			}
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}
	return messages, nextIdx, nil
}

func (archive *BoltArchive) Write(log hatchet.Logger, msg logtap.Message) {
	err := archive.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(msg.Type))
		if err != nil {
			return err
		}

		value, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		// this needs to ensure lexographical order
		key := &bytes.Buffer{}
		nextLine := b.NextSequence()
		if err = binary.Write(key, binary.BigEndian, nextLine); err != nil {
			return err
		}

		if err = bucket.Put(key, value); err != nil {
			return err
		}

		// trim the bucket to size
		c := bucket.Cursor()
		c.First()
		for key_count := bucket.Stats().KeyN; key_count > archive.maxBucketSize; key_count-- {
			c.Delete()
			c.Next()
		}

		// I don't know how to do this other then scanning the collection periodically.
		// delete entries that are older then needed
		// c.First()
		// for {
		// 	c.Delete()
		// 	c.Next()
		// }

	})

	if err != nil {
		log.Error("[LOGTAP][Historical][write]" + err.Error())
	}
}
