package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const MAX_BUFFER_SIZE = 4 * 1024
const TOKEN_SIZE = 8
const SUCCESS_STATUS_CODE = "200"
const FILE_DATA_BUFFER_SIZE = 1024

func main() {
	fileName, relayURL := getUserInput()
	filePath := getAbsFilePath(fileName)
	if !isFileExists(filePath) {
		fmt.Println("Please check if your file exists under current directory and try again")
	}

	conn, err := net.Dial("tcp", relayURL)
	if err != nil {
		log.Println("Error: ", err)
	}

	token, err := GenerateRandomToken(TOKEN_SIZE)
	if err != nil {
		log.Println("Error: ", err)
	}
	fmt.Println("Token for receiver: ", token)
	tokenBuffer := []byte(token)
	conn.Write(tokenBuffer)

	reader := bufio.NewReader(conn)
	messageBuffer := make([]byte, 4)
	for {
		_, err := reader.Read(messageBuffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("Error: ", err)
		}
		message := string(messageBuffer)
		fmt.Println("Message Received:", message)
		if strings.Trim(message, "\n") == SUCCESS_STATUS_CODE {
			err := sendFile(conn, filePath)
			if err != nil {
				log.Println("Error: ", err)
			}
			break
		}
	}
}

func getUserInput() (string, string) {
	serverURL := ""
	fileName := ""
	if len(os.Args) > 2 {
		serverURL = os.Args[1]
		fileName = os.Args[2]
	} else {
		fmt.Println("Please provide a valid server URL and a file name exists in current directory")
		os.Exit(1)
	}

	return fileName, serverURL
}

func getAbsFilePath(fileName string) string {
	filePath, err := filepath.Abs(fileName)
	if err != nil {
		log.Println("Error", err)
	}

	return filePath
}

func sendFile(conn net.Conn, filePath string) error {
	defer conn.Close()

	file, err := os.Open(filePath)
	if err != nil {
		conn.Close()
		return err
	}
	defer file.Close()

	fileData, err := file.Stat()
	if err != nil {
		return err
	}

	fileName := fillString(fileData.Name(), FILE_DATA_BUFFER_SIZE)
	fileSize := fillString(strconv.FormatInt(fileData.Size(), 10), FILE_DATA_BUFFER_SIZE)
	conn.Write(fileName)
	conn.Write(fileSize)

	fmt.Println("Start sending file!")
	sendBuffer := make([]byte, MAX_BUFFER_SIZE)
	for {

		n, err := file.Read(sendBuffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("Error: ", err)
		}
		_, connErr := conn.Write(sendBuffer[:n])
		if connErr != nil {
			log.Println(connErr)
			break
		}
	}
	fmt.Println("File has been sent, closing connection!")
	return nil
}

func isFileExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Println("Error: ", err)
		return false
	}

	return true
}

func fillString(returnString string, toLength int) []byte {
	for {
		if len([]byte(returnString)) < toLength {
			returnString = ":" + returnString + ":"
			continue
		}
		break
	}
	return []byte(returnString)
}

func GenerateRandomToken(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
