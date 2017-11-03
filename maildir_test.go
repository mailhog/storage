package storage

import (
	"testing"
)

func TestMaildirStore(t *testing.T) {
	storage := CreateMaildir("testdata")

	if storage.Count() != 0 {
		t.Errorf("storage.Count() expected: %d, got: %d", 0, storage.Count())
	}

	fillStorage(storage, 25)

	if storage.Count() != 25 {
		t.Errorf("storage.Count() expected: %d, got: %d", 25, storage.Count())
	}

	storage.DeleteAll()

	if storage.Count() != 0 {
		t.Errorf("storage.Count() expected: %d, got: %d", 0, storage.Count())
	}
}

func TestMaildirLoad(t *testing.T) {
	storage := CreateMaildir("testdata")

	storage.Store(newMessage("123"))

	item, err := storage.Load("123")
	if item == nil || err != nil {
		t.Errorf("storage.Load(\"123\") expected: not nil, got: nil")
	}

	unexisting, err := storage.Load("321")
	if unexisting != nil || err == nil {
		t.Errorf("storage.Load(\"321\") expected: nil, got: not nil")
	}
	storage.DeleteAll()
}

func TestMaildirList(t *testing.T) {
	storage := CreateMaildir("testdata")

	fillStorage(storage, 25)

	result, err := storage.List(0, 10)
	if len(*result) != 10 || err != nil {
		t.Errorf("len(result) expected: %d, got: %d", 10, len(*result))
	}

	result, err = storage.List(20, 30)
	if len(*result) != 5 || err != nil {
		t.Errorf("len(result) expected: %d, got: %d", 5, len(*result))
	}

	result, err = storage.List(20, 24)
	if len(*result) != 5 || err != nil {
		t.Errorf("len(result) expected: %d, got: %d", 5, len(*result))
	}

	result, err = storage.List(30, 40)
	if len(*result) != 0 || err != nil {
		t.Errorf("len(result) expected: %d, got: %d", 0, len(*result))
	}
	storage.DeleteAll()
}
