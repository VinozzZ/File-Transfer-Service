package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

const MAX_BUFFER_SIZE = 4 * 1024
const TOKEN_SIZE = 16
const SUCCESS_STATUS_CODE = "200\n"
const TIMEOUT_DURIATION = 5 * time.Minute

var senderStack = make(map[string]net.Conn)

func main() {
	url := getHostURL()
	listener, err := net.Listen("tcp", url)
	if err != nil {
		log.Println("Error starting server: ", err)
	}
	defer listener.Close()
	fmt.Println("Starting file server on", url)

	for {
		cleanUpConnectionInStack()
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error: ", err)
		}
		go handleConnection(conn)
	}
}

func getHostURL() string {
	port := ""

	if len(os.Args) > 1 {
		port = os.Args[1]
	} else {
		port = ":9021"
	}

	return "localhost" + port
}

func handleConnection(conn net.Conn) {
	// each connection is separate, we set a timer to be 10min after the sender send in the code
	// if within 10 mins there's no matching code send from a receiver, we close the connection
	// if the code matches, we ask the sender to upload the file and then send it to the receiver
	conn.SetDeadline(time.Now().Add(TIMEOUT_DURIATION))
	tokenBuffer := make([]byte, TOKEN_SIZE)
	conn.Read(tokenBuffer)
	token := string(tokenBuffer)
	if senderConn, exist := senderStack[token]; exist {
		//TODO: connection deadLine should be reset and start a longer one for transfering files
		sendFile(conn, senderConn, token)
	} else {
		senderStack[token] = conn
	}
}

func sendFile(conn net.Conn, senderConn net.Conn, token string) {
	defer conn.Close()
	defer senderConn.Close()

	senderStatusCode := []byte(SUCCESS_STATUS_CODE)
	senderConn.Write(senderStatusCode)

	fileBuffer := make([]byte, MAX_BUFFER_SIZE)
	for {
		n, senderErr := senderConn.Read(fileBuffer)
		if n != 0 {
			_, receiverErr := conn.Write(fileBuffer[:n])
			if receiverErr != nil {
				break
			}
		}
		if senderErr != nil {
			if senderErr == io.EOF {
				break
			}
			log.Println("Error: ", senderErr)
		}
	}
	delete(senderStack, token)
	return
}

func connIsClosed(conn net.Conn) bool {
	conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	oneByte := make([]byte, 1)
	_, err := conn.Read(oneByte)
	if err != nil {
		if err == io.EOF {
			conn.Close()
			conn = nil
			return true
		}
		log.Println("Error: ", err)
	}
	var zero time.Time
	conn.SetReadDeadline(zero)
	return false
}

func cleanUpConnectionInStack() {
	for key, previousConn := range senderStack {
		isClosed := connIsClosed(previousConn)
		if isClosed == true {
			delete(senderStack, key)
		}
	}
}
