package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
)

const (
	ID_RESPONSE_LENGTH = 6
	SERVER_URL         = "http://localhost:8080"
)

func main() {

	tunneledPort := "3000"

	// connect to tcp server on port 3000
	conn, err := net.Dial("tcp", "localhost:5000")

	if err != nil {
		log.Fatal(err)
	}

	// read the tunnel ID
	id := make([]byte, ID_RESPONSE_LENGTH)
	_, err = conn.Read(id)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Tunnling traffic to %s/%s\n", SERVER_URL, id)

	destinationUrl, _ := url.Parse(fmt.Sprintf("http://localhost:%s", tunneledPort))

	reader := bufio.NewReader(conn)

	for {
		request, err := http.ReadRequest(reader)

		if err != nil {
			log.Fatal(err)
		}

		// Modify the request
		request.RequestURI = ""

		// update destination URL path
		destinationUrl.Path = request.URL.Path
		request.URL = destinationUrl

		// send this request to the server
		client := &http.Client{}
		resp, err := client.Do(request)

		if err != nil {
			fmt.Println("errrrrr", err)
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			continue
		}

		resp.Write(conn)
	}

}
