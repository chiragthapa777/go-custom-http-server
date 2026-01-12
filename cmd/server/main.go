package main

import "github.com/chiragthapa777/go-custom-http-server/pkg/http"

func main() {
	server := &http.HttpServer{
		Port:    3030,
		Address: "localhost",
	}
	err := server.Start()
	if err != nil {
		panic(err)
	}
}
