package main

import (
	"context"
	"fmt"
	"github.com/coder/websocket"
	"github.com/omarattia3143/monitor/internal/hardware"
	"log"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	subscriberMessageBuffer int
	mux                     http.ServeMux
	subscribersMutex        sync.Mutex
	subscribers             map[*Subscriber]struct{}
}

type Subscriber struct {
	msgs chan []byte
}

func NewServer() *Server {
	s := &Server{
		subscriberMessageBuffer: 10,
		subscribers:             make(map[*Subscriber]struct{}),
	}
	s.mux.Handle("/", http.FileServer(http.Dir("./htmx")))
	s.mux.HandleFunc("/ws", s.subscribeHandler)
	return s
}

func (s *Server) subscribeHandler(writer http.ResponseWriter, req *http.Request) {
	err := s.subscribe(req.Context(), writer, req)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (s *Server) addSubscriber(subscriber *Subscriber) {
	s.subscribersMutex.Lock()
	defer s.subscribersMutex.Unlock()
	s.subscribers[subscriber] = struct{}{}
	fmt.Println("Added subscriber", subscriber)
}

func (s *Server) subscribe(ctx context.Context, writer http.ResponseWriter, req *http.Request) error {
	var c *websocket.Conn
	subscriber := &Subscriber{msgs: make(chan []byte, s.subscriberMessageBuffer)}
	s.addSubscriber(subscriber)
	c, err := websocket.Accept(writer, req, nil)
	if err != nil {
		return err
	}
	defer func(c *websocket.Conn) {
		err := c.CloseNow()
		if err != nil {
		}
	}(c)
	ctx = c.CloseRead(ctx)
	for {
		select {
		case msg := <-subscriber.msgs:
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			err := c.Write(ctx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Server) broadcast(msg []byte) {
	s.subscribersMutex.Lock()
	defer s.subscribersMutex.Unlock()
	for subscriber := range s.subscribers {
		subscriber.msgs <- msg
	}
}

func main() {
	fmt.Println("Starting system monitor...")
	srv := NewServer()

	go func(s *Server) {
		for {
			systemSection, err := hardware.GetSystemSection()
			if err != nil {
				fmt.Println(err)
			}

			cpuSection, err := hardware.GetCpuSection()
			if err != nil {
				fmt.Println(err)
			}

			diskSection, err := hardware.GetDiskSection()
			if err != nil {
				fmt.Println(err)
			}

			timeStamp := time.Now().Format("2006-01-02 15:04:05")
			htmx := fmt.Sprintf(`
    <div hx-swap-oob="innerHTML:#update-timestamp">%s</div>
    <div hx-swap-oob="innerHTML:#system-data">%s</div>
    <div hx-swap-oob="innerHTML:#cpu-data">%s</div>
    <div hx-swap-oob="innerHTML:#disk-data">%s</div>
`, timeStamp, systemSection, cpuSection, diskSection)

			srv.broadcast([]byte(htmx))

			time.Sleep(time.Second)
		}
	}(srv)

	err := http.ListenAndServe("localhost:8080", &srv.mux)
	if err != nil {
		log.Fatal(err)
	}
}
