package main

import (
	"context"
	"log"

	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveWs(instance *Instance, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	pool := "test"
	limit := uint16(10)

	messageChannel := make(chan interface{})
	claimChannel := make(chan []string)
	markerChannel := make(chan uint32)
	ctx, cancel := context.WithCancel(context.Background())

	session := NewSession(instance)
	session.Start(
		pool,
		messageChannel,
		claimChannel,
		markerChannel,
		ctx,
		limit,
	)

	end := func() {
		conn.Close()
		cancel()
	}

	go func() {
		defer end()
		for {
			_, data, err := conn.ReadMessage()
			var message interface{}
			if err == nil {
				message, err = UnmarshalMessage(data)
			}
			if err != nil {
				log.Printf("error: %v", err)
				break
			}
			messageChannel <- message
		}
	}()

	go func() {
		defer end()
	loop:
		for {
			var data []byte
			select {
			case ids := <-claimChannel:
				data = MarshallClaimMessage(ids)
			case poolSize := <-markerChannel:
				data = MarshallLoadMessage(poolSize)
			case <-ctx.Done():
				break loop
			}
			err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("error: %v", err)
				break loop
			}
		}
	}()
}

var addr = "127.0.0.1:3000"

func main() {
	instance := NewInstance()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(instance, w, r)
	})

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
