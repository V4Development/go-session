package session

import (
	"errors"
	"github.com/v4development/go-session/provider"
	"net/http"
	"strings"
	"time"
)

// TODO: Setup mutexing here on the session itself

type Manager struct {
	SessionType string
	TokenName   string
	TokenType   string
	Provider    provider.Provider
	Lifetime    int64
}

// session_type
const (
	HeaderKey    = "Authorization"
	HeaderPrefix = "Bearer"

	TypeHeader = "Header"
	TypeCookie = "Cookie"

	ErrorLoad = "session load error"
)

var DefaultManager = Manager{
	SessionType: TypeHeader,
	TokenName:   HeaderKey,
	TokenType:   HeaderPrefix,
	Provider: &provider.MemoryProvider{
		Store: map[string]*provider.Session{},
	},
	Lifetime: provider.DefaultSessionExpiration,
}

func (m *Manager) NewSession() *provider.Session {
	sess := provider.NewSession()
	sess.SetExpire(m.Expiration())
	return sess
}

func (m *Manager) NewSessionWithId(id string) *provider.Session {
	sess := provider.NewSessionWithId(id)
	sess.SetExpire(m.Expiration())
	return sess
}

func (m *Manager) Load(sid string) (*provider.Session, error) {
	sess, err := m.Provider.Read(sid)
	if err != nil {
		return &provider.Session{}, errors.New(ErrorLoad)
	}

	return sess, nil
}

// Convenience method for pulling the session off the request header
func (m *Manager) RequestLoad(r *http.Request) (*provider.Session, error) {
	// TODO: Implement type check and cookie loading

	header := r.Header.Get(m.TokenName)
	if header != "" {
		token := strings.TrimSpace(strings.Replace(header, HeaderPrefix, "", 1))
		return m.Load(token)
	}

	return &provider.Session{}, errors.New(ErrorLoad)
}

func (m *Manager) Save(sess *provider.Session) error {
	return m.Provider.Save(sess)
}

func (m *Manager) Extend(sess *provider.Session) {
	sess.SetExpire(m.Expiration())
	_ = m.Save(sess)
}

func (m *Manager) Destroy(sess *provider.Session) error {
	return m.Provider.Destroy(sess.UUID)
}

func (m *Manager) Expiration() time.Time {
	return time.Now().Add(time.Duration(m.Lifetime) * time.Second)
}

func (m *Manager) GarbageCollect() {
	go m.Provider.GarbageCollect()
}
