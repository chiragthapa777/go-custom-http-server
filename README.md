# server flow

socket → bind → listen → accept → read → write → close

1. socket

   - socket is a kernel object represented by a file descriptor (fd is a number pointing to a kernel-managed socket/resource )
   - it represents a network endpoint managed by the os
   - at this stage it is not bound to ip or port
   - ask kernel to give a network endpoint
   - kernel returns a fd
   - kernel returns a fd

2. bind

   - attach socket to ip and port
   - kernel checks port availability and permission
   - ports below 1024 need root access
   - socket is now bound to an address

3. listen

   - here we start listening to the n/w socket
   - mark the socket as passive (means you tell os this socket will not initiate conn rather it will wait and listen for incoming connection)
   - now kernel will:
     - creates a accept queue
     - handle tcp 3 way handshake for your socket (syc -> sync-ack -> ack)

Note:

    - in golang we cannot manually call socket or bind
    - net.Listen -> for server and net.Dial -> client handle this internally
    - after socket + bind this is not a tcp connection
    - this is only a bound socket, not connected to any client
    - a tcp connection exists only after accept (server) and connect (client)

4. accept

   - it accepts a incoming client
   - blocks until client completes tcp handshake
   - returns new client socket

Note:

    - listening server != client socket
    - one listening server can have many client socket

5. read

   - read byte from client (read from tcp receive buffer)
   - tcp guarantees (ordered bytes and no message boundaries)

6. write

   - send byte to client (write byte to tcp send buffer)
   - kernel: splits the packet, handles retransmission and ensures delivery

7. close

   - close client connection
   - kernel: send FIN signal to client, and frees resources later

Minimal Server:

```
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

```

Minimal Client:

```
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

```
