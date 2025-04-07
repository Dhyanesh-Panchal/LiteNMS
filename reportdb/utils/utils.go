package utils

import (
	"path"
	"reportdb/config"
)

func GetStorageDir(date string) string {

	return path.Join(config.ProjectRootPath, "storage", "data", date)

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
