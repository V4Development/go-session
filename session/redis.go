package session

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"sync"
	"time"
)

// Redis Session Provider

const DefaultRedisDatabase = 0

type RedisProvider struct {
	// Servers []string
	Client  *redis.Client
	Options *redis.Options
}

func NewRedisProvider(server, password string, database int) (*RedisProvider, error) {
	o := &redis.Options{
		Addr:     server,
		Password: password,
		DB:       database,
	}

	p := &RedisProvider{
		Options: o,
	}

	p.RedisInit()

	return p, nil
}

func (p *RedisProvider) Read(sid string) (*Session, error) {
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
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	exp := p.CalcExpiration(session)
	p.Client.Set(session.UUID, string(data), exp)
	return nil
}

func (p *RedisProvider) Destroy(sid string) error {
	p.Client.Del(sid)
	return nil
}

func (p *RedisProvider) GarbageCollect() {
	// Not needed with entries that have an expiration
}

func (p *RedisProvider) RedisInit() {
	if p.Client == nil {
		p.Client = redis.NewClient(p.Options)
	}
}

func (p *RedisProvider) CalcExpiration(s *Session) time.Duration {
	return s.Expire.Sub(time.Now())
}
