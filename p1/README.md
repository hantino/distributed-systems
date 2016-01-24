`go run client.go [local UDP ip:port]  [aserver ip:port]  [secret]`

**Problem 1 Description**
Each client implements a sequential control flow, interacting with aserver first, and later with the fserver. The client communicates with both servers over UDP, using binary-encoded JSON messages.

The client is run knowing the UDP IP:port of the aserver and an int64 secret value. It follows the following steps:

1. the client sends a UDP message with arbitrary payload to the aserver
2. the client receives a NonceMessage reply containing an int64 nonce from the aserver
3. the client computes an MD5 hash of the (nonce + secret) value and sends this value as a hex string to the aserver as part of a HashMessage
4. the aserver verifies the received hash and replies with a FortuneInfoMessage that contains information for contacting fserver (its UDP IP:port and an int64 fortune nonce to use when connecting to it)
5. the client sends a FortuneReqMessage to fserver
6. the client receives a FortuneMessage from the fserver
7. the client prints out the received fortune string as the last thing before exiting on a new newline-terminated line

*The communication steps in this protocol are illustrated in the following space-time diagram:*

![*The communication steps in this protocol are illustrated in the following space-time diagram:*](http://www.cs.ubc.ca/~bestchai/teaching/cs416_2015w2/assign1/assign1-proto.jpg)
