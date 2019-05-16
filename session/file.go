package session

type FileProvider struct {
	Path string
}

func (p *FileProvider) Read(sid string) (*Session, error) {
	return &Session{}, nil
}

func (p *FileProvider) Save(session *Session) error {
	return nil
}

func (p *FileProvider) Destroy(sid string) error {
	return nil
}

func (p *FileProvider) GarbageCollect() {

}
