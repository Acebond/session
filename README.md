# session
Fast in-memory backend session storage middleware.

## Usage

```go
package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/Acebond/session"
)

type Session struct {
	Role     string
	Username string
}

var (
	ss session.SessionStore[Session]
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		html := `<html><head></head><form action="/login" method="post"><input type="text" name="username">
		<input type="password" name="password"><input type="submit" value="Submit"></form></head>`
		w.Write([]byte(html))
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "admin" && password == "admin" {
			sess := &Session{}
			sess.Role = "Administrator"
			sess.Username = "Admin"
			ss.PutSession(w, r, sess)
			http.Redirect(w, r, "/admin", http.StatusFound)
		} else {
			http.Error(w, "Incorrect username or password", http.StatusUnauthorized)
		}
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	ss.DeleteSession(r)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func adminPanelHandler(w http.ResponseWriter, r *http.Request) {
	sess := ss.GetSessionFromCtx(r)
	if sess.Role != "Administrator" {
		http.Error(w, "", http.StatusForbidden)
		return
	}
	http.Error(w, fmt.Sprintf("Hello %s\n", sess.Username), http.StatusOK)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	ss.InitStore("SessionID", time.Duration(time.Hour*24*7)) // 1 week
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.Handle("/admin", ss.LoadSession(http.HandlerFunc(adminPanelHandler)))
	http.ListenAndServe(":8090", nil)
}
```