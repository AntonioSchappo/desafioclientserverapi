package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		fmt.Println(req.Body)
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler a resposta: %v\n", err)
	}

	var data Cotacao
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
	}

	fileUtils(data.Bid)

}

func fileUtils(cotacao string) {
	file, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	novaCotacao := fmt.Sprintf("Dolar: %s\n", cotacao)
	if _, err := file.Write([]byte(novaCotacao)); err != nil {
		file.Close()
		log.Fatal(err)
	}
	if err := file.Close(); err != nil {
		log.Fatal(err)
	}
}
