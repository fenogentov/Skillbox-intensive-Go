package main

import "sync"

type wallet map[string]float64

type db struct {
	accaunst map[int64]wallet
	mu       sync.Mutex
}

var DB = db{
	accaunst: map[int64]wallet{},
}

func (db *db) addCoin(id int64, currency string, amount float64) {
	db.mu.Lock()
	defer db.mu.Unlock()

	DB.accaunst[id][currency] += amount
}

func (db *db) subCoin(id int64, currency string, amount float64) {
	db.mu.Lock()
	defer db.mu.Unlock()

	DB.accaunst[id][currency] -= amount
}

func (db *db) delCoin(id int64, currency string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(DB.accaunst[id], currency)
}
