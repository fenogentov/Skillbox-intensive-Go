package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type bnResp struct {
	Price float64 `json:"price,string"`
	Code  int64   `json:"code"`
}

func getExchangeRate(from, to string, amount float64, ch chan float64) {
	defer close(ch)
	cli := &http.Client{}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s%s", from, to), nil) // OK
	if err != nil {
		return
	}
	resp, err := cli.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var jsonResp bnResp
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}
	if jsonResp.Code != 0 {
		return
	}
	ch <- jsonResp.Price * amount
}

func addHandler(idChat int64, command []string) (answer string) {
	if len(command) != 3 {
		return msgIncorrectCommand
	}

	am := strings.ReplaceAll(command[2], ",", ".")
	amount, err := strconv.ParseFloat(am, 64)
	if err != nil {
		return msgIncorrectAmount
	}

	if _, ok := DB.accaunst[idChat]; !ok {
		DB.accaunst[idChat] = wallet{}
	}

	DB.addCoin(idChat, command[1], amount)

	chanUSD := make(chan float64)
	chanRUB := make(chan float64)

	go getExchangeRate(command[1], "USDT", DB.accaunst[idChat][command[1]], chanUSD)
	go getExchangeRate(command[1], "RUB", DB.accaunst[idChat][command[1]], chanRUB)

	usd := <-chanUSD
	rub := <-chanRUB

	return fmt.Sprintf(msgBalanceCurrency, command[1], DB.accaunst[idChat][command[1]], usd, rub)
}

func subHandler(idChat int64, command []string) (answer string) {
	if len(command) != 3 {
		return msgIncorrectCommand
	}

	if _, ok := DB.accaunst[idChat][command[1]]; !ok {
		return fmt.Sprintf(msgNoCurrency, command[1])
	}

	am := strings.ReplaceAll(command[2], ",", ".")
	amount, err := strconv.ParseFloat(am, 64)
	if err != nil {
		return msgIncorrectAmount
	}

	if DB.accaunst[idChat][command[1]] < amount {
		return msgInsufficientFunds
	}

	DB.subCoin(idChat, command[1], amount)

	chanUSD := make(chan float64)
	chanRUB := make(chan float64)

	go getExchangeRate(command[1], "USDT", DB.accaunst[idChat][command[1]], chanUSD)
	go getExchangeRate(command[1], "RUB", DB.accaunst[idChat][command[1]], chanRUB)

	usd := <-chanUSD
	rub := <-chanRUB

	return fmt.Sprintf(msgBalanceCurrency, command[1], DB.accaunst[idChat][command[1]], usd, rub)
}

func delHandler(idChat int64, command []string) (answer string) {
	if len(command) != 2 {
		return msgIncorrectCommand
	}
	if _, ok := DB.accaunst[idChat][command[1]]; !ok {
		return fmt.Sprintf(msgNoCurrency, command[1])
	}

	DB.delCoin(idChat, command[1])
	return fmt.Sprintf(msgDelCurrency, command[1])
}

func statusHandler(idChat int64) (answer string) {
	answer = ""
	summRUB := 0.0
	summUSD := 0.0

	for key, value := range DB.accaunst[idChat] {
		chanUSD := make(chan float64)
		chanRUB := make(chan float64)
		go getExchangeRate(key, "USDT", value, chanUSD)
		go getExchangeRate(key, "RUB", value, chanRUB)
		answer += fmt.Sprintf("%s: $%.2f\n", key, value)
		summUSD += <-chanUSD
		summRUB += <-chanRUB
	}

	answer += fmt.Sprintf("Total: %.2f USD, %.2f RUB\n", summUSD, summRUB)
	return answer
}

func handler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	c := strings.ToUpper(update.Message.Text)
	command := strings.Split(c, " ")
	switch command[0] {
	case "HELP":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msgHelp))
	case "ADD":
		answer := addHandler(update.Message.Chat.ID, command)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, answer))
	case "SUB":
		answer := subHandler(update.Message.Chat.ID, command)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, answer))
	case "DEL":
		answer := delHandler(update.Message.Chat.ID, command)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, answer))
	case "STATUS":
		answer := statusHandler(update.Message.Chat.ID)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, answer))
	default:
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команда не распознана"))
		if err != nil {
			fmt.Println(err)
		}
	}
}
