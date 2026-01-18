package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Response struct {
	StatusCode int
	Headers    map[string]string
	StringBody string
}

func NewResponse() *Response {
	return &Response{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
		StringBody: "OK",
	}
}

func (r *Response) internalServerError() {
	r.StatusCode = 500
	r.StringBody = "Internal Server Error"
}

func (r *Response) notFound() {
	r.StatusCode = 404
	r.StringBody = "Not Found"
}

func (r *Response) badRequest() {
	r.StatusCode = 400
	r.StringBody = "Bad Request"
}

func (r *Response) parseHttpResponse() []byte {
	r.Headers["Server"] = "Go-TCP-HTTP"
	r.Headers["Content-Length"] = strconv.Itoa(len(r.StringBody))
	r.Headers["Connection"] = "close"

	var b strings.Builder

	b.WriteString(fmt.Sprintf(
		"HTTP/1.1 %d %s\r\n",
		r.StatusCode,
		http.StatusText(r.StatusCode),
	))

	for k, v := range r.Headers {
		b.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	b.WriteString("\r\n")
	b.WriteString(r.StringBody)

	return []byte(b.String())
}

// ==========================
// Request
// ==========================

type HttpMessage struct {
	Method   string
	Path     string
	Protocol string
	Headers  map[string]string
	Host     string
	RawBody  string
}

func ParseHttpRequest(n int, buf []byte) (*HttpMessage, error) {
	data := string(buf[:n])

	parts := strings.SplitN(data, "\r\n\r\n", 2)
	if len(parts) == 0 {
		return nil, errors.New("invalid http request")
	}

	lines := strings.Split(parts[0], "\r\n")
	if len(lines) == 0 {
		return nil, errors.New("invalid request line")
	}

	startLine := strings.Split(lines[0], " ")
	if len(startLine) != 3 {
		return nil, errors.New("invalid start line")
	}

	hm := &HttpMessage{
		Method:   startLine[0],
		Path:     startLine[1],
		Protocol: startLine[2],
		Headers:  make(map[string]string),
	}

	for _, h := range lines[1:] {
		if h == "" {
			continue
		}
		kv := strings.SplitN(h, ":", 2)
		if len(kv) != 2 {
			return nil, errors.New("invalid header")
		}
		hm.Headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}

	host, ok := hm.Headers["Host"]
	if !ok {
		return nil, errors.New("missing host header")
	}
	hm.Host = host

	if len(parts) == 2 {
		hm.RawBody = parts[1]
	}

	return hm, nil
}

// ==========================
// Handler Context
// ==========================

type HandlerContext struct {
	Body     string
	Headers  map[string]string
	Context  context.Context
	Method   string
	Path     string
	Response *Response
}

// ==========================
// Server
// ==========================

type HttpServer struct {
	Address  string
	Port     int
	Handlers map[string]func(*HandlerContext) error
}

func (hs *HttpServer) StartMultiThreadedServer() error {
	addr := fmt.Sprintf("%s:%d", hs.Address, hs.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Println("Listening on", addr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("Graceful shutdown...")
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			fmt.Println("Accept error:", err)
			continue
		}
		go hs.HandleRequest(conn)
	}
}

func (hs *HttpServer) HandleRequest(conn net.Conn) {
	defer conn.Close()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered:", r)
			debug.PrintStack()
		}
	}()

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	buf := make([]byte, 8192)
	n, err := conn.Read(buf)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Read error:", err)
		}
		return
	}

	req, err := ParseHttpRequest(n, buf)
	resp := NewResponse()

	if err != nil {
		resp.badRequest()
		conn.Write(resp.parseHttpResponse())
		return
	}

	key := req.Path + "_" + req.Method
	handler, ok := hs.Handlers[key]
	if !ok {
		resp.notFound()
		conn.Write(resp.parseHttpResponse())
		return
	}

	err = handler(&HandlerContext{
		Body:     req.RawBody,
		Headers:  req.Headers,
		Context:  context.Background(),
		Method:   req.Method,
		Path:     req.Path,
		Response: resp,
	})

	if err != nil {
		resp.internalServerError()
	}

	conn.Write(resp.parseHttpResponse())
}
