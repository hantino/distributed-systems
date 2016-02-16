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

// Method to handle receving files
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
	fmt.Print("Please enter 'get <filename>'\n\n")
	userInput, _ := reader.ReadString('\n')
	commandArray := strings.Split(userInput, " ")

	if commandArray[0] == "get" {
		getFile(commandArray[1], conn)
	} else {
		fmt.Println("Incorrect command")
		goto LOOP
	}
}
