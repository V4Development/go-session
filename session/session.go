package session

import (
	"github.com/gofrs/uuid"
	"sync"
	"time"
)

const (
	DefaultSessionExpiration = 1200
)

type Session struct {
	UUID   string                 `json:"uuid"`
	Lock   sync.Mutex             `json:"-"`
	Data   map[string]interface{} `json:"data"`
	Expire time.Time              `json:"expire"`
}

func NewSession() *Session {
	id, _ := uuid.NewV4()
	return NewSessionWithId(id.String())
}

func NewSessionWithId(sid string) *Session {
	return &Session{
		UUID:   sid,
		Lock:   sync.Mutex{},
		Data:   make(map[string]interface{}),
		Expire: time.Now().Add(DefaultSessionExpiration * time.Second),
	}
}

func (sess *Session) Set(k string, v interface{}) {
	sess.Lock.Lock()
	defer sess.Lock.Unlock()

	sess.Data[k] = v
}

func (sess *Session) SetData(m map[string]interface{}) {
	sess.Lock.Lock()
	defer sess.Lock.Unlock()

	for k, v := range m {
		sess.Data[k] = v
	}
}

func (sess *Session) SetExpire(exp time.Time) {
	sess.Lock.Lock()
	defer sess.Lock.Unlock()

	sess.Expire = exp
}
