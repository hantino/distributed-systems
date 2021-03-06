// Author: Professor Ivan
// Version 1.2 [added mutex to protect kvmap with concurrent clients]
// Version 1.1 [removed GoVector vector-timestamping dependency]
//
// A simple key-value store that supports three API calls over rpc:
// - get(key)
// - put(key,val)
// - testset(key,testval,newval)
//
// Usage: go run kvservicemain.go [ip:port] [key-fail-prob]
//
// - [ip:port] : the ip and TCP port on which the service will listen
//               for connections
//
// - [key-fail-prob] : probability in range [0,1] of the key becoming
//                     unavailable during one of the above operations
//                     (permanent key unavailability)
//
// TODOs:
// - [perf] Instead of having a mutex protect the entire kvmap, change
//   mutex to key granularity (one mutex per key).
// - [sim] Simulate netw. partitioning failures
// - [vtime] Ability to optionally turn on vector timestamping
// - [logging] Ability to optionally turn on logging to console
// - [sim] Ability to pass in an optional seed argument for deterministic
//   unavailability pattern

package main

import (
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

// Value in the key-val store.
type MapVal struct {
	value string // the underlying value representation
}

// Map implementing the key-value store.
var kvmap map[string]*MapVal

// Mutex for accessing kvmap from different goroutines safely.
var mapMutex *sync.Mutex

// Reserved value in the service that is used to indicate that the key
// is unavailable: used in return values to clients and internally.
const unavail string = "unavailable"

type KeyValService int

// Lookup a key, and if it's used for the first time, then initialize its value.
func lookupKey(key string) *MapVal {
	// lookup key in store
	val := kvmap[key]
	if val == nil {
		// key used for the first time: create and initialize a MapVal
		// instance to associate with a key
		val = &MapVal{
			value: "",
		}
		kvmap[key] = val
	}
	return val
}

// The probability with which a key operation triggers permanent key
// unavailability.
var failProb float64

// Check whether a key should fail with independent fail probability.
func CheckKeyFail(val *MapVal) bool {
	if val.value == unavail {
		return true
	}
	if rand.Float64() < failProb {
		val.value = unavail // permanent unavailability
		return true
	}
	return false
}

// GET
func (kvs *KeyValService) Get(args *GetArgs, reply *ValReply) error {
	// Acquire mutex for exclusive access to kvmap.
	mapMutex.Lock()
	// Defer mutex unlock to (any) function exit.
	defer mapMutex.Unlock()

	val := lookupKey(args.Key)

	if CheckKeyFail(val) {
		reply.Val = unavail
		return nil
	}

	reply.Val = val.value // execute the get
	return nil
}

// PUT
func (kvs *KeyValService) Put(args *PutArgs, reply *ValReply) error {
	// Acquire mutex for exclusive access to kvmap.
	mapMutex.Lock()
	// Defer mutex unlock to (any) function exit.
	defer mapMutex.Unlock()

	val := lookupKey(args.Key)

	if CheckKeyFail(val) {
		reply.Val = unavail
		return nil
	}

	val.value = args.Val // execute the put
	reply.Val = ""
	return nil
}

// TESTSET
func (kvs *KeyValService) TestSet(args *TestSetArgs, reply *ValReply) error {
	// Acquire mutex for exclusive access to kvmap.
	mapMutex.Lock()
	// Defer mutex unlock to (any) function exit.
	defer mapMutex.Unlock()

	val := lookupKey(args.Key)

	if CheckKeyFail(val) {
		reply.Val = unavail
		return nil
	}

	// Execute the testset.
	if val.value == args.TestVal {
		val.value = args.NewVal
	}

	reply.Val = val.value
	return nil
}

// Main server loop.
func main() {
	// Parse args.
	usage := fmt.Sprintf("Usage: %s ip:port key-fail-prob\n", os.Args[0])
	if len(os.Args) != 3 {
		fmt.Printf(usage)
		os.Exit(1)
	}

	ip_port := os.Args[1]
	arg, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if arg < 0 || arg > 1 {
		fmt.Printf(usage)
		fmt.Printf("\tkey-fail-prob arg must be in range [0,1]\n")
		os.Exit(1)
	}
	failProb = arg

	// Setup randomization.
	rand.Seed(time.Now().UnixNano())

	// Initialize the kvmap mutex.
	mapMutex = &sync.Mutex{}

	// Setup key-value store and register service.
	kvmap = make(map[string]*MapVal)
	kvservice := new(KeyValService)
	rpc.Register(kvservice)
	l, e := net.Listen("tcp", ip_port)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		conn, _ := l.Accept()
		go rpc.ServeConn(conn)
	}
}
