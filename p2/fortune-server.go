/*
Author: Haniel Martino
Implements the solution to assignment 2 for UBC CS 416 2015 W2.
*/

package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strings"
	"sync"
	"time"
)

// Global Variables

// fserverIp:Port
var fserverIpPort string
var fserverTcpG string

// fortune
var fortuneG string

// global udp connect
var conndp *net.UDPConn

// client and nonce mapping
var fserverMap = struct {
	sync.RWMutex
	m map[string]int64
}{m: make(map[string]int64)}

// Types

// An error message from the server.
type ErrMessage struct {
	Error string
}

type FortuneServerRPC struct{}

// Message with details for contacting the fortune-server.
type FortuneInfoMessage struct {
	FortuneServer string // e.g., "127.0.0.1:1234"
	FortuneNonce  int64  // e.g., 2016
}

// Message requesting a fortune from the fortune-server.
type FortuneReqMessage struct {
	FortuneNonce int64
}

// Response from the fortune-server containing the fortune.
type FortuneMessage struct {
	Fortune string
}

// Errors
/////////////////////////////

func handleError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(-1)
	}
}

func sendUnknownAddressError(clientAddr string) {
	var error ErrMessage
	error.Error = "unknown remote client address"

	// encoding ErrorMessage to JSON to aserver
	jsonError, err := json.Marshal(error)
	handleError(err)

	// client address
	clientUdpAddr, err := net.ResolveUDPAddr("udp", clientAddr)
	handleError(err)

	// sending ErrorMessage
	_, err = conndp.WriteToUDP(jsonError, clientUdpAddr)
	handleError(err)
}

func sendMalformMessageError(clientAddr string) {
	var error ErrMessage
	error.Error = "could not interpret message"

	// encoding ErrorMessage to JSON to aserver
	jsonError, err := json.Marshal(error)
	handleError(err)

	// client address
	clientUdpAddr, err := net.ResolveUDPAddr("udp", clientAddr)
	handleError(err)

	// sending ErrorMessage
	_, err = conndp.WriteToUDP(jsonError, clientUdpAddr)
	handleError(err)
}

func sendUnexpectedNonceValueError(clientAddr string) {
	var error ErrMessage
	error.Error = "incorrect fortune nonce"

	// encoding ErrorMessage to JSON to aserver
	jsonError, err := json.Marshal(error)
	handleError(err)

	// client address
	clientUdpAddr, err := net.ResolveUDPAddr("udp", clientAddr)
	handleError(err)

	// sending ErrorMessage
	_, err = conndp.WriteToUDP(jsonError, clientUdpAddr)
	handleError(err)
}

// RPC method that populates the FortuneInfoMessage with required information
func (this *FortuneServerRPC) GetFortuneInfo(clientAddr string, fInfoMsg *FortuneInfoMessage) error {
	rand.Seed(int64(time.Now().Nanosecond()))
	nonce := rand.Int()

	var nonce64 int64
	nonce64 = int64(nonce)

	//var fInfoMsgRPC FortuneInfoMessage
	fInfoMsg.FortuneServer = fserverIpPort
	fInfoMsg.FortuneNonce = nonce64

	// Adding client and nonce value to global map, whilst avoiding any race conditions
	fserverMap.Lock()
	fserverMap.m[clientAddr] = nonce64
	fserverMap.Unlock()

	return nil
}

func sendFortune(clientAddr string) {
	var fortune FortuneMessage
	fortune.Fortune = fortuneG

	// encoding ErrorMessage to JSON to aserver
	jsonFortune, err := json.Marshal(fortune)
	handleError(err)

	// client address
	clientUdpAddr, err := net.ResolveUDPAddr("udp", clientAddr)
	handleError(err)

	// sending ErrorMessage
	_, err = conndp.WriteToUDP(jsonFortune, clientUdpAddr)
	handleError(err)
}

func processReqMessage(frm FortuneReqMessage, clientAddr string) {

	nonce := frm.FortuneNonce
	fserverMap.Lock()
	if nonce64, ok := fserverMap.m[clientAddr]; ok {

		// check if hash value matches, if not, error
		if nonce == nonce64 {
			sendFortune(clientAddr)

		} else {

			sendUnexpectedNonceValueError(clientAddr)
		}

	} else {
		sendUnknownAddressError(clientAddr)
	}
	fserverMap.Unlock()

}

func handleClientConnection(buf []byte, n int, clientAddr string) {
	// check message, if it's a hash, process hash, if not, send nonce
	var fortuneReqMessage FortuneReqMessage
	err := json.Unmarshal(buf[:n], &fortuneReqMessage)
	if err != nil {

		// Sending malform message error
		sendMalformMessageError(clientAddr)

	} else {

		processReqMessage(fortuneReqMessage, clientAddr)

	}
}

// Handles connection from aserver through an rpc interface
func handleRpcConnection() {

	fortuneServerRPC := new(FortuneServerRPC)
	rpc.Register(fortuneServerRPC)

	tcpAddress, err := net.ResolveTCPAddr("tcp", fserverTcpG)
	handleError(err)

	// Listen for Tcp connections
	ln, err := net.ListenTCP("tcp", tcpAddress)
	handleError(err)

	for {

		conn, err := ln.AcceptTCP()
		handleError(err)
		go rpc.ServeConn(conn)
	}

	ln.Close()
}

/*Usage:

go run fortune-server.go [fserver RPC ip:port] [fserver UDP ip:port] [fortune-string]
[fserver RPC ip:port] : the TCP address on which the fserver listens to RPC connections from the aserver
[fserver UDP ip:port] : the UDP address on which the fserver receives client connections
[fortune-string] : a fortune string that may include spaces, but not other whitespace characters

*/

func main() {

	// Process args.

	// the TCP address on which the fserver listens to RPC connections from the aserver
	fserverTcp := os.Args[1]
	fserverTcpG = fserverTcp

	// the UDP address on which the fserver receives client connections
	fserver := os.Args[2]
	fserverUdpAddr, err := net.ResolveUDPAddr("udp", fserver)
	handleError(err)

	msg := make([]byte, 1024)

	// Global fserver ip:port info
	fserverIpPort = fserver

	// Read the rest of the args as a fortune message
	fortune := strings.Join(os.Args[3:], " ")
	fortuneG = fortune

	// Debug to see input from command line args
	fmt.Printf("fserver Listening on %s\nFortune: %s\n", fserverIpPort, fortune)

	// concurrent running of rcp connection

	conn, err := net.ListenUDP("udp", fserverUdpAddr)
	handleError(err)

	go handleRpcConnection()
	defer conn.Close()

	// refactor to global variable
	conndp = conn
	// udp client concurrency
	for {
		n, clientAddr, err := conn.ReadFromUDP(msg)
		handleError(err)
		go handleClientConnection(msg[:], n, clientAddr.String())
	}
}
