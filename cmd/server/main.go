package main

import (
	"encoding/json"

	"github.com/chiragthapa777/go-custom-http-server/pkg/http"
)

func main() {
	server := &http.HttpServer{
		Port:    3030,
		Address: "localhost",
		Handlers: map[string]func(handlerCtx *http.HandlerContext) error{
			"/_GET": func(handlerCtx *http.HandlerContext) error {
				handlerCtx.Response.StatusCode = 200
				handlerCtx.Response.Headers = map[string]string{
					"Content-Type": "application/json",
				}
				data := map[string]any{
					"data": "hello world",
				}
				jsonData, err := json.Marshal(data)
				if err != nil {
					return err
				}
				handlerCtx.Response.StringBody = string(jsonData)
				return nil
			},
		},
	}
	err := server.StartMultiThreadedServer()
	if err != nil {
		panic(err)
	}
}
