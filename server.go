package sus

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

func randString(l int) func() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func() string {
		bs := make([]byte, l)
		for i := range bs {
			bs[i] = chars[rng.Intn(len(chars))]
		}
		return string(bs)
	}
}

type Shortener struct {
	name   string
	keygen func() string
	tmpl   *template.Template
	db     *bolt.DB
}

func NewShortener(name string, keylen int, db *bolt.DB) *Shortener {
	if db == nil || keylen < 0 {
		return nil
	}
	tmpl := template.Must(template.ParseFiles("tmpl/index.html"))
	s := Shortener{name, randString(keylen), tmpl, db}
	return &s
}

func (s *Shortener) executeTmpl(w http.ResponseWriter, val string) {
	err := s.tmpl.Execute(w, val)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Shortener) HandlePost(w http.ResponseWriter, r *http.Request) {
	rawURL := r.FormValue("url")
	if rawURL == "" {
		s.executeTmpl(w, "Error: cannot shorten empty URL!")
		return
	}
	url, err := url.Parse(rawURL)
	if err != nil {
		s.executeTmpl(w, "Error: submitted URL is incorrect!")
		return
	}
	key := s.keygen()
	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		err := b.Put([]byte(key), []byte(url.String()))
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.executeTmpl(w, s.name+"/"+key)
}

func (s *Shortener) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	var val string
	_ = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		val = string(b.Get([]byte(strings.TrimLeft(r.URL.Path, "/"))))
		return nil
	})
	if val != "" {
		http.Redirect(w, r, val, http.StatusFound)
		return
	}
	http.NotFound(w, r)
}

func (s *Shortener) HandleGet(w http.ResponseWriter, r *http.Request) {
	s.executeTmpl(w, "")
}

type Route struct {
	method string
	regexp *regexp.Regexp
	handle http.HandlerFunc
}

func NewRoute(method string, pattern string, h http.HandlerFunc) Route {
	return Route{method, regexp.MustCompile(pattern), h}
}

func Handle(routes []Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var okMethods []string
		method := r.Method
		for _, route := range routes {
			if route.regexp.MatchString(r.URL.Path) {
				if r.Method != route.method {
					okMethods = append(okMethods, route.method)
					continue
				}
				route.handle(w, r)
				return
			}
		}
		if len(okMethods) == 0 {
			http.NotFound(w, r)
			return
		}
		http.Error(w, fmt.Sprintf("Error: 405 method %v not allowed", method),
			http.StatusMethodNotAllowed)
	}
}

func CacheRedirect(h http.HandlerFunc, ttl time.Duration, db *bolt.DB) http.HandlerFunc {
	m := sync.Mutex{}
	cache := make(map[string]string)
	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimLeft(r.URL.Path, "/")
		if val, ok := cache[key]; ok {
			http.Redirect(w, r, val, http.StatusFound)
			return
		}
		h(w, r)
		var val string
		_ = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("urls"))
			val = string(b.Get([]byte(strings.TrimLeft(r.URL.Path, "/"))))
			return nil
		})
		if val == "" {
			return
		}
		m.Lock()
		cache[key] = val
		m.Unlock()
		t := time.NewTimer(ttl)
		go func() {
			<-t.C
			m.Lock()
			defer m.Unlock()
			delete(cache, key)
		}()
	}
}
