package hh

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for hh.ru internal shard endpoints.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new HH API client.
func NewClient() *Client {
	return &Client{
		BaseURL:    "https://hh.ru",
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// GetContactInfo fetches vacancy contact information via /shards/vacancy/contact_info.
func (c *Client) GetContactInfo(vacancyID, employerID string) (int, string, error) {
	reqURL := fmt.Sprintf("%s/shards/vacancy/contact_info?id=%s&employer_id=%s", c.BaseURL, vacancyID, employerID)
	return c.doRequest(http.MethodGet, reqURL, nil, vacancyID)
}

// GetPhone fetches vacancy phone number via /shards/vacancy/phone.
func (c *Client) GetPhone(vacancyID string, index int) (int, string, error) {
	reqURL := fmt.Sprintf("%s/shards/vacancy/phone?id=%s&index=%d", c.BaseURL, vacancyID, index)
	return c.doRequest(http.MethodGet, reqURL, nil, vacancyID)
}

// RegisterShowContactsAttempt sends POST to /shards/vacancy/register_show_contacts_attempt.
func (c *Client) RegisterShowContactsAttempt(vacancyID string) (int, string, error) {
	reqURL := fmt.Sprintf("%s/shards/vacancy/register_show_contacts_attempt", c.BaseURL)
	payload := []byte(fmt.Sprintf(`{"vacancyId": %s}`, vacancyID))
	return c.doRequest(http.MethodPost, reqURL, payload, vacancyID)
}

func (c *Client) doRequest(method, reqURL string, body []byte, vacancyID string) (int, string, error) {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return 0, "", err
	}

	// Browser emulation headers as required for reverse-engineered internal APIs
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", fmt.Sprintf("https://hh.ru/vacancy/%s", vacancyID))
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	fmt.Printf("\n[HTTP REQUEST]\n  Method: %s\n  URL:    %s\n", method, reqURL)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}

	snippet := string(respBody)
	if len(snippet) > 300 {
		snippet = snippet[:300] + "... (truncated)"
	}

	return resp.StatusCode, snippet, nil
}

// RunPart2 executes the reverse-engineered HH requests for Part 2.
func RunPart2(vacancyID, employerID string) {
	fmt.Println("=================================================================")
	fmt.Println("       PART 2: DEVELOPER TOOLS (hh.ru /shards/vacancy/...)")
	fmt.Println("=================================================================")
	fmt.Println("Subject requirement: 'Полученный в результате запрос должен начинаться с https://hh.ru/shards/vacancy/...'")
	fmt.Println("In Chrome DevTools -> Network -> Fetch/XHR, clicking contacts sends:")
	fmt.Println("  1) GET /shards/vacancy/contact_info")
	fmt.Println("  2) GET /shards/vacancy/phone")
	fmt.Println("  3) POST /shards/vacancy/register_show_contacts_attempt")

	if vacancyID == "" {
		vacancyID = "138003606"
	}
	if employerID == "" {
		employerID = "1455"
	}

	client := NewClient()

	fmt.Printf("\n--- [1/3] Testing GET /shards/vacancy/contact_info ---\n")
	status, body, err := client.GetContactInfo(vacancyID, employerID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("HTTP Status: %d\nResponse:    %s\n", status, body)
	}

	fmt.Printf("\n--- [2/3] Testing GET /shards/vacancy/phone ---\n")
	status, body, err = client.GetPhone(vacancyID, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("HTTP Status: %d\nResponse:    %s\n", status, body)
	}

	fmt.Printf("\n--- [3/3] Testing POST /shards/vacancy/register_show_contacts_attempt ---\n")
	status, body, err = client.RegisterShowContactsAttempt(vacancyID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("HTTP Status: %d\nResponse:    %s\n", status, body)
	}
}
