package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Client is an HTTP client for Telegram Bot API.
type Client struct {
	Token      string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Telegram client.
func NewClient(token string) *Client {
	return &Client{
		Token:      token,
		BaseURL:    "https://api.telegram.org",
		HTTPClient: &http.Client{Timeout: 20 * time.Second},
	}
}

// 4. Получи информацию о боте (метод getMe).
func (c *Client) GetMe() (int, string, error) {
	reqURL := fmt.Sprintf("%s/bot%s/getMe", c.BaseURL, c.Token)
	return c.doGet(reqURL)
}

// 6. Ищи информацию о себе в ответе на метод getUpdates! Сохрани свой user_id и chat_id.
func (c *Client) GetUpdates() (int, string, error) {
	reqURL := fmt.Sprintf("%s/bot%s/getUpdates", c.BaseURL, c.Token)
	return c.doGet(reqURL)
}

// 7. А теперь отправь себе сообщение от имени бота (sendMessage).
func (c *Client) SendMessage(chatID, text string) (int, string, error) {
	reqURL := fmt.Sprintf("%s/bot%s/sendMessage", c.BaseURL, c.Token)
	payload, _ := json.Marshal(map[string]string{
		"chat_id": chatID,
		"text":    text,
	})

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(payload))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("\n[HTTP REQUEST]\n  Method: POST\n  URL:    %s/bot%s/sendMessage\n  Body:   %s\n",
		c.BaseURL, mask(c.Token), string(payload))

	return c.send(req)
}

// 8, 9, 10. Отправка файлов через sendPhoto с использованием multipart/form-data.
func (c *Client) SendPhoto(chatID, filePath string) (int, string, error) {
	reqURL := fmt.Sprintf("%s/bot%s/sendPhoto", c.BaseURL, c.Token)

	file, err := os.Open(filePath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	// Add chat_id field
	if err := writer.WriteField("chat_id", chatID); err != nil {
		return 0, "", err
	}

	// Add photo file field
	part, err := writer.CreateFormFile("photo", filepath.Base(filePath))
	if err != nil {
		return 0, "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return 0, "", err
	}

	if err := writer.Close(); err != nil {
		return 0, "", err
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, bodyBuf)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	fmt.Printf("\n[HTTP REQUEST]\n  Method: POST (multipart/form-data)\n  URL:    %s/bot%s/sendPhoto\n  File:   %s\n",
		c.BaseURL, mask(c.Token), filePath)

	return c.send(req)
}

// 11. Напоследок посмотрим, что бот может узнать о тебе (getUserProfilePhotos).
func (c *Client) GetUserProfilePhotos(userID string) (int, string, error) {
	reqURL := fmt.Sprintf("%s/bot%s/getUserProfilePhotos?user_id=%s", c.BaseURL, c.Token, userID)
	return c.doGet(reqURL)
}

func (c *Client) doGet(reqURL string) (int, string, error) {
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, "", err
	}

	fmt.Printf("\n[HTTP REQUEST]\n  Method: GET\n  URL:    %s\n", maskURL(reqURL))
	return c.send(req)
}

func (c *Client) send(req *http.Request) (int, string, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}

	return resp.StatusCode, string(respBody), nil
}

// RunPart3 executes all required Telegram Bot API tasks.
func RunPart3(token, chatID, userID string) {
	fmt.Println("=================================================================")
	fmt.Println("             PART 3: TELEGRAM BOT API")
	fmt.Println("=================================================================")

	if token == "" {
		fmt.Println("⚠️  TELEGRAM_BOT_TOKEN is not set in .env! Cannot perform calls.")
		fmt.Println("   1. Create a bot using @BotFather in Telegram.")
		fmt.Println("   2. Put your token in .env (TELEGRAM_BOT_TOKEN=...)")
		return
	}

	client := NewClient(token)

	// Step 4: getMe
	fmt.Println("\n--- [1/7] Step 4: getMe ---")
	status, body, err := client.GetMe()
	printResult("getMe", status, body, err)

	// Step 6: getUpdates
	fmt.Println("\n--- [2/7] Step 6: getUpdates (find your user_id and chat_id) ---")
	status, body, err = client.GetUpdates()
	printResult("getUpdates", status, body, err)

	if chatID == "" {
		fmt.Println("⚠️  TELEGRAM_CHAT_ID is not set in .env! Skip sendMessage & sendPhoto tests.")
		fmt.Println("   Look at getUpdates output above to find your chat_id and user_id, then set them in .env.")
		return
	}

	// Step 7: sendMessage
	fmt.Println("\n--- [3/7] Step 7: sendMessage ---")
	status, body, err = client.SendMessage(chatID, "Привет от Go-клиента BE1_HttpAuth!")
	printResult("sendMessage", status, body, err)

	// Step 8: sendPhoto (valid JPEG)
	fmt.Println("\n--- [4/7] Step 8: sendPhoto (with sample.jpg) ---")
	sampleJPG := findFile("sample.jpg", "src/sample.jpg")
	status, body, err = client.SendPhoto(chatID, sampleJPG)
	printResult("sendPhoto (JPEG)", status, body, err)

	// Step 9: sendPhoto (with PDF)
	fmt.Println("\n--- [5/7] Step 9: sendPhoto (with sample.pdf - Expected Error) ---")
	fmt.Println(">> Notice: The assignment asks: 'А что, если вместо фотографии прикрепить pdf-документ? Результат точно удивит!'")
	samplePDF := findFile("sample.pdf", "src/sample.pdf")
	status, body, err = client.SendPhoto(chatID, samplePDF)
	printResult("sendPhoto (PDF)", status, body, err)
	fmt.Printf("   >> Telegram API response code: %d (Expected: 400 Bad Request: PHOTO_INVALID_DIMENSIONS / IMAGE_PROCESS_FAILED)\n", status)

	// Step 10: sendPhoto (with DOCX)
	fmt.Println("\n--- [6/7] Step 10: sendPhoto (with sample.docx - Expected Error) ---")
	fmt.Println(">> Notice: The assignment asks: 'Теперь вместо pdf-документа попробуй прикрепить docx-файл. Какой будет ответ?'")
	sampleDOCX := findFile("sample.docx", "src/sample.docx")
	status, body, err = client.SendPhoto(chatID, sampleDOCX)
	printResult("sendPhoto (DOCX)", status, body, err)
	fmt.Printf("   >> Telegram API response code: %d (Expected: 400 Bad Request: IMAGE_PROCESS_FAILED)\n", status)

	// Step 11: getUserProfilePhotos
	if userID == "" {
		userID = chatID // in direct private chats, chat_id == user_id
	}
	fmt.Println("\n--- [7/7] Step 11: getUserProfilePhotos ---")
	status, body, err = client.GetUserProfilePhotos(userID)
	printResult("getUserProfilePhotos", status, body, err)
}

func printResult(action string, status int, body string, err error) {
	if err != nil {
		fmt.Printf("  Status: %d | Error: %v\n", status, err)
		return
	}
	fmt.Printf("  HTTP Status: %d\n", status)
	if len(body) > 300 {
		fmt.Printf("  Response:    %s... (truncated)\n", body[:300])
	} else {
		fmt.Printf("  Response:    %s\n", body)
	}
}

func findFile(candidates ...string) string {
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return candidates[0]
}

func mask(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func maskURL(u string) string {
	return u
}
