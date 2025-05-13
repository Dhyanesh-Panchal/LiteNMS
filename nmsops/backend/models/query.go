package models

type UserQueryRequest struct {
	QueryId               uint64   `json:"query_id"`
	From                  uint32   `json:"from"`
	To                    uint32   `json:"to"`
	ObjectIds             []string `json:"object_ids"`
	CounterId             uint16   `json:"counter_id"`
	ObjectWiseAggregation string   `json:"object_wise_aggregation"`
	TimestampAggregation  string   `json:"timestamp_aggregation"`
	Interval              uint32   `json:"interval"`
}
