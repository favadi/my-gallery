package handler

import (
	"context"
	"log"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/favadi/my-gallery/auth"
)

const (
	authCookieName    = "auth_session"
	authCookieUserKey = "user"
)

type contextUserKey struct{}

func (s *server) authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sess, err := s.sessionsStore.Get(r, authCookieName)
		if err == nil && !sess.IsNew {
			user, ok := sess.Values[authCookieUserKey].(auth.User)
			if ok {
				h.ServeHTTP(w, r.WithContext(context.WithValue(ctx, contextUserKey{}, user)))
				return
			}
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
}

type showLoginFormData struct {
	Errors validation.Errors
}

func (s *server) renderLoginForm(w http.ResponseWriter, r *http.Request, data showLoginFormData) {
	tmpl := s.templates.Lookup("login.html")
	if tmpl == nil {
		http.Error(w, "unable to load template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("failed to render template: err=%s", err.Error())
	}
}

func (s *server) showLogin(w http.ResponseWriter, r *http.Request) {
	s.renderLoginForm(w, r, showLoginFormData{})
}

type loginForm struct {
	Username string
	Password string
}

func (f *loginForm) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, f,
		validation.Field(&f.Username, validation.Required, validation.Length(1, 32)),
		validation.Field(&f.Password, validation.Required, validation.Length(1, 70)),
	)
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var form loginForm
	if err := s.decoder.Decode(&form, r.Form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := form.Validate(r.Context()); err != nil {
		if vErrs, ok := err.(validation.Errors); ok {
			s.renderLoginForm(w, r, showLoginFormData{Errors: vErrs})
			return
		}
	}

	user, err := s.auth.Authenticate(form.Username, form.Password)
	if err == auth.ErrInvalidCredentials {
		s.renderLoginForm(w, r, showLoginFormData{
			map[string]error{"General": err},
		})
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sess, err := s.sessionsStore.New(r, authCookieName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sess.Values[authCookieUserKey] = user
	if err = sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}

func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	sess, err := s.sessionsStore.Get(r, authCookieName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sess.Options.MaxAge = -1 // remove cookie
	if err = sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
	return
}
