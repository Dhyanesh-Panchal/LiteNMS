package storage

import (
	. "reportdb/global"
	. "reportdb/storage/containers"
	"testing"
	"time"
)

func TestEngine_Put(t *testing.T) {
	engine := NewEngine()

	key := Key{
		1, 1, GetDate(time.Now()),
	}

	values := []DataPoint{
		{
			1704067215,
			703,
		},
		{
			1704067216,
			704,
		},
		{
			1704067220,
			705,
		},
		{
			1704067225,
			704,
		},
		{
			1704067230,
			705,
		},
		{
			1704067235,
			705,
		},
		{
			1704067240,
			705,
		},
		{
			1704067245,
			705,
		},
		{
			1704067250,
			705,
		},
		{
			1704067255,
			705,
		},
	}

	engine.Put(key, values)
}
