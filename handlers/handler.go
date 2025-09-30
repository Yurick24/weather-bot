package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/Yurick24/weather-bot/clients/openweather"
	"github.com/Yurick24/weather-bot/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot      *tgbotapi.BotAPI
	owClient *openweather.OpenWeatherClient
	userRepo userRepository
}

type userRepository interface {
	GetUserCity(ctx context.Context, userID int64) (string, error)
	CreateUser(ctx context.Context, userID int64) error
	UpdateCity(ctx context.Context, userID int64, city string) error
	GetUser(ctx context.Context, userID int64) (*models.User, error)
}

func New(bot *tgbotapi.BotAPI, owClient *openweather.OpenWeatherClient, userRepo userRepository) *Handler {
	return &Handler{
		bot:      bot,
		owClient: owClient,
		userRepo: userRepo,
	}
}

func (h *Handler) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.bot.GetUpdatesChan(u)

	for update := range updates {
		h.handleUpdate(update)
	}
}

func (h *Handler) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	ctx := context.Background()

	if update.Message.IsCommand() {
		err := h.ensureUserExists(ctx, update)
		if err != nil {
			log.Println("error ensure user: ", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка")
			msg.ReplyToMessageID = update.Message.MessageID
			h.bot.Send(msg)
			return
		}

		switch update.Message.Command() {
		case "start":
			h.handleStart(update)
			h.handleHelp(update)
		case "help":
			h.handleHelp(update)
		case "city":
			h.handleSetCity(ctx, update)
			return
		case "weather":
			h.handleSendWeather(ctx, update)
			return
		default:
			h.handleUnknownCommandDefault(update)
			return
		}
	}
}

func (h *Handler) handleSetCity(ctx context.Context, update tgbotapi.Update) {
	city := update.Message.CommandArguments()
	err := h.userRepo.UpdateCity(ctx, update.Message.From.ID, city)
	if err != nil {
		log.Println("error update city: ", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка")
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf("Город %s сохранен!", city),
	)
	msg.ReplyToMessageID = update.Message.MessageID
	h.bot.Send(msg)
}

func (h *Handler) handleSendWeather(ctx context.Context, update tgbotapi.Update) {
	city, err := h.userRepo.GetUserCity(ctx, update.Message.From.ID)

	if err != nil {
		log.Println("error update city: ", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка")
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	if city == "" {
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Сначала введите свой город с помощью команды /city",
		)
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	coordinates, err := h.owClient.Coordinates(city)
	if err != nil {
		log.Printf("error get coordinates in handler: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Мы не смогли получить координаты этой местности(")
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	weather, err := h.owClient.Weather(coordinates.Lat, coordinates.Lon)
	if err != nil {
		log.Printf("error get weather in handler: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Мы не смогли получить погоду в этой местности(")
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	cityName := weather.NameCity
	temp := fmt.Sprintf("Температура: %s°C", strconv.FormatFloat(weather.Temp, 'f', 1, 64))
	feelsTemp := fmt.Sprintf("Ощущуается как: %s°C", strconv.FormatFloat(weather.FeelsLike, 'f', 1, 64))
	description := fmt.Sprintf("Описание: %s", weather.Description)
	precipitation := fmt.Sprintf("Осадки: %s мм/ч", strconv.FormatFloat(weather.Precipitation, 'f', 2, 64))
	wind := fmt.Sprintf("Ветер: %s м/с, порыв: %s м/с", strconv.FormatFloat(weather.WindSpeed, 'f', 2, 64), strconv.FormatFloat(weather.WindGust, 'f', 2, 64))
	pressure := fmt.Sprintf("Давление: %s мм рт. ст.", strconv.FormatFloat(float64(weather.GrndLevel)*0.750063755419211, 'f', 2, 64))
	humidity := fmt.Sprintf("Влажность: %d%%", weather.Humidity)
	visibility := fmt.Sprintf("Видимость: %d метров", weather.Visibility)
	clouds := fmt.Sprintf("Облака: %d%%", weather.Clouds)

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf(
			"%s \n%s \n%s \n%s \n%s \n%s \n%s \n%s \n%s \n%s \n",
			cityName,
			temp,
			feelsTemp,
			description,
			precipitation,
			wind,
			pressure,
			humidity,
			visibility,
			clouds,
		),
	)
	msg.ReplyToMessageID = update.Message.MessageID

	h.bot.Send(msg)
}

func (h *Handler) handleUnknownCommandDefault(update tgbotapi.Update) {
	log.Printf("Unknown comand [%s] %s", update.Message.From.UserName, update.Message.Text)
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"Такая команда не доступна",
	)
	msg.ReplyToMessageID = update.Message.MessageID
	h.bot.Send(msg)
}

func (h *Handler) handleStart(update tgbotapi.Update) {

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf("Добро пожаловать, %s!", update.Message.From.UserName),
	)

	h.bot.Send(msg)
}

func (h *Handler) handleHelp(update tgbotapi.Update) {
	comCity := "Установите свой город с помощью команды /city номер_города"
	comWeather := "Чтобы узнать погоду в выбранном вами городе введите /weather"
	comHelp := "Чтобы еще раз ознакомиться с данной информацией введите команду /help"

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf(
			"%s \n\n%s \n\n%s",
			comCity,
			comWeather,
			comHelp,
		),
	)

	h.bot.Send(msg)
}

func (h *Handler) ensureUserExists(ctx context.Context, update tgbotapi.Update) error {
	user, err := h.userRepo.GetUser(ctx, update.Message.From.ID)
	if err != nil {
		return fmt.Errorf("error get user: %w", err)
	}

	if user == nil {
		err := h.userRepo.CreateUser(ctx, update.Message.From.ID)
		if err != nil {
			return fmt.Errorf("error create user: %w", err)
		}
	}

	return nil
}
