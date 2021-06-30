package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type bnResp struct {
	Price float64 `json:"price,string"`
	Code  int64   `json:"code"`
}

type wallet map[string]float64

var DB = map[int64]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI("")
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		command := strings.Split(update.Message.Text, " ")
		switch command[0] {
		case "ADD":
			if len(command) != 3 {
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команда некорректна"))
				if err != nil {
					fmt.Println(err)
				}
			}
			amount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				if err != nil {
					fmt.Println(err)
				}
				continue
			}
			if _, ok := DB[update.Message.Chat.ID]; !ok {
				DB[update.Message.Chat.ID] = wallet{}
			}
			DB[update.Message.Chat.ID][command[1]] += amount
			balanceText := fmt.Sprintf("%s %f", command[1], DB[update.Message.Chat.ID][command[1]])
			_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, balanceText))
			if err != nil {
				fmt.Println(err)
			}
		case "SUB":
			if len(command) != 3 {
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команда некорректна"))
				if err != nil {
					fmt.Println(err)
				}
			}
			amount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				if err != nil {
					fmt.Println(err)
				}
				continue
			}
			if _, ok := DB[update.Message.Chat.ID]; !ok {
				continue
			}
			if DB[update.Message.Chat.ID][command[1]] < amount {
				mesg := fmt.Sprintf("Недостаточно средств [%s = %f]", command[1], DB[update.Message.Chat.ID][command[1]])
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, mesg))
				if err != nil {
					fmt.Println(err)
				}
				continue
			}
			DB[update.Message.Chat.ID][command[1]] -= amount
			balanceText := fmt.Sprintf("%s %f", command[1], DB[update.Message.Chat.ID][command[1]])
			_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, balanceText))
			if err != nil {
				fmt.Println(err)
			}
		case "DEL":
			if len(command) != 2 {
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команда некорректна"))
				if err != nil {
					fmt.Println(err)
				}
			}
			delete(DB[update.Message.Chat.ID], command[1])
		case "SHOW":
			msg := ""
			summ := 0.0
			rate, _ := getRateRUB()
			for key, value := range DB[update.Message.Chat.ID] {
				price, _ := getPrice(key)
				summ += value * price
				msg += fmt.Sprintf("%s: %f [$%.2f]\n", key, value, value*price)
			}
			msg += fmt.Sprintf("Total: $%.2f [%.2f RUB]\n", summ, summ*rate)
			_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
			if err != nil {
				fmt.Println(err)
			}
		default:
			_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команда не распознана"))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func getPrice(symbol string) (price float64, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", symbol))
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	var jsonResp bnResp
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return 0, err
	}
	if jsonResp.Code != 0 {
		return 0, errors.New("неверный символ")
	}
	price = jsonResp.Price

	return price, nil
}

func getRateRUB() (rate float64, err error) {
	resp, err := http.Get("https://api.binance.com/api/v3/ticker/price?symbol=USDTRUB")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	var jsonResp bnResp
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return 0, err
	}
	if jsonResp.Code != 0 {
		return 0, errors.New("Ошибка сайта курсов валют")
	}
	rate = jsonResp.Price

	return rate, nil
}
