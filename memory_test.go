package storage

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mailhog/data"
)

func TestStore(t *testing.T) {
	storage := CreateInMemory()

	if storage.Count() != 0 {
		t.Errorf("storage.Count() expected: %d, got: %d", 0, storage.Count())
	}

	var wg sync.WaitGroup
	wg.Add(25)
	for i := 0; i < 25; i++ {
		go func(i int) {
			msg := &data.Message{
				ID:      data.MessageID(i),
				Created: time.Now(),
			}
			storage.Store(msg)
			wg.Done()
		}(i)
	}
	wg.Wait()

	if storage.Count() != 25 {
		t.Errorf("storage.Count() expected: %d, got: %d", 25, storage.Count())
	}
}

func TestDeleteAll(t *testing.T) {
	storage := CreateInMemory()

	if storage.Count() != 0 {
		t.Errorf("storage.Count() expected: %d, got: %d", 0, storage.Count())
	}

	for i := 0; i < 25; i++ {
		storage.Store(&data.Message{ID: data.MessageID(i), Created: time.Now()})
	}

	if storage.Count() != 25 {
		t.Errorf("storage.Count() expected: %d, got: %d", 25, storage.Count())
	}

	storage.DeleteAll()

	if storage.Count() != 0 {
		t.Errorf("storage.Count() expected: %d, got: %d", 0, storage.Count())
	}
}

func TestDeleteOne(t *testing.T) {
	storage := CreateInMemory()

	if storage.Count() != 0 {
		t.Errorf("storage.Count() expected: %d, got: %d", 0, storage.Count())
	}

	for i := 0; i < 25; i++ {
		storage.Store(&data.Message{ID: data.MessageID(fmt.Sprintf("%d", i)), Created: time.Now()})
	}

	storage.DeleteOne("1")

	if storage.Count() != 24 {
		t.Errorf("storage.Count() expected: %d, got: %d", 24, storage.Count())
	}

	storage.DeleteOne("34789")

	if storage.Count() != 24 {
		t.Errorf("storage.Count() expected: %d, got: %d", 24, storage.Count())
	}
}
