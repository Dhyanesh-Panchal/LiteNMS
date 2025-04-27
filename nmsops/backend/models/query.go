package models

type DataPoint struct {
	Timestamp uint32      `json:"timestamp"`
	Value     interface{} `json:"value"`
}
