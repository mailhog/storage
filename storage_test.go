package storage

import (
	"strconv"
	"sync"
	"time"

	"github.com/mailhog/data"
)

func fillStorage(storage Storage, itemsCount int) {
	var wg sync.WaitGroup
	wg.Add(itemsCount)
	for i := 0; i < itemsCount; i++ {
		go func(i int) {
			msg := newMessage(strconv.Itoa(i))
			storage.Store(msg)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func newMessage(messageID string) *data.Message {
	return &data.Message{
		ID:      data.MessageID(messageID),
		Created: time.Now(),
		Raw: &data.SMTPMessage{
			From: "from@email.com",
			To:   []string{"to@email.com"},
			Data: "some data string",
			Helo: "helo string",
		},
	}
}
