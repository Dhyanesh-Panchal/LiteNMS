package storage

import (
	"fmt"
	. "reportdb/config"
	"testing"
)

func TestNewStorage(t *testing.T) {

	_, err := NewStorage(ProjectRootPath+"/storage/data/2025/04/2/1/", 5, 120)

	if err != nil {
		t.Error(err)
	}
}

func TestStorage_Put(t *testing.T) {

	storage, err := NewStorage(ProjectRootPath+"/storage/data/2025/04/2/1/", 5, 120)

	if err != nil {
		t.Error(err)
	}

	data := []struct {
		o uint32
		d string
	}{
		{1, "Data11"},
		{1, "Data21"},
		{1, "Data31"},
		{1, "Data41"},
		{1, "Data51"},
		{1, "Data61"},
		{1, "Data12"},
		{2, "Data22"},
		{3, "Data32"},
		{4, "Data42"},
		{5, "Data52"},
		{6, "Data62"},
	}

	for _, d := range data {
		err = storage.Put(d.o, []byte(d.d))

	}

	if err != nil {
		t.Error(err)
	}

}

func TestStorage_Get(t *testing.T) {

	storage, err := NewStorage(ProjectRootPath+"/storage/data/2025/04/2/01/", 5, 120)

	if err != nil {

		t.Error(err)
	}

	data, err := storage.Get(4)

	if err != nil {

		t.Error(err)
	}

	fmt.Println(string(data))

}
