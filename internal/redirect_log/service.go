package redirectlog

import (
	"github.com/google/uuid"
	"github.com/hadygust/url-shortener/internal/model"
)

type Service interface {
	CreateRedirectLog(urlId string, ipAddress string, userAgent string) error
}

func (s *rlogService) CreateRedirectLog(urlId string, ipAddress string, userAgent string) error {
	id := uuid.New()

	rlog := model.RedirectLog{
		ID:        id,
		UrlId:     uuid.MustParse(urlId),
		IpAddress: ipAddress,
		UserAgent: userAgent,
	}

	err := s.repo.CreateRedirectLog(rlog)
	if err != nil {
		return err
	}

	return nil
}

type rlogService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &rlogService{
		repo: repo,
	}
}
