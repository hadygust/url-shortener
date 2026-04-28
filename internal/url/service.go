package url

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hadygust/url-shortener/internal/cache"
	"github.com/hadygust/url-shortener/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateUrl(CreateUrlRequest, string) (UrlResponse, error)
	GetAllUserUrl(string) ([]UrlResponse, error)
	GetOrigin(string) (string, error)
	DeleteUrl(string, string) (UrlResponse, error)
}

func (s *urlService) CreateUrl(reqUrl CreateUrlRequest, userId string) (UrlResponse, error) {
	userUuid, err := uuid.Parse(userId)
	if err != nil {
		log.Println("user id parsing fails " + err.Error())
		return UrlResponse{}, err
	}

	url := model.Url{
		ID:          uuid.New(),
		UserId:      userUuid,
		ShortCode:   reqUrl.ShortCode,
		OriginalUrl: reqUrl.OriginalUrl,
		ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(time.Hour * 24), Valid: true},
	}

	resUrl, err := s.repo.CreateUrl(url)
	if err != nil {
		log.Println("Repo fails " + err.Error())
		return UrlResponse{}, err
	}

	urlResp := NewUrlResponse(resUrl)

	return *urlResp, nil
}

func (s *urlService) GetAllUserUrl(userId string) ([]UrlResponse, error) {
	urls, err := s.repo.GetAllUserUrl(userId)
	if err != nil {
		return []UrlResponse{}, err
	}

	urlResps := []UrlResponse{}

	for _, url := range urls {
		urlResps = append(urlResps, *NewUrlResponse(url))
	}

	return urlResps, nil
}

func (s *urlService) GetOrigin(shortCode string) (string, error) {

	// Check cache
	origin, err := s.cache.Get("shortCode:" + shortCode)
	if err != nil {
		return "", err
	}

	if origin != nil {
		originStr, ok := origin.(string)
		if !ok {
			return "", errors.New("cache type mismatch")
		}
		return originStr, nil
	}

	originStr, err := s.repo.GetOrigin(shortCode)
	return originStr, err
}

func (s *urlService) DeleteUrl(shortCode string, userId string) (UrlResponse, error) {
	url, err := s.repo.DeleteUrl(shortCode)
	if err != nil {
		return UrlResponse{}, err
	}

	if url.UserId.String() != userId {
		return UrlResponse{}, errors.New("You are not authorized to delete this url")
	}

	res := *NewUrlResponse(url)

	return res, nil
}

type urlService struct {
	repo  Repository
	cache cache.Cache
}

func NewService(repo Repository, cache cache.Cache) *urlService {
	return &urlService{
		repo:  repo,
		cache: cache,
	}
}
