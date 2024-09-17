package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
)

var (
	connMap = make(map[string]net.Conn)
	mu      sync.RWMutex
)

func main() {
	go startTcpServer()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Specify the ID to send the request to tunnel "))
	})

	http.HandleFunc("/{id}", handleRequest)
	http.HandleFunc("/{id}/*", handleRequest)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func startTcpServer() {

	listener, err := net.Listen("tcp", ":5000")

	if err != nil {
		log.Fatal("Error starting TCP server:", err)
	}
	defer listener.Close()

	fmt.Println("TCP server listening on :5000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)

	}
}

func handleConnection(conn net.Conn) {
	fmt.Println("Connection received from:", conn.RemoteAddr())

	var id string

	// keep genrating random id until we get a unique one
	for {
		id = generateRandomID(6)

		mu.RLock()
		_, ok := connMap[id]
		mu.RUnlock()

		if !ok {
			break
		}
	}

	mu.Lock()
	connMap[id] = conn
	conn.Write([]byte(id))
	mu.Unlock()

	// send the id back to the client

}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	path := r.URL.Path[len("/"+id):]

	if path == "" {
		path = "/"
	}

	conn, ok := connMap[id]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Path not found"))
		return
	}

	// check if the connection is still open
	if _, err := conn.Write([]byte("")); err != nil {
		mu.Lock()
		delete(connMap, id)
		mu.Unlock()
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Path not found"))
		return
	}

	fmt.Println("----------------------------")
	fmt.Println("Request received for ID:", id, conn.RemoteAddr())

	// remove the path by removing id from the request
	r.URL.Path = path

	if err := r.Write(conn); err != nil {
		http.Error(w, "Error writing request to backend", http.StatusInternalServerError)
		return
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)

	if err != nil {
		http.Error(w, "Error reading response from backend server", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	// add all headers from the response
	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

}

func generateRandomID(len int) string {

	ran_str := make([]byte, len)

	// Generating Random string
	for i := 0; i < len; i++ {
		ran_str[i] = byte(65 + rand.Intn(25))
	}

	// Displaying the random string
	return string(ran_str)

}
