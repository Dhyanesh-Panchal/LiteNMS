package db

import (
	"nms-backend/models"
)

type ReportDB interface {
	QueryHistogram(from, to int64, counterID uint16, objectIDs []uint32) (map[uint32][]models.HistogramPoint, error)
}
