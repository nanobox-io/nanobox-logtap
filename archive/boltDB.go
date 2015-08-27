// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//
package archive

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/pagodabox/golang-hatchet"
	"github.com/pagodabox/nanobox-logtap"
)

type (
	BoltArchive struct {
		DB            *bolt.DB
		MaxBucketSize uint64
	}
)

func (archive *BoltArchive) Slice(name string, offset, limit uint64, level int) ([]logtap.Message, uint64, error) {
	var messages []logtap.Message
	var nextIdx uint64
	err := archive.DB.View(func(tx *bolt.Tx) error {
		messages = make([]logtap.Message, 0)
		fmt.Printf("opening the bucket %v\n", name)
		bucket := tx.Bucket([]byte(name))
		if bucket == nil {
			return nil
		}
		c := bucket.Cursor()
		k, _ := c.First()
		if k == nil {
			return nil
		}

		// I need to skip to the correct id
		initial := &bytes.Buffer{}
		if err := binary.Write(initial, binary.BigEndian, offset); err != nil {
			return err
		}

		c.Seek(initial.Bytes())

		for k, v := c.First(); k != nil && limit > 0; k, v = c.Next() {
			msg := logtap.Message{}
			if err := json.Unmarshal(v, &msg); err != nil {
				return err
			}

			err := binary.Read(bytes.NewBuffer(k), binary.BigEndian, &nextIdx)
			if err != nil {
				return err
			}

			if level < msg.Priority {
				limit--
				messages = append(messages, msg)
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
	err := archive.DB.Update(func(tx *bolt.Tx) error {
		fmt.Printf("processing the message %v\n", msg)
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
		nextLine, err := bucket.NextSequence()
		if err != nil {
			return err
		}

		if err = binary.Write(key, binary.BigEndian, nextLine); err != nil {
			return err
		}
		fmt.Printf("going to store the value %v\n", string(value))
		if err = bucket.Put(key.Bytes(), value); err != nil {
			return err
		}

		// trim the bucket to size
		c := bucket.Cursor()
		c.First()
		for key_count := uint64(bucket.Stats().KeyN); key_count > archive.MaxBucketSize; key_count-- {
			fmt.Printf("removing a value %v %v\n", key_count, archive.MaxBucketSize)
			c.Delete()
			c.Next()
		}

		fmt.Printf("bucket trimmed to size %v\n", bucket.Stats().KeyN)
		value = bucket.Get(key.Bytes())
		fmt.Printf("value that was stored %v\n", string(value))

		// I don't know how to do this other then scanning the collection periodically.
		// delete entries that are older then needed
		// c.First()
		// for {
		// 	c.Delete()
		// 	c.Next()
		// }

		return nil
	})

	if err != nil {
		fmt.Printf("%v\n", err)
		log.Error("[LOGTAP][Historical][write]" + err.Error())
	}
}
