package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type APIResponse struct {
	Usdbrl struct {
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

func getDolar() (*APIResponse, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r APIResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func saveToDB(data *APIResponse) error {
	db, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer db.Close()

	stmt, err := db.Prepare(`INSERT INTO cotacoes (cotacao) VALUES(?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()

	_, err = stmt.ExecContext(ctx, data.Usdbrl.Bid)
	if err != nil {
		return err
	}
	return nil
}

func getCotacao(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cotacao, err := getDolar()
	if err != nil {
		panic(err)
	}
	err = saveToDB(cotacao)
	if err != nil {
		panic(err)
	}

	type ResponseType struct {
		Cotacao string `json:"cotacao"`
	}
	response := ResponseType{cotacao.Usdbrl.Bid}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	w.Write([]byte(jsonResponse))
}

func createTable() {
	fmt.Println("Creating table")
	db, err := sql.Open("sqlite3", "./sqlite-database.db")
	if err != nil {
		fmt.Println("panic creating database")
		panic(err)
	}
	defer db.Close()
	stmt, err := db.Prepare(`CREATE TABLE cotacoes(
                                id INTEGER PRIMARY KEY AUTOINCREMENT,
                                cotacao VARCHAR(4)
                              );`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		panic(err)
	}
	fmt.Println("Table created")
}

func main() {
	os.Remove("sqlite-database.db")
	log.Println("creating sqlite-database.db...")
	file, err := os.Create("sqlite-database.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()

	log.Println("sqlite-database.db created")

	createTable()

	http.HandleFunc("/cotacao", getCotacao)
	http.ListenAndServe(":8080", nil)
}
