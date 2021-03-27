package session

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestStore(t *testing.T) {
	var store *Store
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		store = New(r, rw)
		err := store.Start()
		if err != nil {
			t.Fatal(err)
		}
		if len(store.GetID()) != 40 {
			t.Fatal(err)
		}
		store.Put("user_id", 1)

		userID, err := store.Get("user_id")
		if err != nil {
			t.Fatalf("error get session: %v\n", err)
		}

		if value, ok := userID.(int); !ok || value != 1 {
			t.Fatalf("error get session: %s\n", userID)
		}

		store.Save()
		rw.Write([]byte("test"))
	})
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	cookies := w.Result().Cookies()
	sessionID := ""
	for _, cookie := range cookies {
		if cookie.Name == "go_sessid" {
			sessionID = cookie.Value
			break
		}
	}
	handler := store.Handler.(*FileSessionHandler)
	file, err := os.OpenFile(filepath.Join(handler.Path, sessionID), os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("session [%s] not exists\n", sessionID)
		}
		t.Fatal(err)
	}
	finfo, _ := file.Stat()
	if finfo.IsDir() {
		t.Fatalf("session [%s] not exists\n", sessionID)
	}
	defer file.Close()
}

func TestStoreWithCookie(t *testing.T) {
	id := randomString(40)
	values := Value{
		"user_id": 1,
	}
	sessionData, _ := json.Marshal(values)
	err := fileSession.Write(id, sessionData)
	var store *Store
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		store = New(r, rw)
		err := store.Start()
		if err != nil {
			t.Fatal(err)
		}
		if len(store.GetID()) != 40 {
			t.Fatal(err)
		}

		userID, err := store.Get("user_id")
		if err != nil {
			t.Fatalf("error get session [%s]: %v\n", id, err)
		}

		value, err := userID.(json.Number).Int64()
		if err != nil || value != 1 {
			t.Fatalf("wrong value in session: %d, %v\n", userID, value)
		}

		store.Save()
		rw.Write([]byte("test"))
	})
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  "go_sessid",
		Value: id,
	})
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	cookies := w.Result().Cookies()
	sessionID := ""
	for _, cookie := range cookies {
		if cookie.Name == "go_sessid" {
			sessionID = cookie.Value
			break
		}
	}
	handler := store.Handler.(*FileSessionHandler)
	file, err := os.OpenFile(filepath.Join(handler.Path, sessionID), os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("session [%s] not exists\n", sessionID)
		}
		t.Fatal(err)
	}
	finfo, _ := file.Stat()
	if finfo.IsDir() {
		t.Fatalf("session [%s] not exists\n", sessionID)
	}
	defer file.Close()
}
