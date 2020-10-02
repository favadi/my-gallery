package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/favadi/my-gallery/auth"
	"github.com/favadi/my-gallery/gallerybuilder"
	"github.com/favadi/my-gallery/handler"
	"github.com/favadi/my-gallery/storage"
)

const driver = "postgres"

const (
	environmentDev  = "development"
	environmentProd = "production"
)

func main() {
	var (
		environment  = flag.String("env", environmentDev, "application running mode")
		addr         = flag.String("addr", ":5000", "address to listen to")
		dbString     = flag.String("data-source-name", "postgres://my-gallery:my-gallery@127.0.0.1:5432/my-gallery?sslmode=disable", "PostgresQL database DSN")
		assetsDir    = flag.String("assets-dir", "assets", "path to images directory")
		cookieSecret = flag.String("cookie-secret", "change-me-please", "secret key to sign auth cookie")
		csrfSecret   = flag.String("csrf-secret", "change-me-please", "secret key to generate CSRF token")
	)

	flag.Parse()

	db, err := sqlx.Open(driver, *dbString)
	if err != nil {
		log.Fatal(err)
	}

	builder := gallerybuilder.NewGallery(gallerybuilder.NewPostgres(db), filepath.Join(*assetsDir, "gallery"))
	if err = builder.Build(); err != nil {
		log.Fatal(err)
	}

	gob.Register(auth.User{}) // to store user information to cookie
	cookieStore := sessions.NewCookieStore([]byte(*cookieSecret))
	cookieStore.Options.HttpOnly = true
	if *environment == environmentProd {
		cookieStore.Options.Secure = true
	}

	var csrfOptions []csrf.Option
	if *environment == environmentDev {
		csrfOptions = append(csrfOptions, csrf.Secure(false))
	}
	csrfMW := csrf.Protect([]byte(*csrfSecret), csrfOptions...)

	h, err := handler.New(storage.NewPostgres(db), *assetsDir, cookieStore, auth.NewAuthenticator(db), csrfMW)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:    *addr,
		Handler: h,
	}
	log.Printf("starting server at %s", *addr)
	log.Fatal(server.ListenAndServe())
}
