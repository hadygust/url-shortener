package redirectlog

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hadygust/url-shortener/internal/dto"
	"github.com/hadygust/url-shortener/internal/model"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	CreateRedirectLog(model.RedirectLog) error
	GetUrlDailyClicks(shortCode string) ([]dto.DailyClicks, error)
	GetTotalClicks(shortCode string) (int, error)
	GetTopUserAgents(shortCode string) ([]dto.UserAgent, error)
	GetStats(shortCode string) (dto.UrlStatsResponse, error)
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

func (repo *rlogRepository) GetUrlClickCount(urlId string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM redirect_logs
		WHERE url_id = '8fd023fc-efe3-4093-a647-38e0e0de22dc';
	`

	fmt.Println(query)
	return 0, nil
}

func (repo *rlogRepository) GetUrlDailyClicks(shortCode string) ([]dto.DailyClicks, error) {
	// TODO: redirect log find cant search with short code. rlog schema only hav urlId identifier
	query := `
		SELECT 
			d.day AS date,
			COUNT(log.id) AS count
		FROM 
			generate_series(
				date_trunc('day', now()) - INTERVAL '6 days',
				date_trunc('day', now()),
				INTERVAL '1 day'
			) AS d(day)
		LEFT JOIN (
			SELECT r.id, r.accessed_at
			FROM redirect_logs r
			JOIN urls u
				ON u.id = r.url_id
			WHERE u.short_code = $1
		) log ON date_trunc('day', log.accessed_at) = d.day
		GROUP BY d.day
		ORDER BY d.day;
	`

	dClicks := []dto.DailyClicks{}
	rows, err := repo.db.Queryx(query, shortCode)
	if err != nil {
		return []dto.DailyClicks{}, err
	}

	for rows.Next() {
		dClick := dto.DailyClicks{}
		err := rows.StructScan(&dClick)
		if err != nil {
			return []dto.DailyClicks{}, err
		}
		dClicks = append(dClicks, dClick)
	}

	return dClicks, nil
}

func (repo *rlogRepository) GetTotalClicks(shortCode string) (int, error) {
	query := `
		SELECT COUNT(*) count
		FROM redirect_logs AS r
		JOIN urls AS u 
			ON u.id = r.url_id
		WHERE u.short_code = $1
	`
	var count int
	err := repo.db.Get(&count, query, shortCode)

	return count, err
}

func (repo *rlogRepository) GetTopUserAgents(shortCode string) ([]dto.UserAgent, error) {
	query := `
		SELECT r.user_agent name, COUNT(*) count
		FROM urls AS u
		JOIN redirect_logs as r
			ON u.id = r.url_id
			AND u.short_code = $1
		GROUP BY r.user_agent
		ORDER BY COUNT(*) DESC
	`

	userAgents := []dto.UserAgent{}
	rows, err := repo.db.Queryx(query, shortCode)
	if err != nil {
		return []dto.UserAgent{}, err
	}

	for rows.Next() {
		agent := dto.UserAgent{}
		err := rows.StructScan(&agent)
		if err != nil {
			return []dto.UserAgent{}, err
		}
		userAgents = append(userAgents, agent)
	}

	return userAgents, nil
}

func (repo *rlogRepository) GetStats(shortCode string) (dto.UrlStatsResponse, error) {
	query := `
	WITH url_target AS (
		SELECT id FROM urls WHERE short_code = $1
	),
	base AS (
		SELECT r.user_agent, r.accessed_at
		FROM redirect_logs r
		JOIN url_target u ON r.url_id = u.id
	),
	total AS (
		SELECT COUNT(*) AS total_clicks FROM base
	),
	daily AS (
		SELECT 
			date_trunc('day', accessed_at) AS date,
			COUNT(*) AS count
		FROM base
		WHERE accessed_at >= now() - INTERVAL '7 days'
		GROUP BY date
	),
	agents AS (
		SELECT 
			user_agent AS name,
			COUNT(*) AS count
		FROM base
		GROUP BY user_agent
		ORDER BY count DESC
		LIMIT 5
	)
	SELECT 
		(SELECT total_clicks FROM total) AS total_clicks,
		COALESCE((SELECT json_agg(daily ORDER BY date) FROM daily), '[]') AS daily_clicks,
		COALESCE((SELECT json_agg(agents) FROM agents), '[]') AS top_agents;
	`

	var (
		totalClicks int
		dailyJSON   []byte
		agentsJSON  []byte
	)

	err := repo.db.QueryRow(query, shortCode).Scan(
		&totalClicks,
		&dailyJSON,
		&agentsJSON,
	)
	if err != nil {
		return dto.UrlStatsResponse{}, err
	}

	var dailyClicks []dto.DailyClicks
	var topAgents []dto.UserAgent

	if err := json.Unmarshal(dailyJSON, &dailyClicks); err != nil {
		return dto.UrlStatsResponse{}, err
	}

	if err := json.Unmarshal(agentsJSON, &topAgents); err != nil {
		return dto.UrlStatsResponse{}, err
	}

	return dto.UrlStatsResponse{
		ShortCode:     shortCode,
		TotalClicks:   totalClicks,
		ClicksPerDay:  dailyClicks,
		TopUserAgents: topAgents,
	}, nil
}

type rlogRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &rlogRepository{
		db: db,
	}
}
