package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

const create string = `
		CREATE TABLE IF NOT EXISTS cotacoes (
		id VARCHAR(50) NOT NULL PRIMARY KEY,
		time DATETIME NOT NULL,
		cotacao TEXT
		);`

type CotacaoDTO struct {
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

var db *sql.DB

func main() {
	var err error

	db, err = sql.Open("mysql", "root:root@tcp(localhost:3306)/goexpert")

	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(create)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/cotacao", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cotacao, err := BuscaCotacao(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(cotacao.USDBRL)
}

func BuscaCotacao(ctx context.Context) (CotacaoDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer res.Body.Close()

	var c CotacaoDTO
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return c, err
	}

	err = json.Unmarshal(body, &c)
	if err != nil {
		fmt.Println(err)
		return c, err
	}

	insertCotacao(c, ctx)

	return c, nil
}

func insertCotacao(cotacao CotacaoDTO, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, "INSERT INTO cotacoes(id, time, cotacao) VALUES(?,?,?);", uuid.New().String(), cotacao.USDBRL.CreateDate, cotacao.USDBRL.Bid)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

//O server.go deverá consumir a API contendo o câmbio de Dólar e Real no endereço:
// https://economia.awesomeapi.com.br/json/last/USD-BRL e em seguida deverá retornar no formato JSON o resultado para o cliente.

//Usando o package "context", o server.go deverá registrar no banco de dados SQLite cada cotação recebida,
//sendo que o timeout máximo para chamar a API de cotação do dólar deverá ser de 200ms e
//o timeout máximo para conseguir persistir os dados no banco deverá ser de 10ms.

//O endpoint necessário gerado pelo server.go para este desafio será: /cotacao e a porta a ser utilizada pelo servidor HTTP será a 8080.
