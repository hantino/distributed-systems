// Author: Haniel Martino

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func sendFile(fileName string, conn net.Conn) {

	fmt.Println("send to client: ", fileName)

	// file to read
	file, err := os.Open(strings.TrimSpace(fileName))
	defer file.Close()

	if err != nil {
		//conn.Write([]byte("-1"))
		log.Fatal(err)
	}

	// send file to server
	n, err := io.Copy(conn, file)
	if err != nil {
		log.Fatal(err)
	}
    // need to implement for persistent connection
    conn.Close()
	fmt.Println(n, "bytes sent")
}

func getFile(fileName string, conn net.Conn) {

    copyFileArray := strings.Split(strings.TrimSpace(fileName), ".")

    // adds copy flag for testing purposes as file is being tested in same folder
    copyFileName := copyFileArray[0] + "Copy." + copyFileArray[1]
    file, err := os.Create(strings.TrimSpace(copyFileName))

    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    conn.Write([]byte("get " + fileName))

    
    n, _ := io.Copy(file, conn)
    fmt.Println(n, "bytes received")
}

func main() {

	//get port and ip address to dial
	if len(os.Args) != 3 {
		fmt.Println("usage example: ./clientGO localhost 31415")
		return
	}

	ip := os.Args[1]
	port := os.Args[2]

	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		log.Fatal("Error trying to connect:", err)
	}

LOOP:
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter 'get <filename>' or 'send <filename>' \n\n")
	userInput, _ := reader.ReadString('\n')
	commandArray := strings.Split(userInput, " ")

	if commandArray[0] == "get" {
		getFile(commandArray[1], conn)
	} else if commandArray[0] == "send" {
		sendFile(commandArray[1], conn)
	} else {
		fmt.Println("Incorrect command")
		goto LOOP
	}
}
