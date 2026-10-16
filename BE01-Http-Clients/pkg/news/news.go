package news

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Article represents a NewsAPI article item.
type Article struct {
	Source struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"source"`
	Author      string `json:"author"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	PublishedAt string `json:"publishedAt"`
}

// NewsResponse represents the standard JSON response from NewsAPI.
type NewsResponse struct {
	Status       string    `json:"status"`
	Code         string    `json:"code,omitempty"`
	Message      string    `json:"message,omitempty"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

// Client is a client for News API.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new News API client.
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		BaseURL:    "https://newsapi.org",
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// executeRequest performs an HTTP GET request with Bearer authorization header.
func (c *Client) executeRequest(endpoint string, params url.Values) (*NewsResponse, int, error) {
	reqURL := fmt.Sprintf("%s%s?%s", c.BaseURL, endpoint, params.Encode())

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Task requirement: "API KEY нужно прикладывать в качестве bearer token в заголовке авторизации!"
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	req.Header.Set("User-Agent", "BE1_HttpAuth-GoClient/1.0")

	fmt.Printf("\n[HTTP REQUEST]\n  Method: GET\n  URL:    %s\n  Header: Authorization: Bearer %s\n",
		reqURL, maskKey(c.APIKey))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	var newsResp NewsResponse
	if err := json.Unmarshal(bodyBytes, &newsResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to parse JSON response (%s): %w", string(bodyBytes), err)
	}

	return &newsResp, resp.StatusCode, nil
}

// 1. Необходимо получить все новости по теме linux.
func (c *Client) GetLinuxNews() (*NewsResponse, int, error) {
	params := url.Values{}
	params.Set("q", "linux")
	return c.executeRequest("/v2/everything", params)
}

// 2. Получить все новости по теме «Разработка» на русском языке за последние 15 дней.
func (c *Client) GetDevNewsRussianLast15Days() (*NewsResponse, int, error) {
	fromDate := time.Now().AddDate(0, 0, -15).Format("2006-01-02")
	params := url.Values{}
	params.Set("q", "Разработка")
	params.Set("language", "ru")
	params.Set("from", fromDate)
	return c.executeRequest("/v2/everything", params)
}

// 3. Получить все новости по теме linux на русском языке на 3 странице, на каждой из которых по 10 новостей.
func (c *Client) GetLinuxNewsRussianPage3() (*NewsResponse, int, error) {
	params := url.Values{}
	params.Set("q", "linux")
	params.Set("language", "ru")
	params.Set("pageSize", "10")
	params.Set("page", "3")
	return c.executeRequest("/v2/everything", params)
}

// 4. Получить все заголовки новостей для страны Россия по категории «наука» (science).
func (c *Client) GetScienceHeadlinesRussia() (*NewsResponse, int, error) {
	params := url.Values{}
	params.Set("country", "ru")
	params.Set("category", "science")
	return c.executeRequest("/v2/top-headlines", params)
}

// RunPart1 executes all 4 requests of Part 1.
func RunPart1(apiKey string) {
	fmt.Println("=================================================================")
	fmt.Println("             PART 1: NEWS API (Bearer Auth)")
	fmt.Println("=================================================================")

	client := NewClient(apiKey)

	tasks := []struct {
		title string
		fn    func() (*NewsResponse, int, error)
	}{
		{
			title: "Task 1: All news about Linux",
			fn:    client.GetLinuxNews,
		},
		{
			title: "Task 2: Development news in Russian for the last 15 days",
			fn:    client.GetDevNewsRussianLast15Days,
		},
		{
			title: "Task 3: Linux news in Russian on page 3 (10 per page)",
			fn:    client.GetLinuxNewsRussianPage3,
		},
		{
			title: "Task 4: Science headlines for Russia",
			fn:    client.GetScienceHeadlinesRussia,
		},
	}

	for i, task := range tasks {
		fmt.Printf("\n--- [%d/4] %s ---\n", i+1, task.title)
		resp, status, err := task.fn()
		if err != nil {
			fmt.Printf("Status: %d | Error: %v\n", status, err)
			continue
		}

		fmt.Printf("HTTP Status: %d OK\n", status)
		fmt.Printf("API Status:  %s\n", resp.Status)
		if resp.Status == "error" {
			fmt.Printf("API Error Code:    %s\n", resp.Code)
			fmt.Printf("API Error Message: %s\n", resp.Message)
		} else {
			fmt.Printf("Total Results: %d\n", resp.TotalResults)
			fmt.Printf("Articles returned in this page: %d\n", len(resp.Articles))
			for idx, a := range resp.Articles {
				if idx >= 3 {
					fmt.Printf("  ... and %d more articles\n", len(resp.Articles)-3)
					break
				}
				fmt.Printf("  [%d] %s (%s)\n", idx+1, a.Title, a.PublishedAt)
			}
		}
	}
}

func maskKey(k string) string {
	if k == "" {
		return "<NOT_SET>"
	}
	if len(k) <= 8 {
		return "****"
	}
	return k[:4] + "..." + k[len(k)-4:]
}
