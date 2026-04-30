package url

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hadygust/url-shortener/internal/dto"
	"github.com/hadygust/url-shortener/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
)

// MockUrlRepository for testing
type MockUrlRepository struct {
	CreateUrlFunc         func(model.Url) (model.Url, error)
	GetAllUserUrlFunc     func(string) ([]model.Url, error)
	GetUrlbyShortCodeFunc func(string) (model.Url, error)
	DeleteUrlFunc         func(string, string) (model.Url, error)
}

func (m *MockUrlRepository) CreateUrl(url model.Url) (model.Url, error) {
	if m.CreateUrlFunc != nil {
		return m.CreateUrlFunc(url)
	}
	return url, nil
}

func (m *MockUrlRepository) GetAllUserUrl(userId string) ([]model.Url, error) {
	if m.GetAllUserUrlFunc != nil {
		return m.GetAllUserUrlFunc(userId)
	}
	return []model.Url{}, nil
}

func (m *MockUrlRepository) GetUrlbyShortCode(shortCode string) (model.Url, error) {
	if m.GetUrlbyShortCodeFunc != nil {
		return m.GetUrlbyShortCodeFunc(shortCode)
	}
	return model.Url{}, nil
}

func (m *MockUrlRepository) DeleteUrl(shortCode string, userId string) (model.Url, error) {
	if m.DeleteUrlFunc != nil {
		return m.DeleteUrlFunc(shortCode, userId)
	}
	return model.Url{}, nil
}

// MockRedirectLogService for testing
type MockRedirectLogService struct {
	CreateRedirectLogFunc func(string, string, string) error
}

func (m *MockRedirectLogService) CreateRedirectLog(urlId string, ipAddress string, userAgent string) error {
	if m.CreateRedirectLogFunc != nil {
		return m.CreateRedirectLogFunc(urlId, ipAddress, userAgent)
	}
	return nil
}

func (m *MockRedirectLogService) GetAllRedirectLogs(urlId string) ([]dto.RedirectLogResponse, error) {
	return []dto.RedirectLogResponse{}, nil
}

// MockCache for URL service testing
type MockUrlCache struct {
	GetUrlFunc func(string) (*dto.UrlCache, error)
	SetUrlFunc func(model.Url) error
	DeleteFunc func(string) error
	GetFunc    func(string) (any, error)
	SetFunc    func(string, any, time.Duration) error
}

func (m *MockUrlCache) GetUrl(shortCode string) (*dto.UrlCache, error) {
	if m.GetUrlFunc != nil {
		return m.GetUrlFunc(shortCode)
	}
	return nil, nil
}

func (m *MockUrlCache) SetUrl(url model.Url) error {
	if m.SetUrlFunc != nil {
		return m.SetUrlFunc(url)
	}
	return nil
}

func (m *MockUrlCache) Delete(key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(key)
	}
	return nil
}

func (m *MockUrlCache) Get(key string) (any, error) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	return nil, nil
}

func (m *MockUrlCache) Set(key string, value any, ttl time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, value, ttl)
	}
	return nil
}

func (m *MockUrlCache) GetUrlPostLimit(userID string) (*interface{}, error) {
	return nil, nil
}

func (m *MockUrlCache) SetUrlPostLimit(userID string, counter *interface{}, ttl time.Duration) error {
	return nil
}

func (m *MockUrlCache) GetUrlGetLimit(key string) (*interface{}, error) {
	return nil, nil
}

func (m *MockUrlCache) SetUrlGetLimit(key string, counter *interface{}, ttl time.Duration) error {
	return nil
}

func TestCreateUrl_Success(t *testing.T) {
	// Arrange
	userId := uuid.New()
	userIdStr := userId.String()
	shortCode := "abc123"
	originalUrl := "https://example.com"

	mockRepo := &MockUrlRepository{
		CreateUrlFunc: func(url model.Url) (model.Url, error) {
			return url, nil
		},
	}
	mockRedirectLog := &MockRedirectLogService{}
	mockCache := &MockUrlCache{}

	service := &urlService{
		repo:        mockRepo,
		redirectLog: mockRedirectLog,
		cache:       mockCache,
	}

	req := dto.CreateUrlRequest{
		ShortCode:   shortCode,
		OriginalUrl: originalUrl,
	}

	// Act
	resp, err := service.CreateUrl(req, userIdStr)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp.ShortCode != shortCode {
		t.Errorf("Expected short code %s, got %s", shortCode, resp.ShortCode)
	}
	if resp.OriginalUrl != originalUrl {
		t.Errorf("Expected original URL %s, got %s", originalUrl, resp.OriginalUrl)
	}
}

func TestCreateUrl_InvalidUserId(t *testing.T) {
	// Arrange
	mockRepo := &MockUrlRepository{}
	mockRedirectLog := &MockRedirectLogService{}
	mockCache := &MockUrlCache{}

	service := &urlService{
		repo:        mockRepo,
		redirectLog: mockRedirectLog,
		cache:       mockCache,
	}

	req := dto.CreateUrlRequest{
		ShortCode:   "abc123",
		OriginalUrl: "https://example.com",
	}

	// Act
	_, err := service.CreateUrl(req, "invalid-uuid")

	// Assert
	if err == nil {
		t.Errorf("Expected error for invalid UUID, got nil")
	}
}

func TestCreateUrl_RepositoryError(t *testing.T) {
	// Arrange
	userId := uuid.New()
	userIdStr := userId.String()

	mockRepo := &MockUrlRepository{
		CreateUrlFunc: func(url model.Url) (model.Url, error) {
			return model.Url{}, errors.New("database error")
		},
	}
	mockRedirectLog := &MockRedirectLogService{}
	mockCache := &MockUrlCache{}

	service := &urlService{
		repo:        mockRepo,
		redirectLog: mockRedirectLog,
		cache:       mockCache,
	}

	req := dto.CreateUrlRequest{
		ShortCode:   "abc123",
		OriginalUrl: "https://example.com",
	}

	// Act
	_, err := service.CreateUrl(req, userIdStr)

	// Assert
	if err == nil {
		t.Errorf("Expected error from repository, got nil")
	}
}

func TestGetAllUserUrl_Success(t *testing.T) {
	// Arrange
	userId := uuid.New()
	userIdStr := userId.String()

	mockUrls := []model.Url{
		{
			ID:          uuid.New(),
			UserId:      userId,
			ShortCode:   "abc123",
			OriginalUrl: "https://example.com",
			CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
			ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
		},
	}

	mockRepo := &MockUrlRepository{
		GetAllUserUrlFunc: func(uId string) ([]model.Url, error) {
			return mockUrls, nil
		},
	}
	mockRedirectLog := &MockRedirectLogService{}
	mockCache := &MockUrlCache{}

	service := &urlService{
		repo:        mockRepo,
		redirectLog: mockRedirectLog,
		cache:       mockCache,
	}

	// Act
	urls, err := service.GetAllUserUrl(userIdStr)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(urls) != 1 {
		t.Errorf("Expected 1 URL, got %d", len(urls))
	}
}

func TestGetOrigin_FromCache(t *testing.T) {
	// Arrange
	shortCode := "abc123"
	originalUrl := "https://example.com"
	urlId := uuid.New()

	mockRepo := &MockUrlRepository{}
	mockRedirectLog := &MockRedirectLogService{
		CreateRedirectLogFunc: func(urlId string, ip string, ua string) error {
			return nil
		},
	}
	mockCache := &MockUrlCache{
		GetUrlFunc: func(code string) (*dto.UrlCache, error) {
			return &dto.UrlCache{
				ID:          urlId.String(),
				OriginalUrl: originalUrl,
			}, nil
		},
	}

	service := &urlService{
		repo:        mockRepo,
		redirectLog: mockRedirectLog,
		cache:       mockCache,
	}

	// Act
	origin, err := service.GetOrigin(shortCode, "127.0.0.1", "Mozilla")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if origin != originalUrl {
		t.Errorf("Expected URL %s, got %s", originalUrl, origin)
	}
}

func TestGetOrigin_FromDatabase(t *testing.T) {
	// Arrange
	shortCode := "abc123"
	originalUrl := "https://example.com"
	userId := uuid.New()

	mockRepo := &MockUrlRepository{
		GetUrlbyShortCodeFunc: func(code string) (model.Url, error) {
			return model.Url{
				ID:          uuid.New(),
				UserId:      userId,
				ShortCode:   code,
				OriginalUrl: originalUrl,
				CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
				ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
			}, nil
		},
	}
	mockRedirectLog := &MockRedirectLogService{
		CreateRedirectLogFunc: func(urlId string, ip string, ua string) error {
			return nil
		},
	}
	mockCache := &MockUrlCache{
		GetUrlFunc: func(code string) (*dto.UrlCache, error) {
			return nil, errors.New("cache miss")
		},
		SetUrlFunc: func(url model.Url) error {
			return nil
		},
	}

	service := &urlService{
		repo:        mockRepo,
		redirectLog: mockRedirectLog,
		cache:       mockCache,
	}

	// Act
	origin, err := service.GetOrigin(shortCode, "127.0.0.1", "Mozilla")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if origin != originalUrl {
		t.Errorf("Expected URL %s, got %s", originalUrl, origin)
	}
}

func TestGetOrigin_ExpiredUrl(t *testing.T) {
	// Arrange
	shortCode := "abc123"
	userId := uuid.New()

	mockRepo := &MockUrlRepository{
		GetUrlbyShortCodeFunc: func(code string) (model.Url, error) {
			return model.Url{
				ID:          uuid.New(),
				UserId:      userId,
				ShortCode:   code,
				OriginalUrl: "https://example.com",
				CreatedAt:   pgtype.Timestamp{Time: time.Now().Add(-48 * time.Hour), Valid: true},
				ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(-24 * time.Hour), Valid: true},
			}, nil
		},
	}
	mockRedirectLog := &MockRedirectLogService{}
	mockCache := &MockUrlCache{
		GetUrlFunc: func(code string) (*dto.UrlCache, error) {
			return nil, errors.New("cache miss")
		},
	}

	service := &urlService{
		repo:        mockRepo,
		redirectLog: mockRedirectLog,
		cache:       mockCache,
	}

	// Act
	_, err := service.GetOrigin(shortCode, "127.0.0.1", "Mozilla")

	// Assert
	if err == nil {
		t.Errorf("Expected error for expired URL, got nil")
	}
	if err.Error() != "url expired" {
		t.Errorf("Expected 'url expired' error, got %v", err)
	}
}

func TestDeleteUrl_Success(t *testing.T) {
	// Arrange
	userId := uuid.New()
	userIdStr := userId.String()
	shortCode := "abc123"

	deletedUrl := model.Url{
		ID:          uuid.New(),
		UserId:      userId,
		ShortCode:   shortCode,
		OriginalUrl: "https://example.com",
		CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
		ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
	}

	mockRepo := &MockUrlRepository{
		DeleteUrlFunc: func(code string, uid string) (model.Url, error) {
			return deletedUrl, nil
		},
	}
	mockRedirectLog := &MockRedirectLogService{}
	mockCache := &MockUrlCache{
		DeleteFunc: func(key string) error {
			return nil
		},
	}

	service := &urlService{
		repo:        mockRepo,
		redirectLog: mockRedirectLog,
		cache:       mockCache,
	}

	// Act
	resp, err := service.DeleteUrl(shortCode, userIdStr)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp.ShortCode != shortCode {
		t.Errorf("Expected short code %s, got %s", shortCode, resp.ShortCode)
	}
}

func TestDeleteUrl_NotFound(t *testing.T) {
	// Arrange
	userId := uuid.New()
	userIdStr := userId.String()
	shortCode := "nonexistent"

	mockRepo := &MockUrlRepository{
		DeleteUrlFunc: func(code string, uid string) (model.Url, error) {
			return model.Url{}, errors.New("not found")
		},
	}
	mockRedirectLog := &MockRedirectLogService{}
	mockCache := &MockUrlCache{}

	service := &urlService{
		repo:        mockRepo,
		redirectLog: mockRedirectLog,
		cache:       mockCache,
	}

	// Act
	_, err := service.DeleteUrl(shortCode, userIdStr)

	// Assert
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
