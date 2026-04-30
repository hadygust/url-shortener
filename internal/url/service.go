package url

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hadygust/url-shortener/internal/cache"
	"github.com/hadygust/url-shortener/internal/dto"
	"github.com/hadygust/url-shortener/internal/model"
	redirectlog "github.com/hadygust/url-shortener/internal/redirect_log"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateUrl(dto.CreateUrlRequest, string) (dto.UrlResponse, error)
	GetAllUserUrl(string) ([]dto.UrlResponse, error)
	GetOrigin(shortCode string, ipAddress string, userAgent string) (string, error)
	DeleteUrl(string, string) (dto.UrlResponse, error)
}

func (s *urlService) CreateUrl(reqUrl dto.CreateUrlRequest, userId string) (dto.UrlResponse, error) {
	userUuid, err := uuid.Parse(userId)
	if err != nil {
		log.Println("user id parsing fails " + err.Error())
		return dto.UrlResponse{}, err
	}

	exp := reqUrl.ExpiresAt
	if exp == nil {
		time := time.Now().Add(time.Hour * 24)
		exp = &time
	}

	url := model.Url{
		ID:          uuid.New(),
		UserId:      userUuid,
		ShortCode:   reqUrl.ShortCode,
		OriginalUrl: reqUrl.OriginalUrl,
		ExpiresAt:   pgtype.Timestamptz{Time: *exp, Valid: true},
	}

	resUrl, err := s.repo.CreateUrl(url)
	if err != nil {
		log.Println("Repo fails " + err.Error())
		return dto.UrlResponse{}, err
	}

	urlResp := dto.NewUrlResponse(resUrl)

	return *urlResp, nil
}

func (s *urlService) GetAllUserUrl(userId string) ([]dto.UrlResponse, error) {
	urls, err := s.repo.GetAllUserUrl(userId)
	if err != nil {
		return []dto.UrlResponse{}, err
	}

	urlResps := []dto.UrlResponse{}

	for _, url := range urls {
		urlResps = append(urlResps, *dto.NewUrlResponse(url))
	}

	return urlResps, nil
}

func (s *urlService) GetOrigin(shortCode string, ipAddress string, userAgent string) (string, error) {

	// Check cache
	urlCache, err := s.cache.GetUrl(shortCode)
	if err == nil && urlCache != nil {
		log.Println("Got from cache")

		s.redirectLog.CreateRedirectLog(urlCache.ID, ipAddress, userAgent)
		return urlCache.OriginalUrl, err
	}

	// Cache failed -> fetch db
	url, err := s.repo.GetUrlbyShortCode(shortCode)
	if err != nil {
		return "", err
	}
	if time.Until(url.ExpiresAt.Time) <= 0 {
		return "", errors.New("url expired")
	}

	// Set cache
	s.cache.SetUrl(url)

	s.redirectLog.CreateRedirectLog(url.ID.String(), ipAddress, userAgent)
	return url.OriginalUrl, err
}

func (s *urlService) DeleteUrl(shortCode string, userId string) (dto.UrlResponse, error) {

	url, err := s.repo.DeleteUrl(shortCode, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.New("not found or not authorized")
		}
		return dto.UrlResponse{}, err
	}

	_ = s.cache.Delete("url:" + shortCode)

	res := *dto.NewUrlResponse(url)

	return res, nil
}

type urlService struct {
	repo        Repository
	redirectLog redirectlog.Service
	cache       cache.Cache
}

func NewService(repo Repository, redirectLog redirectlog.Service, cache cache.Cache) *urlService {
	return &urlService{
		repo:        repo,
		redirectLog: redirectLog,
		cache:       cache,
	}
}
