package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func callApi(url string) *http.Response {
	req, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Erro ao buscar CEP - %v] %v\n", url, err)
		return nil
	}
	return req
}

func apiViaCep(cep string, cViaCep chan<- *http.Response) {
	cViaCep <- callApi("http://viacep.com.br/ws/" + cep + "/json/")
}

func apiCdn(cep string, cCdn chan<- *http.Response) {
	cep = cep[:5] + "-" + cep[5:]
	cCdn <- callApi("https://cdn.apicep.com/file/apicep/" + cep + ".json")
}

func getRequestInfo(req *http.Response) {
	fmt.Println("Api Response:", req.Request.URL.Host)
	defer req.Body.Close()

	res, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Erro buscar dados do CEP] %v\n", err)
	}

	var mEndereco interface{}
	err = json.Unmarshal(res, &mEndereco)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Erro ao executar Unmarshal] %v\n", err)
	}

	endereco, err := json.Marshal(mEndereco)
	if fmt.Println(string(endereco)); err != nil {
		fmt.Fprintf(os.Stderr, "[Erro ao executar Marshal para print] %v\n", err)
	}
}

func main() {
	cViaCep := make(chan *http.Response)
	cCdn := make(chan *http.Response)

	for _, cep := range os.Args[1:] {
		go apiViaCep(cep, cViaCep)
		go apiCdn(cep, cCdn)

		select {
		case reqViaCep := <-cViaCep:
			getRequestInfo(reqViaCep)
		case reqCdn := <-cCdn:
			getRequestInfo(reqCdn)
		case <-time.After(time.Second):
			fmt.Println("Api timeout")
		}
	}
}
