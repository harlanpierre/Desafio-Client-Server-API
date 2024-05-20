package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	apiURL     = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	dbName     = "cotacoes.db"
	apiTimeout = 300 * time.Millisecond
	dbTimeout  = 10 * time.Millisecond
)

type Cotacao struct {
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

func main() {
	initDB()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /cotacao", handleCotacao)
	http.ListenAndServe(":8080", mux)

}

func handleCotacao(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cotacao, err := getCotacao(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao buscar cotação: %v", err)
		return
	}

	if err := saveCotacao(ctx, cotacao); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao salvar cotação no banco de dados: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cotacao.USDBRL.Bid); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao converter Data para Json: %v", err)
		return
	}

}

func getCotacao(ctx context.Context) (*Cotacao, error) {
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao realizar requisição: %v", err)
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao receber resposta da requisição: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	var data Cotacao
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao decodificar resposta: %v\n", err)
		return nil, err
	}

	return &data, err
}

func saveCotacao(ctx context.Context, cotacao *Cotacao) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao abrir DB: %v", err)
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx,
		"INSERT INTO cotacoes (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) "+
			" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		cotacao.USDBRL.Code, cotacao.USDBRL.Codein, cotacao.USDBRL.Name, cotacao.USDBRL.High, cotacao.USDBRL.Low,
		cotacao.USDBRL.VarBid, cotacao.USDBRL.PctChange, cotacao.USDBRL.Bid, cotacao.USDBRL.Ask,
		cotacao.USDBRL.Timestamp, cotacao.USDBRL.CreateDate)
	return err
}

func initDB() {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT,
		codein TEXT,
		name TEXT,
		high TEXT,
		low TEXT,
		varBid TEXT,
		pctChange TEXT,
		bid TEXT,
		ask TEXT,
		timestamp TEXT,
		create_date TEXT
	)`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar DB: %v", err)
	}
}
