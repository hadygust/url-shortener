package dto

type UrlStatsResponse struct {
	ShortCode     string        `json:"short_code"`
	TotalClicks   int           `json:"total_clicks"`
	ClicksPerDay  []DailyClicks `json:"clicks_per_day"`
	TopUserAgents []UserAgent   `json:"top_user_agents"`
}

type DailyClicks struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type UserAgent struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
