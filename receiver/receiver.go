package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const FILE_DATA_BUFFER_SIZE = 1024
const MAX_BUFFER_SIZE = 4 * 1024

var fileSize int64

func main() {
	serverURL, token, path := getUserInput()
	conn, err := net.Dial("tcp", serverURL)
	if err != nil {
		log.Println("Error: ", err)
	}
	defer conn.Close()

	conn.Write([]byte(token))

	newFile, err := createNewFile(conn, path)
	if err != nil {
		log.Println("Error: ", err)
	}
	defer newFile.Close()
	var receivedBytes int64

	for {
		if (fileSize - receivedBytes) < MAX_BUFFER_SIZE {
			io.CopyN(newFile, conn, (fileSize - receivedBytes))
			conn.Read(make([]byte, (receivedBytes+MAX_BUFFER_SIZE)-fileSize))
			break
		}
		io.CopyN(newFile, conn, MAX_BUFFER_SIZE)
		receivedBytes += MAX_BUFFER_SIZE
	}

	fmt.Println("File has received")
}

func getUserInput() (string, string, string) {
	serverURL := ""
	token := ""
	path := ""
	switch int(len(os.Args)) {
	case 4:
		serverURL = os.Args[1]
		token = os.Args[2]
		path = os.Args[3]
	case 3:
		serverURL = os.Args[1]
		token = os.Args[2]
	default:
		log.Println("Please provide a valid serverURL, a token, and a path for storting your file")
		os.Exit(1)
	}

	return serverURL, token, path
}

func createNewFile(conn net.Conn, path string) (*os.File, error) {
	fileName, err := getFileName(conn)
	if err != nil {
		return nil, err
	}

	createDirectory(path)

	filePath := createValidFilePath(fileName, path)
	return os.Create(filePath)
}

func getFileName(conn net.Conn) (string, error) {
	bufferFileName := make([]byte, FILE_DATA_BUFFER_SIZE)
	bufferFileSize := make([]byte, FILE_DATA_BUFFER_SIZE)

	conn.Read(bufferFileName)
	conn.Read(bufferFileSize)

	fileName := strings.Trim(string(bufferFileName), ":")
	fileSize, _ = strconv.ParseInt((strings.Trim(string(bufferFileSize), ":")), 10, 64)

	if !strings.Contains(fileName, ".") {
		return "", errors.New("No file has been sent from the sender")
	}
	return fileName, nil
}

func createValidFilePath(fileName string, path string) string {
	filePath, _ := filepath.Abs(path)
	filePath = filePath + "/" + fileName
	if isFileExists(filePath) {
		filePath = createUniqueFileName(filePath)
	}
	return filePath
}

func isFileExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Println("Error: ", err)
		return false
	}

	return true
}

func createUniqueFileName(fileName string) string {
	currentTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	fileNameAndExtention := strings.Split(fileName, ".")
	return fileNameAndExtention[0] + currentTimestamp + "." + fileNameAndExtention[1]
}

func createDirectory(directoryPath string) {
	if len(directoryPath) < 1 {
		return
	}
	pathErr := os.MkdirAll(directoryPath, 0777)

	if pathErr != nil {
		log.Println("Error: ", pathErr)
	}
}
