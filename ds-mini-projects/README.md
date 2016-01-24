##Mini-Project Description

*This mini project was a means to give me practice in setting up a simple file transfer between a client and a server, such that the client/server makes a send or get request with the filename (i.e. full path included: /Full/Path/To/File/tooMuch.JPG) and the server responds by sending that file to the client. 
*This project is in its infancy and will later support adding files to different folders. Currently it makes a copy of the file and places it in the same folder.
*This project was tested on one host which acted as both client/server, hence retrieving and copying file to same folder. 

#### Building
`go build clientGO.go`

`go build serverGO.go`

#### Running: seperate tabs
`./clientGO serverIP serverPort`

`./serverGO`

#### Output and Usage on client
```
Please enter 'get <filename>' or 'send <filename>' 

get /Full/Path/To/File/tooMuch.JPG
4317196 bytes received
```
