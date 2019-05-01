package session

import (
	"errors"
	"github.com/gofrs/uuid"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Session struct {
	UUID   string                 `json:"uuid"`
	Lock   sync.Mutex             `json:"-"`
	Data   map[string]interface{} `json:"data"`
	Expire time.Time              `json:"expire"`
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

// TODO: Setup mutexing here on the session itself

type Manager struct {
	SessionType string
	TokenName   string
	TokenType   string
	Provider    Provider
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
	Provider: &MemoryProvider{
		Store: map[string]*Session{},
	},
	Lifetime: DefaultSessionExpire,
}

func (m *Manager) Init() *Session {
	sess := m.Provider.Init(m.UUID())
	sess.SetExpire(m.Expiration())
	return sess
}

func (m *Manager) Load(sid string) (*Session, error) {
	sess, err := m.Provider.Read(sid)
	if err != nil {
		return &Session{}, errors.New(ErrorLoad)

	}

	return sess, nil
}

// Convenience method for pulling the session off the request header
func (m *Manager) RequestLoad(r *http.Request) (*Session, error) {
	// TODO: Implement type check and cookie loading

	header := r.Header.Get(m.TokenName)
	if header != "" {
		token := strings.TrimSpace(strings.Replace(header, HeaderPrefix, "", 1))
		return m.Load(token)
	}

	return &Session{}, errors.New(ErrorLoad)
}

func (m *Manager) Save(sess *Session) error {
	return m.Provider.Save(sess)
}

func (m *Manager) Extend(sess *Session) {
	sess.SetExpire(m.Expiration())
}

func (m *Manager) Destroy(sess *Session) error {
	return m.Provider.Destroy(sess.UUID)
}

func (m *Manager) UUID() string {
	id, _ := uuid.NewV4()
	return id.String()
}

func (m *Manager) Expiration() time.Time {
	return time.Now().Add(time.Duration(m.Lifetime) * time.Second)
}

func (m *Manager) GarbageCollect() {
	go m.Provider.GarbageCollect()
}
