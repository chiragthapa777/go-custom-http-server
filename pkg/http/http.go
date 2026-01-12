package http

import (
	"fmt"
	"io"
	"net"
)

type HttpServer struct {
	Port    int
	Address string
}

func (hs *HttpServer) Start() error {
	address := fmt.Sprintf("%s:%d", hs.Address, hs.Port)
	// step 1, 2 and 3 : socket -> bind -> listen
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	fmt.Println("Waiting for client...")

	// step 3 : waiting for client and accepting connection in blocking fashion
	tcpClientConn, err := listener.Accept()
	if err != nil {
		return err
	}

	// step 4 : read the data from client tcp conn
	bytes := make([]byte, 10000) // 10 kb byte array
	totalBytes, err := tcpClientConn.Read(bytes)
	if err != nil {
		if err == io.EOF {
			//client closed connection
			if err = tcpClientConn.Close(); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	body := string(bytes[:totalBytes])
	fmt.Printf("Received Data : %s \n total bytes : %d \n", body, totalBytes)

	// This is where HTTP parsing would happen

	resp := []byte("OK")

	// step 5 : write the data to client tcp conn
	totalBytes, err = tcpClientConn.Write(resp)
	if err != nil {
		return err
	}
	fmt.Printf("Sent Data : %s \n total bytes : %d \n", string(resp), totalBytes)

	// step 6: close the client tcp connection
	if err = tcpClientConn.Close(); err != nil {
		return err
	}
	if err = listener.Close(); err != nil {
		return err
	}
	return nil

}
