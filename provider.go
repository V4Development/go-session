package session

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"
)

const (
	DefaultSessionExpire     = 1200
	ErrorSessionDoesNotExist = "ERR_SESSION_EXISTS"
)

// Session Provider Interface

type Provider interface {
	Init(sid string) *Session
	Read(sid string) (*Session, error)
	Save(session *Session) error
	Destroy(sid string) error
	GarbageCollect()
}

// Memory Session Provider

type MemoryProvider struct {
	Provider
	Store map[string]*Session
}

func (p *MemoryProvider) Init(sid string) *Session {
	return &Session{
		UUID:   sid,
		Lock:   sync.Mutex{},
		Data:   make(map[string]interface{}),
		Expire: time.Now().Add(DefaultSessionExpire * time.Second),
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

func (p *MemoryProvider) Save(session *Session) error {
	p.Store[session.UUID] = session
	return nil
}

func (p *MemoryProvider) Destroy(sid string) error {
	delete(p.Store, sid)
	return nil
}

func (p *MemoryProvider) GarbageCollect() {

}

// Redis Session Provider

const DefaultRedisDatabase = 0

type RedisProvider struct {
	Provider
	// Servers []string
	Client *redis.Client
	Config *RedisConfig
}

type RedisConfig struct {
	Server   string
	Password string
	Database int
}

func (p *RedisProvider) Init(sid string) *Session {
	s := &Session{
		UUID:   sid,
		Lock:   sync.Mutex{},
		Data:   make(map[string]interface{}),
		Expire: time.Now().Add(DefaultSessionExpire * time.Second),
	}

	return s
}

func (p *RedisProvider) Read(sid string) (*Session, error) {
	// TODO: Remove from here....should be moved out to initialization
	p.RedisInit()

	data, err := p.Client.Get(sid).Result()
	if err != nil {
		fmt.Println("RedisProvider - Get Data: ", err)
		return &Session{}, err
	}

	sess := &Session{}
	err = json.Unmarshal([]byte(data), &sess)
	if err != nil {
		fmt.Println("RedisProvider - Unmarshal: ", err)
		return &Session{}, err
	}

	sess.Lock = sync.Mutex{}

	return sess, nil
}

func (p *RedisProvider) Save(session *Session) error {
	// TODO: Remove from here....should be moved out to initialization
	p.RedisInit()

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	exp := p.CalcExpiration(session)
	p.Client.Set(session.UUID, string(data), exp)
	return nil
}

func (p *RedisProvider) Destroy(sid string) error {
	// TODO: Remove from here....should be moved out to initialization
	p.RedisInit()
	p.Client.Del(sid)
	return nil
}

func (p *RedisProvider) GarbageCollect() {

}

func (p *RedisProvider) RedisInit() {
	if p.Client == nil {
		fmt.Println("Redis Init")
		p.Client = redis.NewClient(&redis.Options{
			Addr:     p.Config.Server,
			Password: p.Config.Password,
			DB:       p.Config.Database,
		})
	}
}

func (p *RedisProvider) CalcExpiration(s *Session) time.Duration {
	return s.Expire.Sub(time.Now())
}

// Memcache Session Provider

type MemcacheProvider struct {
	Provider
	Servers []string
}

func (p *MemcacheProvider) Init(sid string) Session {
	return Session{
		UUID:   sid,
		Lock:   sync.Mutex{},
		Data:   make(map[string]interface{}),
		Expire: time.Now().Add(DefaultSessionExpire * time.Second),
	}
}

func (p *MemcacheProvider) Read(sid string) (Session, error) {
	return Session{}, nil
}

func (p *MemcacheProvider) Save(session Session) error {
	return nil
}

func (p *MemcacheProvider) Destroy(sid string) error {
	return nil
}

func (p *MemcacheProvider) GarbageCollect() {

}

// TODO: MySQL Based Sessions

const DefaultMySQLTableName = "session"

type MySQLProvider struct {
	Provider
	*sql.DB
	Table    string
	SysCheck bool
}

func (p *MySQLProvider) Init(sid string) *Session {
	return &Session{
		UUID:   sid,
		Lock:   sync.Mutex{},
		Data:   make(map[string]interface{}),
		Expire: time.Now().Add(DefaultSessionExpire * time.Second),
	}
}

func (p *MySQLProvider) Read(sid string) (*Session, error) {
	var sess Session
	var d []byte
	q := "SELECT * FROM " + p.Table + " WHERE uuid=?"
	row := p.QueryRow(q, sid)
	if err := row.Scan(&sess.UUID, &d, &sess.Expire); err != nil {
		return nil, err
	}

	err := json.Unmarshal(d, &sess.Data)
	if err != nil {
		return nil, err
	}

	sess.Lock = sync.Mutex{}

	return &sess, nil
}

func (p *MySQLProvider) Save(session *Session) error {
	q := "INSERT INTO " + p.Table + " SET uuid=?, data=?, expire=? " +
		"ON DUPLICATE KEY UPDATE data=?, expire=?"
	stmt, err := p.Prepare(q)
	if err != nil {
		return err
	}

	data, err := json.Marshal(session.Data)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(session.UUID, data, session.Expire, data, session.Expire)
	if err != nil {
		return err
	}

	return nil
}

func (p *MySQLProvider) Destroy(sid string) error {
	q := "DELETE FROM " + p.Table + " WHERE uuid=?"
	stmt, err := p.Prepare(q)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(sid)
	if err != nil {
		return err
	}

	return nil
}

func (p *MySQLProvider) GarbageCollect() {

}

func (p *MySQLProvider) MySQLInit() error {
	if !p.SysCheck {
		var exists bool
		q := "SELECT 1 FROM " + p.Table + " LIMIT 1"
		row := p.QueryRow(q)
		if err := row.Scan(&exists); err != nil {
			// create the table

			q = "CREATE TABLE " + p.Table + " (" +
				"uuid varchar(36) not null," +
				"data blob null," +
				"expire datetime default CURRENT_TIMESTAMP not null," +
				"constraint " + p.Table + "_pk " +
				"primary key (uuid))"

			_, err := p.Exec(q)
			if err != nil {
				p.SysCheck = false
				return err
			}

			p.SysCheck = true
		} else {
			p.SysCheck = true
		}
	}

	return nil
}

// TODO: File Based Sessions
