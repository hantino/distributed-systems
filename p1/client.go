// Author: Haniel Martino 

/*

Usage:
$ go run client.go [local UDP ip:port] [aserver UDP ip:port] [secret]

Example:
$ go run client.go 127.0.0.1:2020 127.0.0.1:7070 1984

*/

package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

/////////// Msgs used by both auth and fortune servers:

// An error message from the server.
type ErrMessage struct {
	Error string
}

/////////// Auth server msgs:

// Message containing a nonce from auth-server.
type NonceMessage struct {
	Nonce int64
}

// Message containing an MD5 hash from client to auth-server.
type HashMessage struct {
	Hash string
}

// Message with details for contacting the fortune-server.
type FortuneInfoMessage struct {
	FortuneServer string
	FortuneNonce  int64
}

/////////// Fortune server msgs:

// Message requesting a fortune from the fortune-server.
type FortuneReqMessage struct {
	FortuneNonce int64
}

// Response from the fortune-server containing the fortune.
type FortuneMessage struct {
	Fortune string
}

/*Usage:
$ go run client.go [local UDP ip:port] [aserver UDP ip:port] [secret]

Example:
$ go run client.go 127.0.0.1:2020 198.162.52.206:1999 1984

*/
func handleError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(-1)
	}
}

// Main workhorse method.
func main() {

	msg := make([]byte, 1024)
	local := os.Args[1]
	aserver := os.Args[2]
	secretString := os.Args[3]

	// secret to 64 bit
	secret, err := strconv.ParseInt(secretString, 10, 64)
	handleError(err)

	serverAddr, err := net.ResolveUDPAddr("udp", aserver)
	handleError(err)

	localAddr, err := net.ResolveUDPAddr("udp", local)
	handleError(err)

	conn, err := net.DialUDP("udp", localAddr, serverAddr)
	handleError(err)

	// Contacting aserver
	fmt.Fprintf(conn, "Hello aserver")
	n, err := conn.Read(msg[:])
	handleError(err)

	// putting json received into nonce struct
	var nonce NonceMessage
	json.Unmarshal(msg[:n], &nonce)

	// calculating nonce + secret for MD5
	var nonceAndSecret int64
	nonceAndSecret = nonce.Nonce + secret
	// fmt.Println(nonceAndSecret)
	buf := make([]byte, 16)
	nb := binary.PutVarint(buf, nonceAndSecret)
	slicedBuf := buf[:nb]

	// Creating MD5 hash to aserver
	hash := md5.New()
	hash.Write(slicedBuf)

	var hashMessage HashMessage
	hashMessage.Hash = hex.EncodeToString(hash.Sum(nil))

	// Sending JSON encoding to aserver
	jsonMD5, err := json.Marshal(hashMessage)
	handleError(err)

	// Contacting aserver with nonce + secret
	_, err = conn.Write(jsonMD5)
	handleError(err)

	// Reading msg from aserver
	n, err = conn.Read(msg[:])
	handleError(err)
	conn.Close()

	// putting json received into Fortune struct
	var fortuneInfoMessage FortuneInfoMessage
	json.Unmarshal(msg[:n], &fortuneInfoMessage)

	// connecting to FortuneServer
	fserverAddr, err := net.ResolveUDPAddr("udp", fortuneInfoMessage.FortuneServer)
	handleError(err)
	conn, err = net.DialUDP("udp", localAddr, fserverAddr)
	handleError(err)

	var fortuneReqMessage FortuneReqMessage
	fortuneReqMessage.FortuneNonce = fortuneInfoMessage.FortuneNonce

	// Sending JSON encoding to aserver
	jsonReqMessage, err := json.Marshal(fortuneReqMessage)
	handleError(err)

	// Contacting fserver with nonce + secret
	_, err = conn.Write(jsonReqMessage)
	handleError(err)

	// Reading msg from aserver
	n, err = conn.Read(msg[:])
	handleError(err)

	var fortuneMessage FortuneMessage
	json.Unmarshal(msg[:n], &fortuneMessage)

	// Fortune message received from fserver
	fmt.Println(fortuneMessage.Fortune)

	conn.Close()
}
