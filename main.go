package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
)

type StratumRequest struct {
	ID     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type StratumResponse struct {
	ID     int           `json:"id"`
	Result interface{}   `json:"result"`
	Error  []interface{} `json:"error"`
}

// Use this list https://minerstat.com/pool-status-checker
const (
	SERVER   string = "etc.2miners.com:1010"
	USERNAME string = "0x63a14c53f676f34847b5e6179c4f5f5a07f0b1ed"
	PASSWORD string = "x"
)

func main() {

	conn, err := net.Dial("tcp", SERVER)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("error closing connection: %v", err)
		}
	}(conn)
	log.Println("Connected to Stratum-server")

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	subscribeReq := StratumRequest{
		ID:     1,
		Method: "mining.subscribe",
		Params: []interface{}{"my-client/0.1"},
	}
	sendRequest(writer, subscribeReq)

	authorizeReq := StratumRequest{
		ID:     2,
		Method: "mining.authorize",
		Params: []interface{}{USERNAME, PASSWORD},
	}
	sendRequest(writer, authorizeReq)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			log.Fatalf("error reading: %v", err)
		}
		handleResponse(line)
	}

}

func sendRequest(writer *bufio.Writer, request StratumRequest) {
	reqBytes, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("error marshaling JSON: %v", err)
	}
	_, err = writer.Write(reqBytes)
	if err != nil {
		log.Fatalf("error writing bytes: %v", err)
	}
	_, err = writer.Write([]byte("\n"))
	if err != nil {
		log.Fatalf("error writing bytes: %v", err)
	}
	err = writer.Flush()
	if err != nil {
		log.Fatalf("error writing bytes: %v", err)
	}
}

func handleResponse(line []byte) {
	var response StratumResponse
	err := json.Unmarshal(line, &response)
	if err != nil {
		log.Printf("error unmarshaling JSON: %v", err)
		return
	}

	if response.ID == 1 {
		log.Println("Subscription:", response.Result)
	} else if response.ID == 2 {
		log.Println("Authorization:", response.Result)
	} else {
		log.Println("Response:", string(line))
	}
}
