package levelup_test

import (
	"testing"

	"github.com/fiatjaf/levelup"
	"github.com/fiatjaf/levelup/stringlevelup"
	. "gopkg.in/check.v1"
)

var db stringlevelup.DB
var err error

func Test(updb levelup.DB, t *testing.T) {
	db = stringlevelup.StringDB(updb)
	TestingT(t)
}

type BasicSuite struct{}

var _ = Suite(&BasicSuite{})

// the tests must be run in order. yes.
// each suite depends on the the previous, that's why they have numbers in their names.

func (s *BasicSuite) Test1PutGetDel(c *C) {
	value, err := db.Get("key-x")
	c.Assert(err, DeepEquals, levelup.NotFound)
	c.Assert(value, DeepEquals, "")

	err = db.Put("key-x", "some value")
	c.Assert(err, IsNil)
	value, _ = db.Get("key-x")
	c.Assert(value, DeepEquals, "some value")

	err = db.Del("key-x")
	c.Assert(err, IsNil)
	value, err = db.Get("key-x")
	c.Assert(err, DeepEquals, levelup.NotFound)
	c.Assert(value, DeepEquals, "")
}

func (s *BasicSuite) Test2BatchPut(c *C) {
	somevalues := map[string]string{
		"letter:a": "a",
		"letter:b": "b",
		"letter:c": "c",
		"number:1": "1",
		"number:2": "2",
		"number:3": "3",
	}
	batch := []levelup.Operation{}
	for k, v := range somevalues {
		batch = append(batch, stringlevelup.Put(k, v))
	}
	err = db.Batch(batch)
	c.Assert(err, IsNil)

	iter := db.ReadRange(nil)
	retrieved := []string{}
	for ; iter.Valid(); iter.Next() {
		c.Assert(iter.Error(), IsNil)
		retrieved = append(retrieved, iter.Key(), iter.Value())
	}
	c.Assert(iter.Error(), IsNil)
	c.Assert(retrieved, DeepEquals, []string{
		"letter:a", "a",
		"letter:b", "b",
		"letter:c", "c",
		"number:1", "1",
		"number:2", "2",
		"number:3", "3",
	})
	iter.Release()
}

func (s *BasicSuite) Test3ReadRange(c *C) {
	// start-end
	iter := db.ReadRange(&stringlevelup.RangeOpts{
		Start: "letter:b",
		End:   "letter:~",
	})
	retrieved := []string{}
	for ; iter.Valid(); iter.Next() {
		c.Assert(iter.Error(), IsNil)
		retrieved = append(retrieved, iter.Key(), iter.Value())
	}
	c.Assert(iter.Error(), IsNil)
	c.Assert(retrieved, DeepEquals, []string{"letter:b", "b", "letter:c", "c"})
	iter.Release()

	// *-end
	iter = db.ReadRange(&stringlevelup.RangeOpts{
		End: "letter:c", /* non-inclusive */
	})
	retrieved = []string{}
	for ; iter.Valid(); iter.Next() {
		c.Assert(iter.Error(), IsNil)
		retrieved = append(retrieved, iter.Key(), iter.Value())
	}
	c.Assert(iter.Error(), IsNil)
	c.Assert(retrieved, DeepEquals, []string{"letter:a", "a", "letter:b", "b"})
	iter.Release()

	// start-* limit
	iter = db.ReadRange(&stringlevelup.RangeOpts{
		Start: "letter:c",
		Limit: 2,
	})
	retrieved = []string{}
	for ; iter.Valid(); iter.Next() {
		c.Assert(iter.Error(), IsNil)
		retrieved = append(retrieved, iter.Key(), iter.Value())
	}
	c.Assert(iter.Error(), IsNil)
	c.Assert(retrieved, DeepEquals, []string{"letter:c", "c", "number:1", "1"})
	iter.Release()

	// reverse
	iter = db.ReadRange(&stringlevelup.RangeOpts{
		Reverse: true,
	})
	retrieved = []string{}
	for ; iter.Valid(); iter.Next() {
		c.Assert(iter.Error(), IsNil)
		retrieved = append(retrieved, iter.Key(), iter.Value())
	}
	c.Assert(iter.Error(), IsNil)
	c.Assert(retrieved, DeepEquals, []string{
		"number:3", "3", "number:2", "2", "number:1", "1",
		"letter:c", "c", "letter:b", "b", "letter:a", "a",
	})
	iter.Release()

	// reverse start-end
	iter = db.ReadRange(&stringlevelup.RangeOpts{
		Start:   "letter:c",
		End:     "number:1~",
		Reverse: true,
	})
	retrieved = []string{}
	for ; iter.Valid(); iter.Next() {
		c.Assert(iter.Error(), IsNil)
		retrieved = append(retrieved, iter.Key(), iter.Value())
	}
	c.Assert(iter.Error(), IsNil)
	c.Assert(retrieved, DeepEquals, []string{"number:1", "1", "letter:c", "c"})
	iter.Release()

	// reverse *-end limit
	iter = db.ReadRange(&stringlevelup.RangeOpts{
		End:     "number:3", /* non-inclusive */
		Reverse: true,
		Limit:   3,
	})
	retrieved = []string{}
	for ; iter.Valid(); iter.Next() {
		c.Assert(iter.Error(), IsNil)
		retrieved = append(retrieved, iter.Key(), iter.Value())
	}
	c.Assert(iter.Error(), IsNil)
	c.Assert(retrieved, DeepEquals, []string{"number:2", "2", "number:1", "1", "letter:c", "c"})
	iter.Release()
}

func (s *BasicSuite) Test4MoreBatches(c *C) {
	batch := []levelup.Operation{
		stringlevelup.Del("number:2"),
		stringlevelup.Del("number:1"),
		stringlevelup.Put("number:3", "33"),
		stringlevelup.Del("number:4"),
		stringlevelup.Del("letter:a"),
		stringlevelup.Del("number:3"),
		stringlevelup.Del("letter:b"),
		stringlevelup.Del("letter:c"),
		stringlevelup.Put("number:3", "333"),
		stringlevelup.Del("letter:d"),
		stringlevelup.Put("letter:d", "dd"),
		stringlevelup.Del("letter:e"),
	}
	err = db.Batch(batch)
	c.Assert(err, IsNil)

	value, err := db.Get("number:1")
	c.Assert(err, DeepEquals, levelup.NotFound)
	c.Assert(value, DeepEquals, "")

	value, err = db.Get("letter:e")
	c.Assert(err, DeepEquals, levelup.NotFound)
	c.Assert(value, DeepEquals, "")

	value, _ = db.Get("number:3")
	c.Assert(value, DeepEquals, "333")

	value, _ = db.Get("letter:d")
	c.Assert(value, DeepEquals, "dd")
}
