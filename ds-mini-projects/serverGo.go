// Author: Haniel Martino

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const PORT = "31415"
const BUFFER_SIZE = 1024

// Method to handle multiple client connections
func connectionHandler(conn net.Conn) {
	buf := make([]byte, BUFFER_SIZE)
	//loop until disconntect
	for {
		n, err := conn.Read(buf)
		if err != nil {
			conn.Close()
			break
		}

		fmt.Println("Command recieved: " + string(buf[0:n]))

		commandString := strings.TrimSpace(string(buf[0:n]))
		commandArray := strings.Split(commandString, " ")

		if commandArray[0] == "get" {
			sendFile(commandArray[1], conn)
		} else {
			_, err = conn.Write([]byte("Incorrect Command"))
			if err != nil {
				conn.Close()
				break
			}
		}
	}
    
	log.Printf("Connection from %v closed.", conn.RemoteAddr())
}

// Method to handle sending files to client
func sendFile(fileName string, conn net.Conn) {

	fmt.Println("send to client:", fileName)

	// file to read
	file, err := os.Open(strings.TrimSpace(fileName))
	defer file.Close()

	if err != nil {
		//conn.Write([]byte("-1"))
		log.Fatal(err)
	}
	// send file to client
	n, err := io.Copy(conn, file)
	if err != nil {
		log.Fatal(err)
	}
    // need to implement method for persistant connection
    defer conn.Close()
	fmt.Println(n, "bytes sent")
}

func main() {

	fmt.Println("Listening for connections...")

	server, err := net.Listen("tcp", "localhost:"+PORT)
	if err != nil {
		fmt.Println("Server Error:",err)
		return
	}
    defer server.Close()
	// loop to keep listening and accepting clients
	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println("Client Error:", err)
			return
		}
		log.Printf("%v connected", conn.RemoteAddr())
        // goroutine to handle each client
		go connectionHandler(conn)
	}
}
