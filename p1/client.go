/*
Author: Haniel Martino

Usage:
$ go run client.go [local UDP ip:port] [aserver UDP ip:port] [secret]
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

func dialServer(laddr string, raddr string) *net.UDPConn {
	serverAddr, err := net.ResolveUDPAddr("udp", raddr)
	handleError(err)

	localAddr, err := net.ResolveUDPAddr("udp", laddr)
	handleError(err)

	conn, err := net.DialUDP("udp", localAddr, serverAddr)
	handleError(err)
	return conn
}

func handleFserverConnection(conn net.Conn, fInfoMessage FortuneInfoMessage) string {
	msg := make([]byte, 1024)

	var fortuneReqMessage FortuneReqMessage
	fortuneReqMessage.FortuneNonce = fInfoMessage.FortuneNonce

	// Sending JSON encoding to aserver
	jsonReqMessage, err := json.Marshal(fortuneReqMessage)
	handleError(err)

	// Contacting fserver with nonce + secret
	_, err = conn.Write(jsonReqMessage)
	handleError(err)

	// Reading msg from aserver
	n, err := conn.Read(msg[:])
	handleError(err)

	var fortuneMessage FortuneMessage
	json.Unmarshal(msg[:n], &fortuneMessage)

	fortune := fortuneMessage.Fortune
	return fortune
}

// returns FortuneInfoMessage
func handleAserverConnection(conn net.Conn, secret int64) FortuneInfoMessage {
	msg := make([]byte, 1024)
	// Contacting aserver
	fmt.Fprintf(conn, "ping")
	n, err := conn.Read(msg[:])
	handleError(err)

	// putting json received into nonce struct
	var nonce NonceMessage
	json.Unmarshal(msg[:n], &nonce)

	var hashMessage HashMessage
	hashMessage.Hash = computeNonceSecretHash(nonce.Nonce, secret)

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
	var fInfoMessage FortuneInfoMessage
	json.Unmarshal(msg[:n], &fInfoMessage)

	return fInfoMessage
}

// Returns the MD5 hash as a hex string for the (nonce + secret) value.
func computeNonceSecretHash(nonce int64, secret int64) string {
	sum := nonce + secret
	buf := make([]byte, 512)
	n := binary.PutVarint(buf, sum)
	h := md5.New()
	h.Write(buf[:n])
	str := hex.EncodeToString(h.Sum(nil))
	return str
}

// If err is non-nil, print it out and halt.
func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

// Main
func main() {
	
	//var msg[] byte
	local := os.Args[1]
	aserver := os.Args[2]
	secretString := os.Args[3]

	// secret to 64 bit
	secret, err := strconv.ParseInt(secretString, 10, 64)
	handleError(err)

	// contact Aserver
	conn := dialServer(local, aserver)
	defer conn.Close()
	//hashMessage := handleAserverConnection(conn, secret)

	// Get FortuneInfoMessage
	fInfoMessage := handleAserverConnection(conn, secret)
	
	// connecting to FortuneServer
	conn = dialServer(local, fInfoMessage.FortuneServer)
	defer conn.Close()
	// Get Fortune
	fortune := handleFserverConnection(conn, fInfoMessage)

	// Fortune message received from fserver
	fmt.Println(fortune)
}
