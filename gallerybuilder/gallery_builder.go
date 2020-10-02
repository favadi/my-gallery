package gallerybuilder

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// allowedImageFormats is the list of allowed image file types.
// https://mimesniff.spec.whatwg.org/#matching-an-image-type-pattern
var allowedImageFormats = map[string]struct{}{
	"image/x-icon": {},
	"image/bmp":    {},
	"image/gif":    {},
	"image/webp":   {},
	"image/png":    {},
	"image/jpeg":   {},
}

type storage interface {
	Store([]Image) error
}

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(db *sqlx.DB) *Postgres {
	return &Postgres{db: db}
}

func (p *Postgres) Store(images []Image) error {
	const query = `INSERT INTO images (name, format)
VALUES (unnest($1::TEXT[]), unnest($2::TEXT[]))
ON CONFLICT DO NOTHING;`

	var (
		names   []string
		formats []string
	)
	for _, img := range images {
		names = append(names, img.Name)
		formats = append(formats, img.Format)
	}
	_, err := p.db.Exec(query, pq.Array(names), pq.Array(formats))
	return err
}

type Gallery struct {
	storage   storage
	imagesDir string
}

func NewGallery(storage storage, imagesDir string) *Gallery {
	return &Gallery{storage: storage, imagesDir: imagesDir}
}

type Image struct {
	Name   string
	Format string
	Valid  bool
}

func parseImage(filePath string) (Image, error) {
	imgFile, err := os.Open(filePath)
	if err != nil {
		return Image{}, err
	}
	defer func() { _ = imgFile.Close() }()

	imgHeaderBytes := make([]byte, 512)
	if _, err = imgFile.Read(imgHeaderBytes); err != nil {
		return Image{}, err
	}

	contentType := http.DetectContentType(imgHeaderBytes)

	img := Image{
		Name:   filepath.Base(filePath),
		Format: contentType,
		Valid:  false,
	}

	if _, ok := allowedImageFormats[contentType]; ok {
		img.Valid = true
	}
	return img, nil
}

func (g *Gallery) lookupImages() ([]Image, error) {
	var images []Image
	err := filepath.Walk(g.imagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		img, err := parseImage(path)
		if err != nil {
			return err
		}

		if img.Valid {
			images = append(images, img)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return images, nil
}

// Build finds all images in given directory to build a gallery and insert to database.
func (g *Gallery) Build() error {
	log.Printf("looking up for images")
	images, err := g.lookupImages()
	if err != nil {
		return err
	}
	log.Printf("storing images metadata to database: images=%d", len(images))
	return g.storage.Store(images)
}
