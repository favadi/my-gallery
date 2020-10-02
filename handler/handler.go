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
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
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
	storage       storage
	templates     *template.Template
	imagesDir     string
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
		Images []Image
	}{
		Images: images,
	}); err != nil {
		log.Printf("failed to render template: err=%s", err.Error())
	}
}

func New(storage storage, imagesDir string, sessionsStore sessions.Store) (http.Handler, error) {
	templates, err := template.New("my-gallery").ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}

	s := &server{templates: templates, imagesDir: imagesDir, storage: storage, sessionsStore: sessionsStore}

	r := mux.NewRouter()

	r.Path("/").Methods(http.MethodGet).HandlerFunc(s.index)

	r.Path("/login").Methods(http.MethodGet).HandlerFunc(s.login)

	r.PathPrefix("/css/").Methods(http.MethodGet).Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir(filepath.Join(imagesDir, "css")))))
	r.PathPrefix("/gallery/").Methods(http.MethodGet).Handler(
		http.StripPrefix("/gallery/", http.FileServer(http.Dir(filepath.Join(imagesDir, "gallery")))))

	return handlers.LoggingHandler(os.Stdout, r), nil
}
