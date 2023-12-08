package smsproxy

import (
	"sync"
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
	// save given MessageID with ACCEPTED status. If given MessageID already exists, return an error
	return nil
}

func (r *inMemoryRepository) get(id MessageID) (MessageStatus, error) {
	// return status of given message, by it's MessageID. If not found, return NOT_FOUND status
	return "", nil
}

func (r *inMemoryRepository) update(id MessageID, newStatus MessageStatus) error {
	// Set new status for a given message.
	// If message is not in ACCEPTED state already - return an error.
	// If current status is FAILED or DELIVERED - don't update it and return an error. Those are final statuses and cannot be overwritten.
	return nil
}

func newRepository() repository {
	return &inMemoryRepository{db: make(map[MessageID]MessageStatus), lock: sync.RWMutex{}}
}
