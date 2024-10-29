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
	Params interface{}   `json:"params"`
}

type StratumResponse struct {
	ID     int           `json:"id"`
	Result interface{}   `json:"result"`
	Error  interface{}   `json:"error"`
}

type MiningSubscribeParams struct {
	Client string `json:"client"`
}

type MiningAuthorizeParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type MiningSubscribeResult struct {
	SubscriptionID string `json:"subscription_id"`
}

type MiningAuthorizeResult struct {
	Authorized bool `json:"authorized"`
}

type StratumError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
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
		Params: MiningSubscribeParams{Client: "my-client/0.1"},
	}
	sendRequest(writer, subscribeReq)

	authorizeReq := StratumRequest{
		ID:     2,
		Method: "mining.authorize",
		Params: MiningAuthorizeParams{Username: USERNAME, Password: PASSWORD},
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

	switch response.ID {
	case 1:
		result, ok := response.Result.(map[string]interface{})
		if !ok {
			log.Printf("unexpected result type for subscription: %T", response.Result)
			return
		}
		subscribeResult := MiningSubscribeResult{
			SubscriptionID: result["subscription_id"].(string),
		}
		log.Println("Subscription:", subscribeResult)
	case 2:
		result, ok := response.Result.(map[string]interface{})
		if !ok {
			log.Printf("unexpected result type for authorization: %T", response.Result)
			return
		}
		authorizeResult := MiningAuthorizeResult{
			Authorized: result["authorized"].(bool),
		}
		log.Println("Authorization:", authorizeResult)
	default:
		log.Println("Response:", string(line))
	}
}
