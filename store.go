package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"
)

// Value session data
type Value map[string]interface{}

func New(r *http.Request, rw http.ResponseWriter) *Store {
	store := &Store{
		Handler: &FileSessionHandler{
			Path:        "/tmp/go_session",
			MaxLifeTime: time.Hour * 24 * 7,
		},
		CookieOptions: &CookieOptions{
			Name:     "go_sessid",
			Path:     "/",
			Domain:   "",
			MaxAge:   time.Hour * 24 * 7,
			Secure:   true,
			HttpOnly: true,
		},
		Request:        r,
		ResponseWriter: rw,
	}
	return store
}

type CookieOptions struct {

	// http.Cookie.Name
	Name string

	// http.Cookie.Path
	Path string

	// http.Cookie.Domain
	Domain string

	// http.Cookie.Duration
	MaxAge time.Duration

	// http.Cookie.Secure
	Secure bool

	// http.Cookie.HttpOnly
	HttpOnly bool

	// http.Cookie.SameSite
	SameSite http.SameSite

	// Raw http.Cookie.Raw
	Raw string

	// TODO. encrypt cookie
	Encrypt bool
}

func (c *CookieOptions) newCookie(value string) *http.Cookie {
	return &http.Cookie{
		Name:     c.Name,
		Value:    value,
		HttpOnly: c.HttpOnly,
		Domain:   c.Domain,
		Path:     c.Path,
		Expires:  time.Now().Add(c.MaxAge),
		MaxAge:   int(c.MaxAge),
		Secure:   c.Secure,
		Raw:      c.Raw,
		SameSite: c.SameSite,
	}
}

// Store session store
type Store struct {
	// session id
	id string

	// session handler
	Handler HandlerInterface

	// cookie options
	CookieOptions *CookieOptions

	// session data
	attributes Value

	// http request
	Request *http.Request

	// http response writer
	ResponseWriter http.ResponseWriter
}

// Start a session
func (s *Store) Start() error {
	s.attributes = make(Value)
	if err := s.intializeSessionID(); err != nil {
		return err
	}
	if err := s.loadSession(); err != nil {
		return err
	}
	return nil
}

func (s *Store) intializeSessionID() error {
	regex, err := regexp.Compile(s.CookieOptions.Name)
	if err != nil {
		return err
	}
	for _, cookie := range s.Request.Cookies() {
		if regex.MatchString(cookie.Name) {
			s.SetID(cookie.Value)
			return nil
		}
	}
	s.SetID(s.generateSessionID())
	return nil
}

func (s *Store) loadSession() error {
	data, err := s.readFromHandler()
	if err != nil {
		return err
	}
	for key, v := range data {
		s.attributes[key] = v
	}
	return nil
}

func (s *Store) readFromHandler() (Value, error) {
	data, err := s.Handler.Read(s.GetID())
	if err != nil {
		if os.IsNotExist(err) {
			return Value{}, nil
		}
		return nil, err
	}
	if len(data) <= 0 {
		return nil, nil
	}
	return s.unserialize(data)
}

func (s *Store) serialize(data Value) ([]byte, error) {
	return json.Marshal(data)
}

func (s *Store) unserialize(data []byte) (Value, error) {
	var values Value
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	err := decoder.Decode(&values)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// Save session
func (s *Store) Save() error {
	data, err := s.serialize(s.attributes)
	if err != nil {
		return err
	}
	s.SetCookie()
	return s.Handler.Write(s.GetID(), data)
}

func (s *Store) SetCookie() {
	cookie := s.CookieOptions.newCookie(s.GetID())
	http.SetCookie(s.ResponseWriter, cookie)
}

// All get all data
func (s *Store) All() Value {
	return s.attributes
}

// Exists check if the key exists
func (s *Store) Exists(key string) bool {
	_, ok := s.attributes[key]
	return ok
}

// Get an item from the session
func (s *Store) Get(key string) (interface{}, error) {
	if !s.Exists(key) {
		return nil, fmt.Errorf("The key [%s] not exists", key)
	}
	return s.attributes[key], nil
}

// Put put a key / value pair or array of key / value pairs in the session
func (s *Store) Put(key string, value interface{}) {
	s.attributes[key] = value
}

// Remove remove one or more items from the session
func (s *Store) Remove(keys ...string) {
	for _, v := range keys {
		delete(s.attributes, v)
	}
}

// Flush flush the session data
func (s *Store) Flush() {
	s.attributes = Value{}
}

// Regenerate generate a new session identifier
func (s *Store) Regenerate() {
	s.Handler.Destroy(s.GetID())
	s.SetID(s.generateSessionID())
}

func (s *Store) generateSessionID() string {
	return randomString(40)
}

func isValidID(id string) bool {
	return len(id) == 40
}

// GetID get session id
func (s *Store) GetID() string {
	return s.id
}

// SetID set session id
func (s *Store) SetID(id string) {
	if isValidID(id) {
		s.id = id
	} else {
		s.id = s.generateSessionID()
	}
}

func (s *Store) GetHandler() HandlerInterface {
	return s.Handler
}
