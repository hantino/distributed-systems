/* Author: Haniel Martino
 */
package main

import (
	"fmt"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

// args in get(args)
type GetArgs struct {
	Key string // key to look up
}

// args in put(args)
type PutArgs struct {
	Key string // key to associate value with
	Val string // value
}

// args in testset(args)
type TestSetArgs struct {
	Key     string // key to test
	TestVal string // value to test against actual value
	NewVal  string // value to use if testval equals to actual value
}

// Reply from service for all three API calls above.
type ValReply struct {
	Val string // value; depends on the call
}

type KeyValService int

type PingBit struct {
	value string
}

var idsPing map[string]*PingBit
var node_keys []string
var myKey string
var myID string
var pingBit int
var client *rpc.Client
var myIdPingGlobal string
var leader bool

// function to constantly ping server, flipping between 1 and 0
func pingServer() {
	var kvVal ValReply

	// flip between 1 and 0, inactive node if same bit comes twice
	pingBit = 1 - pingBit

	pingStr := strconv.Itoa(pingBit)

	// Put("myKey", myID-ping-bit)
	myIdPing := myID + pingStr
	myIdPingGlobal = myIdPing

	putArgs := PutArgs{
		Key: myKey,
		Val: myIdPing}
	err := client.Call("KeyValService.Put", putArgs, &kvVal)
	checkError(err)
}

// function to assign key to node
func assignKey() {
	key := 0

	pingStr := strconv.Itoa(pingBit)
	myIdPing := myID + pingStr
	myIdPingGlobal = myIdPing

	for {
		var kvVal ValReply
		myKey := strconv.Itoa(key)
		tsArgs := TestSetArgs{
			Key:     myKey,
			TestVal: "",
			NewVal:  myIdPingGlobal,
		}
		err := client.Call("KeyValService.TestSet", tsArgs, &kvVal)
		checkError(err)
		if kvVal.Val == myIdPing {
			break
		}
		key++
	}
	myKey = strconv.Itoa(key)
}

// keep track of the living nodes in kvService
func isAlive(id, idBit string) bool {
	val := idsPing[id]

	// assigning new node id to bit
	if val == nil || val.value != idBit {
		val = &PingBit{
			value: idBit,
		}
		idsPing[id] = val
		return true
	}

	return false
}

// flag to indicate dead key in kvService
func setKeyToDeadNode(key string) {
	var kvVal ValReply

	putArgs := PutArgs{
		Key: key,
		Val: "dead"}
	err := client.Call("KeyValService.Put", putArgs, &kvVal)
	checkError(err)
}

// simple algorithm to check if node is at the head of list
func leaderAlgorithm(ids []string) {
	if ids[0] == myID {
		leader = true
	} else {
		leader = false
	}
}

// print ids where leader is the first in list
func printIDs(ids []string) {
	leaderAlgorithm(ids)

	for _, id := range ids {
		_id := id[:len(id)-1]
		fmt.Print(_id)
		fmt.Print(" ")
	}
	fmt.Println()
}

// get the IDs of available nodes
func getIDs() {
	key := 0
	ids := []string{}

	// loop through keys 0 - N
	for {
		var kvVal ValReply
		keyS := strconv.Itoa(key)
		getArgs := GetArgs{keyS}
		err := client.Call("KeyValService.Get", getArgs, &kvVal)
		checkError(err)

		if kvVal.Val == "" {
			break
		}

		n := len(kvVal.Val)

		// only adding keys that are available
		if kvVal.Val != "unavailable" && kvVal.Val != "dead" {
			id := kvVal.Val[:n-1]
			//fmt.Println(id)
			idBit := kvVal.Val[n-1:]
			// checking for node's heartbeat (1/1 or 0/0 = dead, 0/1 = alive)
			if isAlive(id, idBit) {
				ids = append(ids, kvVal.Val)
			} else {
				// set node key to dead when node fails, only leader does this
				if leader {
					setKeyToDeadNode(keyS)
				}
			}
		}
		key++
	}
	// Print keys from kvService, but first check if my key became unavailable
	if reAssignKeyCheck(ids) {
		printIDs(ids)
	} else {
		// if unvailable, don't print, but assign new key to self
		assignKey()
	}
}

// Check if myid is part of the available list
func reAssignKeyCheck(ids []string) bool {
	for _, id := range ids {
		if id == myIdPingGlobal {
			return true
		}
	}
	return false
}

/*go run node.go [ip:port] [id]
[ip:port] : address of the key-value service
[id] : a unique string identifier for the node (no spaces)*/

// Main server loop.
func main() {
	// parse args
	usage := fmt.Sprintf("Usage: %s ip:port id\n", os.Args[0])
	if len(os.Args) != 3 {
		fmt.Printf(usage)
		os.Exit(1)
	}

	kvAddr := os.Args[1]

	// Set as current value to associate with keys or nodes
	id := os.Args[2]
	myID = id

	// Connect to the KV-service via RPC.
	kvService, err := rpc.Dial("tcp", kvAddr)
	client = kvService
	checkError(err)

	idsPing = make(map[string]*PingBit)
	assignKey()

	// timer for constant tick
	t := time.NewTicker(5000 * time.Millisecond)
	for {
		// Get keys from kvService
		getIDs()
		pingServer()
		<-t.C
	}
}

// If error is non-nil, print it out and halt.
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		os.Exit(1)
	}
}
