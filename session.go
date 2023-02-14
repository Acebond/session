package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

type sessionKey struct{}

// SessionStore holds the session data and settings
type SessionStore[T any] struct {
	name       string
	sessions   map[string]*T
	lock       sync.RWMutex
	ctxKey     sessionKey
	expiration time.Duration
}

// Init will initialize the SessionStore object
func (st *SessionStore[T]) InitStore(name string, itemExpiry time.Duration) {
	st.name = name
	st.sessions = make(map[string]*T)
	st.ctxKey = sessionKey{}
	st.expiration = itemExpiry
}

func randBase64String(entropyBytes int) string {
	b := make([]byte, entropyBytes)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// PutSession will store the session in the SessionStore.
// The session will automatically expire after defined SessionStore.sessionExpiration.
func (st *SessionStore[T]) PutSession(w http.ResponseWriter, r *http.Request, sess *T) {
	cookieValue := randBase64String(33) // 33 bytes entropy

	time.AfterFunc(st.expiration, func() {
		st.lock.Lock()
		delete(st.sessions, cookieValue)
		st.lock.Unlock()
	})

	st.lock.Lock()
	st.sessions[cookieValue] = sess
	st.lock.Unlock()

	cookie := &http.Cookie{
		Name:     st.name,
		Value:    cookieValue,
		Expires:  time.Now().Add(st.expiration),
		HttpOnly: true,
		Secure:   r.URL.Scheme == "https",
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

// DeleteSession will delete the session from the SessionStore.
func (st *SessionStore[T]) DeleteSession(r *http.Request) {
	cookie, err := r.Cookie(st.name)
	if err != nil {
		return
	}
	st.lock.Lock()
	delete(st.sessions, cookie.Value)
	st.lock.Unlock()
}

// GetSessionFromRequest retrieves the session from the http.Request cookies.
// The function will return nil if the session does not exist within the http.Request cookies.
func (st *SessionStore[T]) GetSessionFromRequest(r *http.Request) *T {
	cookie, err := r.Cookie(st.name)
	if err != nil {
		return nil
	}
	st.lock.RLock()
	sess := st.sessions[cookie.Value]
	st.lock.RUnlock()
	return sess
}

// LoadSession will load the session into the http.Request context.
// A http.StatusUnauthorized will be retuned to the client if no session can be found.
func (st *SessionStore[T]) LoadSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := st.GetSessionFromRequest(r)
		if sess == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), st.ctxKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetSessionFromCtx retrieves the session from the http.Request context.
// The function will return nil if the session does not exist within the http.Request context.
func (st *SessionStore[T]) GetSessionFromCtx(r *http.Request) *T {
	return r.Context().Value(st.ctxKey).(*T)
}
