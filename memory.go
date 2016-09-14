package storage

import (
	"strings"

	"github.com/mailhog/data"
)

// InMemory is an in memory storage backend
type InMemory struct {
	MessageIDIndex map[string]int
	Messages       []*data.Message

	sliceIndexMap map[int]string
	messageLimit  int
	writeIndex    int
}

// CreateInMemory creates a new in memory storage backend
func CreateInMemory() *InMemory {
	return &InMemory{
		MessageIDIndex: make(map[string]int),
		Messages:       make([]*data.Message, 0),
		sliceIndexMap:  make(map[int]string),
	}
}

// Set a limit on the number of messages that will be stored.
// Once this limit is reached, the oldest message will be overwritten
// by every new message.
func (memory *InMemory) SetMessageLimit(limit int) error {
	memory.messageLimit = limit
	return nil
}

// Store stores a message and returns its storage ID
func (memory *InMemory) Store(m *data.Message) (string, error) {
	if memory.messageLimit > 0 && len(memory.Messages) == memory.messageLimit {
		// overwrite oldest message
		memory.Messages[memory.writeIndex] = m
		messageID := memory.sliceIndexMap[memory.writeIndex]
		delete(memory.MessageIDIndex, messageID)
		memory.MessageIDIndex[string(m.ID)] = memory.writeIndex
		memory.sliceIndexMap[memory.writeIndex] = string(m.ID)
		memory.writeIndex++
		// if we reach the end of the slice then start over
		if memory.writeIndex == len(memory.Messages) {
			memory.writeIndex = 0
		}
	} else {
		memory.Messages = append(memory.Messages, m)
		memory.MessageIDIndex[string(m.ID)] = len(memory.Messages) - 1
		memory.sliceIndexMap[len(memory.Messages)-1] = string(m.ID)
	}
	return string(m.ID), nil
}

// Count returns the number of stored messages
func (memory *InMemory) Count() int {
	return len(memory.Messages)
}

// Search finds messages matching the query
func (memory *InMemory) Search(kind, query string, start, limit int) (*data.Messages, int, error) {
	// FIXME needs optimising, or replacing with a proper db!
	query = strings.ToLower(query)
	var filteredMessages = make([]*data.Message, 0)
	for _, m := range memory.Messages {
		doAppend := false

		switch kind {
		case "to":
			for _, to := range m.To {
				if strings.Contains(strings.ToLower(to.Mailbox+"@"+to.Domain), query) {
					doAppend = true
					break
				}
			}
			if !doAppend {
				if hdr, ok := m.Content.Headers["To"]; ok {
					for _, to := range hdr {
						if strings.Contains(strings.ToLower(to), query) {
							doAppend = true
							break
						}
					}
				}
			}
		case "from":
			if strings.Contains(strings.ToLower(m.From.Mailbox+"@"+m.From.Domain), query) {
				doAppend = true
			}
			if !doAppend {
				if hdr, ok := m.Content.Headers["From"]; ok {
					for _, from := range hdr {
						if strings.Contains(strings.ToLower(from), query) {
							doAppend = true
							break
						}
					}
				}
			}
		case "containing":
			if strings.Contains(strings.ToLower(m.Content.Body), query) {
				doAppend = true
			}
			if !doAppend {
				for _, hdr := range m.Content.Headers {
					for _, v := range hdr {
						if strings.Contains(strings.ToLower(v), query) {
							doAppend = true
						}
					}
				}
			}
		}

		if doAppend {
			filteredMessages = append(filteredMessages, m)
		}
	}

	var messages = make([]data.Message, 0)

	if len(filteredMessages) == 0 || start > len(filteredMessages) {
		msgs := data.Messages(messages)
		return &msgs, 0, nil
	}

	if start+limit > len(filteredMessages) {
		limit = len(filteredMessages) - start
	}

	start = len(filteredMessages) - start - 1
	end := start - limit

	if start < 0 {
		start = 0
	}
	if end < -1 {
		end = -1
	}

	for i := start; i > end; i-- {
		//for _, m := range memory.MessageIndex[start:end] {
		messages = append(messages, *filteredMessages[i])
	}

	msgs := data.Messages(messages)
	return &msgs, len(filteredMessages), nil
}

// List lists stored messages by index
func (memory *InMemory) List(start int, limit int) (*data.Messages, error) {
	var messages = make([]data.Message, 0)

	start = (memory.writeIndex - 1) - start
	if start < 0 {
		start = len(memory.Messages) + start
	}

	if len(memory.Messages) == 0 || start > len(memory.Messages) {
		msgs := data.Messages(messages)
		return &msgs, nil
	}

	if limit > len(memory.Messages) {
		limit = len(memory.Messages)
	}

	i := start
	for {
		if len(messages) == limit {
			break
		}
		messages = append(messages, *memory.Messages[i])
		i--
		if i < 0 {
			i = len(memory.Messages) - 1
		}
		// we've gone full circle
		if i == memory.writeIndex {
			break
		}
	}

	msgs := data.Messages(messages)
	return &msgs, nil
}

// DeleteOne deletes an individual message by storage ID
func (memory *InMemory) DeleteOne(id string) error {
	index := memory.MessageIDIndex[id]
	delete(memory.MessageIDIndex, id)
	for k, v := range memory.MessageIDIndex {
		if v > index {
			memory.MessageIDIndex[k] = v - 1
		}
	}
	memory.Messages = append(memory.Messages[:index], memory.Messages[index+1:]...)
	return nil
}

// DeleteAll deletes all in memory messages
func (memory *InMemory) DeleteAll() error {
	memory.Messages = make([]*data.Message, 0)
	memory.MessageIDIndex = make(map[string]int)
	return nil
}

// Load returns an individual message by storage ID
func (memory *InMemory) Load(id string) (*data.Message, error) {
	if idx, ok := memory.MessageIDIndex[id]; ok {
		return memory.Messages[idx], nil
	}
	return nil, nil
}
