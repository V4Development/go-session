package provider

import (
	"errors"
	"sync"
	"time"
)

type MemoryProvider struct {
	Store map[string]*Session
}

func NewMemoryProvider() *MemoryProvider {
	return &MemoryProvider{
		make(map[string]*Session),
	}
}

func (p *MemoryProvider) Read(sid string) (*Session, error) {
	sess, exists := p.Store[sid]

	var err error
	if !exists {
		err = errors.New(ErrorSessionDoesNotExist)
	}

	sess.Lock = sync.Mutex{}

	return sess, err
}

func (p *MemoryProvider) Save(sess *Session) error {
	p.Store[sess.UUID] = sess
	return nil
}

func (p *MemoryProvider) Destroy(sid string) error {
	delete(p.Store, sid)
	return nil
}

func (p *MemoryProvider) GarbageCollect() {
	now := time.Now()
	for sid, sess := range p.Store {
		if sess.Expire.Before(now) {
			delete(p.Store, sid)
		}
	}
}