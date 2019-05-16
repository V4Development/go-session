package session

import (
	"errors"
	"github.com/v4development/go-session/session"
	"net/http"
	"strings"
	"time"
)

// TODO: Setup mutexing here on the session itself

type Manager struct {
	KeyName  string
	KeyType  string
	Provider session.Provider
	Lifetime int64
}

// session_type
const (
	DefaultHeaderKey  = "Authorization"
	DefaultHeaderType = "Bearer"

	ErrorLoad = "session load error"
)

var DefaultManager = Manager{
	KeyName: DefaultHeaderKey,
	KeyType: DefaultHeaderType,
	Provider: &session.MemoryProvider{
		Store: map[string]*session.Session{},
	},
	Lifetime: session.DefaultSessionExpiration,
}

func (m *Manager) NewSession() *session.Session {
	sess := session.NewSession()
	sess.SetExpire(m.Expiration())
	return sess
}

func (m *Manager) NewSessionWithId(id string) *session.Session {
	sess := session.NewSessionWithId(id)
	sess.SetExpire(m.Expiration())
	return sess
}

func (m *Manager) Load(sid string) (*session.Session, error) {
	sess, err := m.Provider.Read(sid)
	if err != nil {
		return &session.Session{}, errors.New(ErrorLoad)
	}

	return sess, nil
}

// Convenience method for pulling the session off the request header
func (m *Manager) HeaderLoad(r *http.Request) (*session.Session, error) {
	// TODO: Implement type check and cookie loading

	header := r.Header.Get(m.KeyName)
	if header != "" {
		token := strings.TrimSpace(strings.Replace(header, m.KeyType, "", 1))
		return m.Load(token)
	}

	return &session.Session{}, errors.New(ErrorLoad)
}

func (m *Manager) CookieLoad(r *http.Request) (*session.Session, error) {
	// TODO: Implement cookie loading
	return nil, nil
}

func (m *Manager) Save(sess *session.Session) error {
	return m.Provider.Save(sess)
}

func (m *Manager) Extend(sess *session.Session) {
	sess.SetExpire(m.Expiration())
	_ = m.Save(sess)
}

func (m *Manager) Destroy(sess *session.Session) error {
	return m.Provider.Destroy(sess.UUID)
}

func (m *Manager) Expiration() time.Time {
	return time.Now().Add(time.Duration(m.Lifetime) * time.Second)
}

func (m *Manager) GarbageCollect() {
	m.Provider.GarbageCollect()
}
