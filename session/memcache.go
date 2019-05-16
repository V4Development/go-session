package session

type MemcacheProvider struct {
	Servers []string
}

func (p *MemcacheProvider) Read(sid string) (*Session, error) {
	return &Session{}, nil
}

func (p *MemcacheProvider) Save(session *Session) error {
	return nil
}

func (p *MemcacheProvider) Destroy(sid string) error {
	return nil
}

func (p *MemcacheProvider) GarbageCollect() {

}
