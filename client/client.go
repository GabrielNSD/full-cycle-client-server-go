package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func main() {
	fmt.Println("main client")
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	type ResponseType struct {
		Cotacao string `json:"cotacao"`
	}
	var r ResponseType
	err = json.Unmarshal(body, &r)
	if err != nil {
		panic(err)
	}

	fmt.Println(r.Cotacao)
	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}

	resultado := fmt.Sprintf("Dólar: {%v}", r.Cotacao)
	tamanho, err := f.WriteString(resultado)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Arquivo criado com sucesso! Tamanho: %d bytes", tamanho)

	f.Close()
}
