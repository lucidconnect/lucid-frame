package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"lucid.frame/protobufs"
)

func main() {

}

func frameHandler() http.HandlerFunc {
	type reqBody struct {
		UntrustedData map[string]any    `json:"untrustedData`
		TrustedData   map[string]string `json:"trustedData`
	}
	return func(w http.ResponseWriter, r *http.Request) {

		// print hello world
		switch r.Method {
		case http.MethodPost:
			frameReqBody := reqBody{}
			err := json.NewDecoder(r.Body).Decode(&frameReqBody)
			if err != nil {
				log.Printf("error decoding request body %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("an unexpected error occured"))
			}

			msgData := frameReqBody.TrustedData["messageBytes"]
			fmt.Println("messageBytes: ", msgData)
			message, err := validateFrameData(msgData)
			if err != nil {
				log.Printf("error validating frame %v", err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("invalid frame message"))
			}

			fmt.Println(message)
		}
	}
}

func validateFrameData(data string) (*protobufs.Message, error) {
	responseBody := protobufs.ValidationResponse{}
	msgDataBytes := []byte(data)

	url := "https://nemes.farcaster.xyz:2281/v1/validateMessage"

	resp, err := http.Post(url, "application/octet-stream", bytes.NewBuffer(msgDataBytes))
	if err != nil {
		log.Printf("failed to send POST request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		fmt.Println("successfully sent message")
	} else {
		fmt.Printf("Failed to send the message. HTTP status: %d\n", resp.StatusCode)
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return nil, err
	}

	if !responseBody.Valid {
		fmt.Println("message is invalid")
		return nil, err
	}

	message := responseBody.Message
	// validate url
	// urlBuffer := message.Data.GetFrameActionBody().Url

	// urlString := hexutil.Encode(urlBuffer)
	return message, nil
}
