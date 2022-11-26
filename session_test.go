package session

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type Session struct {
	name string
}

func FakeUserLogin(ss *SessionStore[Session]) {
	req, _ := http.NewRequest("GET", "/whatever", nil)
	rr := httptest.NewRecorder()
	ss.PutSession(rr, req, &Session{name: "testing"})
}

func TestSessionStore(t *testing.T) {
	var ss SessionStore[Session]
	ss.InitStore("test", time.Duration(time.Minute*5))

	var wg sync.WaitGroup
	numGoroutines := 20
	wg.Add(numGoroutines)
	for i := 0; i <= numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for created := 0; created <= 500000; created++ {
				FakeUserLogin(&ss)
			}
		}()
	}
	wg.Wait()
	fmt.Println("Done, simulated 10 million user sessions")

}

func BenchmarkSessionCreation(b *testing.B) {
	var ss SessionStore[Session]
	ss.InitStore("test", time.Duration(time.Minute))

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/whatever", nil)
	if err != nil {
		b.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		ss.PutSession(rr, req, &Session{name: "testing"})
	}
}
