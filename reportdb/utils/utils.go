package utils

import (
	"path"
	"reportdb/config"
	"reportdb/global"
)

func GetStorageDir(date global.Date) string {

	return path.Join(config.ProjectRootPath, "storage", "data", date.Format())

}

func IntMin(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func UInt32Min(a, b uint32) uint32 {
	if a <= b {
		return a
	}
	return b
}

func UInt32Max(a, b uint32) uint32 {
	if a >= b {
		return a
	}
	return b
}
