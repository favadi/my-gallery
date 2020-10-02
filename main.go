package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/favadi/my-gallery/gallerybuilder"
	"github.com/favadi/my-gallery/handler"
)

const driver = "postgres"

func main() {
	var (
		addr         = flag.String("addr", ":5000", "address to listen to")
		dbString     = flag.String("data-source-name", "postgres://my-gallery:my-gallery@127.0.0.1:5432/my-gallery?sslmode=disable", "PostgresQL database DSN")
		assetsDir    = flag.String("assets-dir", "assets", "path to images directory")
		cookieSecret = flag.String("cookie-secret", "change-me-please", "path to images directory")
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

	cookieStore := sessions.NewCookieStore([]byte(*cookieSecret))
	cookieStore.Options.HttpOnly = true

	h, err := handler.New(handler.NewPostgres(db), *assetsDir, cookieStore)
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
