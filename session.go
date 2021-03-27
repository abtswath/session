package session

// Session interface
type Session interface {
	GetID() string

	SetID(id string)

	Start() error

	Save() error

	All() Value

	Exists(key string) bool

	Get(key string) (interface{}, error)

	Put(key string, value interface{})

	Remove(keys ...string)

	Flush()

	Regenerate()

	GetHandler() HandlerInterface
}

// HandlerInterface interface
type HandlerInterface interface {
	Open() error

	// Read session data
	Read(sessionID string) ([]byte, error)

	// Write session data
	Write(sessionID string, p []byte) error

	// Close the current session
	Close() error

	// Destroy a session
	Destroy(sessionID string) error

	// Clean up old sessions
	Gc() error
}
