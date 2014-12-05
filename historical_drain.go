package logtap


import (
  "github.com/boltdb/bolt"
  "net/http"
  "strconv"
  "fmt"
)

// HistoricalDrain matches the drain interface
type HistoricalDrain struct {
  port int
  max  int
  db   *bolt.DB
}

// NewHistoricalDrain returns a new instance of a HistoricalDrain
func NewHistoricalDrain(port int, file string, max int) HistoricalDrain {
  db, err := bolt.Open(file, 0644, nil)
  if err != nil {
    db, err = bolt.Open("./bolt.db", 0644, nil)
  }
  return HistoricalDrain{
    port: port,
    max:  max,
    db: db,
  }
}

// Start starts the http listener.
// The listener on every request returns a json hash of logs of some arbitrary size
// default size is 100
func (h *HistoricalDrain) Start() {
  go func () {
    http.HandleFunc("/", h.handler)
    http.ListenAndServe(":"+strconv.Itoa(h.port), nil)
  }()
}

// handler handles any web request with any path and returns logs 
// this makes it so a client that talks to pagodabox's logvac
// can communicate with this system
func (h *HistoricalDrain) handler(w http.ResponseWriter, r *http.Request) {
  var limit int64
  if i, err := strconv.ParseInt(r.FormValue("limit"), 10, 64); err == nil {
    limit = i
  } else {
    limit = 10000
  }
  h.db.View(func(tx *bolt.Tx) error {
    // Create a new bucket.
    b := tx.Bucket([]byte("log"))
    c := b.Cursor()

    // move the curser along so we can start dropping logs 
    // in the right order at the right place
    if int64(b.Stats().KeyN) > limit {
      c.First()
      move_forward := int64(b.Stats().KeyN) - limit
      for i := int64(0); i < move_forward; i++ {
        c.Next()
      }
    } else {
      c.First()
    }

    for k, v := c.Next(); k != nil; k, v = c.Next() {
      fmt.Fprintf(w, "%s - %s", k, v)
    }

    return nil
  })    
    
}

// write drops data into a capped collection of logs
// if we hit the limit the last log item will be removed from the beginning
func (h HistoricalDrain) Write(msg Message) {
  fmt.Println(msg)
  h.db.Update(func(tx *bolt.Tx) error { 
    bucket, err := tx.CreateBucketIfNotExists([]byte("log"))
    if err != nil {
      fmt.Println(err)
      return err
    }
    err = bucket.Put([]byte(msg.Time.String()), []byte(msg.Content))
    if err != nil {
      fmt.Println(err)
      return err
    }

    if bucket.Stats().KeyN > h.max {
      delete_count := bucket.Stats().KeyN - h.max
      c := bucket.Cursor()
      for i := 0; i < delete_count; i++ {
        c.First()
        c.Delete()
      }
    }

    return nil
  })

}