package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const UrlDolarExchange = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

const MillisecondTimeoutApi = 200 * time.Millisecond

const MillisecondTimeoutDb = 10 * time.Millisecond

type DollarExchangeRate struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type DollarExchangeRateResponse struct {
	Bid string `json:"bid"`
}

func main() {
	http.HandleFunc("/", GetDollarExchangeRateHandler)
	http.ListenAndServe(":8080", nil)
}

func GetDollarExchangeRateHandler(writer http.ResponseWriter, request *http.Request) {
	dollarExchangeRate, err := CurrentDollarExchangeRate()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	SaveCurrentDolar()
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	print(dollarExchangeRate.USDBRL.Bid)
	json.NewEncoder(writer).Encode(DollarExchangeRateResponse{Bid: dollarExchangeRate.USDBRL.Bid})
}

func SaveCurrentDolar() {
	db := InitDb()
	defer db.Close()
	dolar, err := CurrentDollarExchangeRate()
	if err != nil {
		log.Fatal(err)
	}
	err = InsertCurrentDolar(db, dolar)
	if err != nil {
		log.Fatal(err)
	}
}

func InitDb() *sql.DB {
	db, err := sql.Open("sqlite3", "dolar.db")
	if err != nil {
		log.Fatal(err)
	}
	db.Exec("CREATE TABLE IF NOT EXISTS dolar (id INTEGER PRIMARY KEY AUTOINCREMENT,bid DECIMAL(10, 2))")
	return db
}

func InsertCurrentDolar(db *sql.DB, dolar *DollarExchangeRate) error {
	ctx, cancel := context.WithTimeout(context.Background(), MillisecondTimeoutDb)
	defer cancel()
	statement, err := db.PrepareContext(ctx, "INSERT INTO dolar (bid) VALUES (?)")
	if err != nil {
		return err
	}
	_, err = statement.Exec(dolar.USDBRL.Bid)
	if err != nil {
		return err
	}
	return nil
}

func CurrentDollarExchangeRate() (*DollarExchangeRate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), MillisecondTimeoutApi)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, UrlDolarExchange, nil)
	if err != nil {
		log.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var dollarExchangeRate DollarExchangeRate
	err = json.NewDecoder(response.Body).Decode(&dollarExchangeRate)
	if err != nil {
		return nil, err
	}
	return &dollarExchangeRate, nil
}
