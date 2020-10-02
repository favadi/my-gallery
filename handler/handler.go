package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"

	"github.com/favadi/my-gallery/auth"
	"github.com/favadi/my-gallery/storage"
)

type Storage interface {
	Images(userID int64) ([]storage.Image, error)
	CreateLike(userID, imageID int64) (int64, error)
}

type server struct {
	templates     *template.Template
	auth          *auth.Authenticator
	decoder       *schema.Decoder
	storage       Storage
	sessionsStore sessions.Store
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	var user = r.Context().Value(contextUserKey{}).(auth.User)
	tmpl := s.templates.Lookup("index.html")
	if tmpl == nil {
		http.Error(w, "unable to load template", http.StatusInternalServerError)
		return
	}

	images, err := s.storage.Images(user.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load images: err=%s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err = tmpl.Execute(w, struct {
		CSRFField template.HTML
		User      auth.User
		Images    []storage.Image
	}{
		CSRFField: csrf.TemplateField(r),
		User:      user,
		Images:    images,
	}); err != nil {
		log.Printf("failed to render template: err=%s", err.Error())
	}
}

type likeImageRequest struct {
	ImageID int64 `json:"image_id"`
}

type likeImageResponse struct {
	ID int64 `json:"id"`
}

func (s *server) likeImage(w http.ResponseWriter, r *http.Request) {
	var (
		user = r.Context().Value(contextUserKey{}).(auth.User)
		data likeImageRequest
	)
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	likeID, err := s.storage.CreateLike(user.ID, data.ImageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("user %d likes image %d", user.ID, likeID)
	if err = json.NewEncoder(w).Encode(&likeImageResponse{ID: likeID}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func New(storage Storage, imagesDir string, sessionsStore sessions.Store, auth *auth.Authenticator, csrfMW mux.MiddlewareFunc) (http.Handler, error) {
	templates, err := template.New("my-gallery").ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	s := &server{
		templates:     templates,
		auth:          auth,
		decoder:       decoder,
		storage:       storage,
		sessionsStore: sessionsStore,
	}

	r := mux.NewRouter()
	r.Use(csrfMW)

	r.PathPrefix("/css/").Methods(http.MethodGet).Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir(filepath.Join(imagesDir, "css")))))
	r.PathPrefix("/js/").Methods(http.MethodGet).Handler(
		http.StripPrefix("/js/", http.FileServer(http.Dir(filepath.Join(imagesDir, "js")))))
	r.PathPrefix("/gallery/").Methods(http.MethodGet).Handler(
		http.StripPrefix("/gallery/", http.FileServer(http.Dir(filepath.Join(imagesDir, "gallery")))))

	r.Path("/login").Methods(http.MethodGet).HandlerFunc(s.showLogin)
	r.Path("/login").Methods(http.MethodPost).HandlerFunc(s.login)
	r.Path("/logout").Methods(http.MethodPost).HandlerFunc(s.logout)

	protected := r.PathPrefix("/").Subrouter()
	protected.Use(s.authMiddleware)
	protected.Path("/").Methods(http.MethodGet).HandlerFunc(s.index)
	protected.Path("/likes").Methods(http.MethodPost).HandlerFunc(s.likeImage)

	return handlers.LoggingHandler(os.Stdout, r), nil
}
