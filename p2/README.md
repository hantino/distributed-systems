**Problem 2 Description**

The aserver and fserver communicate via RPC over TCP. The fserver is the server in this RPC interaction and exports a single method to the aserver, GetFortuneInfo, that takes the address of the client and a pointer to FortuneInfoMessage for the result. The fserver computes a new nonce and returns the filled-in FortuneInfoMessage. 

***The exact declaration of GetFortuneInfo and the input/output types is:***

```
type FortuneServerRPC struct{}

// Message with details for contacting the fortune-server.
type FortuneInfoMessage struct {
	FortuneServer string // e.g., "127.0.0.1:1234"
	FortuneNonce  int64  // e.g., 2016
}

func (this *FortuneServerRPC) GetFortuneInfo(clientAddr string,	fInfoMsg *FortuneInfoMessage) error { ... } 
```
*The communication steps in this protocol are illustrated in the following space-time diagram:*

![](http://www.cs.ubc.ca/~bestchai/teaching/cs416_2015w2/assign2/assign2-servers-proto.jpg)
