package url

import (
	"log"

	"github.com/google/uuid"
	"github.com/hadygust/url-shortener/internal/model"
)

type Service interface {
	CreateUrl(CreateUrlRequest, string) (UrlResponse, error)
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
	}

	resUrl, err := s.repo.CreateUrl(url)
	if err != nil {
		log.Println("Repo fails " + err.Error())
		return UrlResponse{}, err
	}

	urlResp := NewUrlResponse(resUrl)

	return *urlResp, nil
}

type urlService struct {
	repo Repository
}

func NewService(repo Repository) *urlService {
	return &urlService{
		repo: repo,
	}
}
