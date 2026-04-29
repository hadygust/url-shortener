package redirectlog

import (
	"log"

	"github.com/hadygust/url-shortener/internal/model"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	CreateRedirectLog(model.RedirectLog) error
}

func (repo *rlogRepository) CreateRedirectLog(rlog model.RedirectLog) error {

	query := `
		INSERT INTO redirect_logs (id, url_id, ip_address, user_agent)
		VALUES ($1, $2, $3, $4)
	`

	_, err := repo.db.Exec(query, rlog.ID, rlog.UrlId, rlog.IpAddress, rlog.UserAgent)
	if err != nil {
		log.Printf("rlog insertion failed: %s", err.Error())
		return err
	}

	log.Println("rlog insertion success")
	return nil
}

type rlogRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &rlogRepository{
		db: db,
	}
}
