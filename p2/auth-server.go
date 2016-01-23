/*
Author: Haniel Martino
Implements the solution to assignment 2 for UBC CS 416 2015 W2.
*/

package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

// Global variables
//////////////////////////////

// secret
var secret int64

// global aserver local address
var aserverUdpAddrG string

// global fserver address
var fserverG string

// global udp connect
var conndp *net.UDPConn

// client and md5 hash mapping (possibly official map)
var aserverClientMD5Map = struct {
	sync.RWMutex
	m map[string]HashMessage
}{m: make(map[string]HashMessage)}

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

func handleError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(-1)
	}
}

// Returns the MD5 hash as a hex string for the (nonce + secret) value -- a1 solution
func computeNonceSecretHash(nonce int64, secret int64) string {
	sum := nonce + secret
	buf := make([]byte, 512)
	n := binary.PutVarint(buf, sum)
	h := md5.New()
	h.Write(buf[:n])
	str := hex.EncodeToString(h.Sum(nil))
	return str
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

func sendUnexpectedHashValueError(clientAddr string) {
	var error ErrMessage
	error.Error = "unexpected hash value"

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

func sendFserverInfo(fInfoMsg FortuneInfoMessage, clientAddr string) {

	// encoding fInfoMsg to JSON to aserver
	jsonFim, err := json.Marshal(fInfoMsg)
	handleError(err)

	// client address
	clientUdpAddr, err := net.ResolveUDPAddr("udp", clientAddr)
	handleError(err)

	// sending fInfoMsg
	_, err = conndp.WriteToUDP(jsonFim, clientUdpAddr)
	handleError(err)
}

func initiateRcpConnection(clientAddr string) {
	// Connecting to fortune server
	client, err := rpc.Dial("tcp", fserverG)
	if err != nil {
		log.Fatal("Error: ", err)
		handleError(err)
	}

	var fInfoMsg FortuneInfoMessage
	err = client.Call("FortuneServerRPC.GetFortuneInfo", clientAddr, &fInfoMsg)
	if err != nil {
		log.Fatal("get fortune info error:", err)
	}
	sendFserverInfo(fInfoMsg, clientAddr)
}

func processHashMessage(clientHash HashMessage, clientAddr string) {
	aserverClientMD5Map.Lock()
	if hash, ok := aserverClientMD5Map.m[clientAddr]; ok {

		// check if hash value matches, if not, error
		if hash.Hash == clientHash.Hash {

			initiateRcpConnection(clientAddr)
			aserverClientMD5Map.Unlock()

		} else {
			sendUnexpectedHashValueError(clientAddr)
		}
	} else {
		sendUnknownAddressError(clientAddr)
	}
}

// Method for sending NonceMessage
func sendNonceMessage(err error, clientAddr string) {

	// generate seed
	rand.Seed(int64(time.Now().Nanosecond()))
	nonce63 := rand.Int()

	// convert int to int64
	var nonce64 int64
	nonce64 = int64(nonce63)

	// create a NonceMessage
	var nonce NonceMessage
	nonce.Nonce = nonce64

	// encoding NonceMessage to JSON to aserver
	jsonNonce, err := json.Marshal(nonce)
	handleError(err)

	// client address
	clientUdpAddr, err := net.ResolveUDPAddr("udp", clientAddr)
	handleError(err)

	// sending NonceMessage
	_, err = conndp.WriteToUDP(jsonNonce, clientUdpAddr)
	handleError(err)

	// Nonce and Secret
	md5Hash := computeNonceSecretHash(nonce64, secret)

	var hash HashMessage
	hash.Hash = md5Hash

	// Adding client and HashMessage value to global map
	aserverClientMD5Map.Lock()
	aserverClientMD5Map.m[clientAddr] = hash
	aserverClientMD5Map.Unlock()
}

func handleClientConnection(buf []byte, n int, clientAddr string) {
	// check message, if it's a hash, process hash, if not, send nonce
	var hash HashMessage
	err := json.Unmarshal(buf[:n], &hash)
	if err != nil {

		sendNonceMessage(err, clientAddr)

	} else {
		processHashMessage(hash, clientAddr)
	}
}

/*Usage:
go run auth-server.go [aserver UDP ip:port] [fserver RPC ip:port] [secret]
[aserver UDP ip:port] : the UDP address on which the aserver receives new client connections
[fserver RPC ip:port] : the TCP address on which the fserver listens to RPC connections from the aserver
[secret] : an int64 secret
*/

func main() {

	// Process args.
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr,
			"Usage: %s [aserver UDP ip:port] [fserver RPC ip:port] [secret]\n",
			os.Args[0])
		os.Exit(1)
	}

	msg := make([]byte, 1024)

	//fmt.Println("Setup addresses")

	// the UDP address on which the aserver receives client connections
	aserver := os.Args[1]
	aserverUdpAddrG = aserver
	aserverUdpAddr, err := net.ResolveUDPAddr("udp", aserver)
	handleError(err)

	// the TCP address on which the fserver listens to RPC connections from the aserver
	fserver := os.Args[2]
	fserverG = fserver

	// agreed upon secret
	secretArg, err := strconv.ParseInt(os.Args[3], 10, 64)
	handleError(err)

	// assign global secret variable
	secret = secretArg

	// Debug to see input from command line args
	fmt.Printf("aserver listening on %s\nSecret: %d\n", aserver, secretArg)

	conn, err := net.ListenUDP("udp", aserverUdpAddr)
	handleError(err)

	// refactor to global variable
	conndp = conn

	defer conn.Close()

	// udp client concurrency
	for {
		// fmt.Println("Listen for clients")
		n, clientAddr, err := conndp.ReadFromUDP(msg)
		handleError(err)
		go handleClientConnection(msg[:n], n, clientAddr.String())
	}
}
