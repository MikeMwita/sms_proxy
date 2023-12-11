package smsproxy

import (
	"errors"
	"sync"
)

const (
	ACCEPTED  MessageStatus = "ACCEPTED"
	FAILED    MessageStatus = "FAILED"
	DELIVERED MessageStatus = "DELIVERED"
	NOT_FOUND MessageStatus = "NOT_FOUND"
)

type repository interface {
	update(id MessageID, newStatus MessageStatus) error
	save(id MessageID) error
	get(id MessageID) (MessageStatus, error)
}

type inMemoryRepository struct {
	db   map[MessageID]MessageStatus
	lock sync.RWMutex
}

func (r *inMemoryRepository) save(id MessageID) error {
	// save given MessageID with ACCEPTED status. If given MessageID already exists, return an errorr.lock.Lock()
	r.lock.Lock()
	defer r.lock.Unlock()

	// Save given MessageID with ACCEPTED status. If given MessageID already exists, return an error.
	if _, exists := r.db[id]; exists {
		return errors.New("message ID already exists")
	}

	r.db[id] = ACCEPTED
	return nil
}

func (r *inMemoryRepository) get(id MessageID) (MessageStatus, error) {
	// return status of given message, by it's MessageID. If not found, return NOT_FOUND status
	r.lock.RLock()
	defer r.lock.RUnlock()

	status, exists := r.db[id]
	if !exists {
		return NOT_FOUND, nil
	}

	return status, nil

}

func (r *inMemoryRepository) update(id MessageID, newStatus MessageStatus) error {
	// Set new status for a given message.
	// If message is not in ACCEPTED state already - return an error.
	// If current status is FAILED or DELIVERED - don't update it and return an error. Those are final statuses and cannot be overwritten.

	r.lock.Lock()
	defer r.lock.Unlock()

	// Set a new status for a given message.
	// If the message is not in ACCEPTED state already - return an error.
	// If the current status is FAILED or DELIVERED - don't update it and return an error.
	// Those are final statuses and cannot be overwritten.
	currentStatus, exists := r.db[id]
	if !exists || currentStatus != ACCEPTED {
		return errors.New("invalid message status for update")
	}

	if currentStatus == FAILED || currentStatus == DELIVERED {
		return errors.New("cannot update final status")
	}

	r.db[id] = newStatus
	return nil
}

func newRepository() repository {
	return &inMemoryRepository{db: make(map[MessageID]MessageStatus), lock: sync.RWMutex{}}
}
