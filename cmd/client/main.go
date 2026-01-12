package main

import (
	"fmt"
	"net"
)

// socket() -> connect() -> write() -> read() -> close()

func main() {
	makeTcpRequest("localhost:3030")
}

func makeTcpRequest(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}

	n, err := conn.Write([]byte("this data is from client"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Total data written : %d\n", n)

	bytes := make([]byte, 10000)

	n, err = conn.Read(bytes)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Data from server : %s \n Total data written : %d\n", string(bytes), n)

	if err := conn.Close(); err != nil {
		panic(err)
	}
}
