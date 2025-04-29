package models

type UserQuery struct {
	QueryId               uint64   `json:"query_id"`
	From                  uint32   `json:"from"`
	To                    uint32   `json:"to"`
	ObjectIds             []string `json:"object_ids"`
	CounterId             uint16   `json:"counter_id"`
	VerticalAggregation   string   `json:"vertical_aggregation"`
	HorizontalAggregation string   `json:"horizontal_aggregation"`
	Interval              uint32   `json:"interval"`
}
