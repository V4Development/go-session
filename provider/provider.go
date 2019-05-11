package provider

const (
	ErrorSessionDoesNotExist = "ERR_SESSION_EXISTS"
)

// Provider Interface

type Provider interface {
	Read(sid string) (*Session, error)
	Save(session *Session) error
	Destroy(sid string) error
	GarbageCollect()
}
