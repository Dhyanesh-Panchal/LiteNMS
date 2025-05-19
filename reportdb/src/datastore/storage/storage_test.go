package storage

import (
	. "datastore/utils"
	"fmt"
	"testing"
)

func TestNewStorage(t *testing.T) {

	_, err := NewStorage(StorageDirectory+"/2025/4/2/1/", 5, 120, true)

	if err != nil {
		t.Error(err)
	}
}

func TestStorage_Put(t *testing.T) {

	storage, err := NewStorage(StorageDirectory+"/2025/4/2/1/", 5, 120, true)

	if err != nil {
		t.Error(err)
	}

	data := []struct {
		o uint32
		d string
	}{
		{1, "Data11"},
		{1, "Data12"},
		{1, "Data13"},
		{1, "Data14"},
		{1, "Data15"},
		{1, "Data16"},
		{1, "Data17"},
		{1, "this is some big data from object1"},
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

	storage, err := NewStorage(StorageDirectory+"/2025/4/2/1/", 5, 120, false)

	if err != nil {

		t.Error(err)
	}

	data, err := storage.Get(1)

	if err != nil {

		t.Error(err)
	}

	fmt.Println(string(data))

}
