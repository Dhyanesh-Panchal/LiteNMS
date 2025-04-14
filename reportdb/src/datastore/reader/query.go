package reader

import . "datastore/containers"

type Query struct {
	QueryId     uint64   `json:"query_id"`
	From        uint32   `json:"from"`
	To          uint32   `json:"to"`
	ObjectIds   []uint32 `json:"object_ids"`
	CounterId   uint16   `json:"counter_id"`
	Aggregation string   `json:"aggregation"`
}

type Result struct {
	QueryId uint64 `json:"query_id"`

	Data map[uint32][]DataPoint `json:"data"`
}
