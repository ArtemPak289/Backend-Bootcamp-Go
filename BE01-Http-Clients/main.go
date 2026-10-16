package main

import (
	"fmt"
	"os"
	"strings"

	"be1-httpauth/pkg/config"
	"be1-httpauth/pkg/hh"
	"be1-httpauth/pkg/news"
	"be1-httpauth/pkg/telegram"
)

const banner = `
===================================================================
     BE1_HttpAuth - School 21 Public HTTP APIs & Auth Runner
===================================================================
`

func printUsage() {
	fmt.Println(banner)
	fmt.Println("Usage: go run . [command]")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  all          Run all parts sequentially (Part 1, Part 2, Part 3)")
	fmt.Println("  news|part1   Run Part 1: NewsAPI (Bearer token auth, 4 queries)")
	fmt.Println("  hh|part2     Run Part 2: HH.ru developer tools (/shards/vacancy/...)")
	fmt.Println("  tg|part3     Run Part 3: Telegram Bot API (7+ requests, file tests)")
	fmt.Println("  status       Print current configuration (.env) status")
	fmt.Println("  help         Show this help screen")
	fmt.Println()
	fmt.Println("Note: Configure credentials in .env file (see .env.example).")
	fmt.Println("      Postman collections are also generated in src/ for peer review.")
}

func main() {
	cfg := config.LoadConfig()

	cmd := "all"
	if len(os.Args) > 1 {
		cmd = strings.ToLower(strings.TrimSpace(os.Args[1]))
	}

	switch cmd {
	case "help", "-h", "--help":
		printUsage()
		return
	case "status", "config":
		fmt.Println(banner)
		cfg.PrintStatus()
		return
	case "news", "part1":
		fmt.Println(banner)
		cfg.PrintStatus()
		news.RunPart1(cfg.NewsAPIKey)
	case "hh", "part2":
		fmt.Println(banner)
		cfg.PrintStatus()
		hh.RunPart2(cfg.HHVacancyID, cfg.HHEmployerID)
	case "tg", "telegram", "part3":
		fmt.Println(banner)
		cfg.PrintStatus()
		telegram.RunPart3(cfg.TelegramBotToken, cfg.TelegramChatID, cfg.TelegramUserID)
	case "all":
		fmt.Println(banner)
		cfg.PrintStatus()

		fmt.Println("\n>>> [1/3] Running Part 1 (News API)...")
		news.RunPart1(cfg.NewsAPIKey)

		fmt.Println("\n>>> [2/3] Running Part 2 (Developer Tools - hh.ru shards)...")
		hh.RunPart2(cfg.HHVacancyID, cfg.HHEmployerID)

		fmt.Println("\n>>> [3/3] Running Part 3 (Telegram Bot API)...")
		telegram.RunPart3(cfg.TelegramBotToken, cfg.TelegramChatID, cfg.TelegramUserID)

		fmt.Println("\n===================================================================")
		fmt.Println(" All requested parts completed!")
		fmt.Println(" Postman collections are saved in src/ ready for School 21 evaluation:")
		fmt.Println("   • src/Part1_NewsAPI.postman_collection.json")
		fmt.Println("   • src/Part2_HH.postman_collection.json")
		fmt.Println("   • src/Part3_TelegramBot.postman_collection.json")
		fmt.Println("   • src/BE1_HttpAuth.postman_collection.json (Unified)")
		fmt.Println("===================================================================")
	default:
		fmt.Printf("Unknown command '%s'\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
