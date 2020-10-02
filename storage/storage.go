package storage

import (
	"time"

	"github.com/jmoiron/sqlx"
)

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
	Size    int64     `db:"size"`
	Created time.Time `db:"created"`
	Liked   bool      `db:"liked"`
}

func (p *Postgres) Images(userID int64) ([]Image, error) {
	const query = `SELECT images.id,
       images.name,
       images.format,
       images.created,
       images.size,
       CASE WHEN likes.id IS NOT NULL THEN TRUE ELSE FALSE END AS liked
FROM images
         LEFT JOIN likes ON images.id = likes.image_id AND likes.user_id = $1
ORDER BY created, name;`
	images := make([]Image, 0)
	if err := p.db.Select(&images, query, userID); err != nil {
		return nil, err
	}
	return images, nil
}

func (p *Postgres) Like(userID, imageID int64) (int64, error) {
	const query = `INSERT INTO likes(user_id, image_id)
VALUES ($1, $2)
ON CONFLICT (user_id, image_id) DO UPDATE SET updated = now()
RETURNING id;`
	var id int64
	err := p.db.Get(&id, query, userID, imageID)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (p *Postgres) Unlike(userID, imageID int64) error {
	const query = `DELETE
FROM likes
WHERE user_id = $1
  AND image_id = $2;`
	_, err := p.db.Exec(query, userID, imageID)
	return err
}
