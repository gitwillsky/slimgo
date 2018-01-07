package session

const sessionIDLength = 64
const sessionExpireTime = 3600
const gcTimeInterval = 30

// Session is an interface for session instance, a session instance contains
// data needed.
type Session interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Delete(key string) error
	SessionID() string
	isExpired() bool
}

// Manager manages a map of sessions.
type Manager interface {
	sessionID() (string, error)
	NewSession() (Session, error)
	Session(string) (Session, error)
	GarbageCollection()
	Count() int
}
