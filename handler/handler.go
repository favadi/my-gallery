package handler

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"

	"github.com/favadi/my-gallery/auth"
)

type storage interface {
	AllImages() ([]Image, error)
}

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(db *sqlx.DB) *Postgres {
	return &Postgres{db: db}
}

type Image struct {
	ID      int64     `db:"id"`
	Name    string    `db:"name"`
	Format  string    `db:"format"`
	Created time.Time `db:"created"`
}

func (p *Postgres) AllImages() ([]Image, error) {
	const query = `SELECT id, name, format, created
FROM images;`
	images := make([]Image, 0)
	if err := p.db.Select(&images, query); err != nil {
		return nil, err
	}
	return images, nil
}

type server struct {
	templates     *template.Template
	auth          *auth.Authenticator
	decoder       *schema.Decoder
	storage       storage
	sessionsStore sessions.Store
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	tmpl := s.templates.Lookup("index.html")
	if tmpl == nil {
		http.Error(w, "unable to load template", http.StatusInternalServerError)
		return
	}

	images, err := s.storage.AllImages()
	if err != nil {
		http.Error(w, "failed to load images", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, struct {
		User   auth.User
		Images []Image
	}{
		User:   r.Context().Value(contextUserKey{}).(auth.User),
		Images: images,
	}); err != nil {
		log.Printf("failed to render template: err=%s", err.Error())
	}
}

func New(storage storage, auth *auth.Authenticator, imagesDir string, sessionsStore sessions.Store) (http.Handler, error) {
	templates, err := template.New("my-gallery").ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}

	s := &server{
		templates:     templates,
		auth:          auth,
		decoder:       schema.NewDecoder(),
		storage:       storage,
		sessionsStore: sessionsStore,
	}

	r := mux.NewRouter()
	r.Path("/login").Methods(http.MethodGet).HandlerFunc(s.showLogin)
	r.Path("/login").Methods(http.MethodPost).HandlerFunc(s.login)
	r.Path("/logout").Methods(http.MethodPost).HandlerFunc(s.logout)

	protected := r.Path("/").Subrouter()
	protected.Use(s.authMiddleware)
	protected.Methods(http.MethodGet).HandlerFunc(s.index)

	r.PathPrefix("/css/").Methods(http.MethodGet).Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir(filepath.Join(imagesDir, "css")))))
	r.PathPrefix("/gallery/").Methods(http.MethodGet).Handler(
		http.StripPrefix("/gallery/", http.FileServer(http.Dir(filepath.Join(imagesDir, "gallery")))))

	return handlers.LoggingHandler(os.Stdout, r), nil
}
