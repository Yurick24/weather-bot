package main

import (
	"context"
	"log"
	"os"

	"github.com/Yurick24/weather-bot/clients/openweather"
	"github.com/Yurick24/weather-bot/handlers"
	"github.com/Yurick24/weather-bot/repo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conn, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("data base connection error: ", err)
	}
	defer conn.Close()

	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatal("data base ping error: ", err)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	owClient := openweather.New(os.Getenv("OPENWEATHERAPI_KEY"))

	userRepo := repo.New(conn)

	botHandler := handlers.New(bot, owClient, userRepo)

	botHandler.Start()
}
