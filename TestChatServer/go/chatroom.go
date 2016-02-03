package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

// Msg stores both the message and the connection
type Msg struct {
	sender *websocket.Conn
	msg    string
}

var upgrader = websocket.Upgrader{} // use default options
var reg = make(chan *websocket.Conn)
var unreg = make(chan *websocket.Conn)
var msg = make(chan Msg)

func run(reg chan *websocket.Conn, unreg chan *websocket.Conn, msg chan Msg) {
	conns := make(map[*websocket.Conn]int)
	for {
		select {
		case c := <-reg:
			conns[c] = 1
			fmt.Printf("%p", c)
		case c := <-unreg:
			delete(conns, c)
		case msg := <-msg:
			for c := range conns {
				if c != msg.sender {
					c.WriteMessage(websocket.TextMessage, []byte(msg.msg))
				}
			}
		}
	}
}

func ChatServer(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("upgrade:", err)
		return
	}
	defer c.Close()
	reg <- c
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			unreg <- c
			break
		}
		msg <- Msg{c, string(message)}

	}
}

func main() {

	http.HandleFunc("/chat", ChatServer)
	http.Handle("/", http.FileServer(http.Dir("../static")))

	go run(reg, unreg, msg)

	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
