package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config holds environment configurations for the BE1_HttpAuth tasks.
type Config struct {
	NewsAPIKey       string
	TelegramBotToken string
	TelegramChatID   string
	TelegramUserID   string
	HHVacancyID      string
	HHEmployerID     string
}

// LoadConfig loads configuration from .env file (if present) and environment variables.
func LoadConfig() *Config {
	// Try loading from .env in current directory, src/.env, or parent directory
	envFiles := []string{".env", "src/.env", "../.env"}
	for _, file := range envFiles {
		if _, err := os.Stat(file); err == nil {
			loadEnvFile(file)
			break
		}
	}

	cfg := &Config{
		NewsAPIKey:       getEnv("NEWS_API_KEY", ""),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:   getEnv("TELEGRAM_CHAT_ID", ""),
		TelegramUserID:   getEnv("TELEGRAM_USER_ID", ""),
		HHVacancyID:      getEnv("HH_VACANCY_ID", "138003606"),
		HHEmployerID:     getEnv("HH_EMPLOYER_ID", "1455"),
	}

	return cfg
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, "\"'`")
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

// PrintStatus prints the loaded config status (masking sensitive tokens).
func (c *Config) PrintStatus() {
	fmt.Println("=== Current Configuration ===")
	fmt.Printf("  • NEWS_API_KEY:        %s\n", mask(c.NewsAPIKey))
	fmt.Printf("  • TELEGRAM_BOT_TOKEN:  %s\n", mask(c.TelegramBotToken))
	fmt.Printf("  • TELEGRAM_CHAT_ID:    %s\n", display(c.TelegramChatID))
	fmt.Printf("  • TELEGRAM_USER_ID:    %s\n", display(c.TelegramUserID))
	fmt.Printf("  • HH_VACANCY_ID:       %s\n", display(c.HHVacancyID))
	fmt.Printf("  • HH_EMPLOYER_ID:      %s\n", display(c.HHEmployerID))
	fmt.Println("=============================")
}

func mask(secret string) string {
	if secret == "" {
		return "<NOT SET> (add to .env)"
	}
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}

func display(val string) string {
	if val == "" {
		return "<NOT SET>"
	}
	return val
}
