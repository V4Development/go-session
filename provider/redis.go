package provider

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
	Client *redis.Client
	Config *RedisConfig
}

type RedisConfig struct {
	Server   string
	Password string
	Database int
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
