package models

type HistogramQueryRequest struct {
    From      int64     `json:"from" binding:"required"`
    To        int64     `json:"to" binding:"required"`
    CounterID uint16    `json:"counterID" binding:"required"`
    ObjectIDs []uint32  `json:"objectIDs" binding:"required"`
}

type HistogramPoint struct {
    Timestamp int64   `json:"timestamp"`
    Value     float64 `json:"value"`
}

type HistogramResponse struct {
    Data map[int][]HistogramPoint `json:"data"`
} 