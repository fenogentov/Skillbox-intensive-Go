package main

const (
	msgInsufficientFunds = "Недостаточно средств для данной операции"
	msgIncorrectCommand  = "Некорректная команда"
	msgIncorrectAmount   = "Количество не распознано"
	msgBalanceCurrency   = "Баланс валюты %s = %f [%.2f USD, %.f2 RUB]"
	msgNoCurrency        = "Валюта %s отсутствует в кошельке"
	msgDelCurrency       = "Валюта %s удалена из кошелька"
	msgHelp              = `ADD <symbol> <amount>
	SUB <symbol> <amount>
	DEL <symbol>
	STATUS`
)
