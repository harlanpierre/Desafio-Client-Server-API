package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	serverURL  = "http://localhost:8080/cotacao"
	timeout    = 300 * time.Millisecond
	outputFile = "cotacao.txt"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao realizar request: %v\n", err)
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao recuperar resposta: %v\n", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Erro: status code %d", res.StatusCode)
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler o corpo da resposta: %v\n", err)
		return
	}

	if len(body) == 0 {
		fmt.Fprintf(os.Stderr, "Erro: corpo da resposta vazio\n")
		return
	}

	var data string
	if err = json.Unmarshal(body, &data); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer parse do JSON: %v\n", err)
		return
	}

	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar arquivo: %v\n", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar: %s", data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao escrever no arquivo: %v\n", err)
	}

	fmt.Printf("Cotação: %+v\n", data)
}
