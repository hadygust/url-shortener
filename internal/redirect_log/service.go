package redirectlog

import (
	"github.com/google/uuid"
	"github.com/hadygust/url-shortener/internal/dto"
	"github.com/hadygust/url-shortener/internal/model"
)

type Service interface {
	CreateRedirectLog(urlId string, ipAddress string, userAgent string) error
	GetUrlDailyClicks(shortCode string) ([]dto.DailyClicks, error)
	GetTotalClicks(shortCode string) (int, error)
	GetTopUserAgents(shortCode string) ([]dto.UserAgent, error)
	GetStats(shortCode string) (dto.UrlStatsResponse, error)
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

func (s *rlogService) GetUrlDailyClicks(shortCode string) ([]dto.DailyClicks, error) {
	return s.repo.GetUrlDailyClicks(shortCode)
}

func (s *rlogService) GetTotalClicks(shortCode string) (int, error) {
	return s.repo.GetTotalClicks(shortCode)
}

func (s *rlogService) GetTopUserAgents(shortCode string) ([]dto.UserAgent, error) {
	return s.repo.GetTopUserAgents(shortCode)
}

func (s *rlogService) GetStats(shortCode string) (dto.UrlStatsResponse, error) {
	return s.repo.GetStats(shortCode)
}

type rlogService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &rlogService{
		repo: repo,
	}
}
