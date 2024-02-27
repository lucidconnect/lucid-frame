package neynar

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type NeynarClient struct {
	client       *http.Client
	apiKey       string
	neynarUrl    string
	farcasterHub string
}

type Option func(*NeynarClient)

func NewNeynarClient(options ...Option) (*NeynarClient, error) {
	neynarClient := &NeynarClient{
		client:       &http.Client{},
		apiKey:       "",
		neynarUrl:    "https://api.neynar.com",
		farcasterHub: "https://hub-api.neynar.com",
	}

	for _, opt := range options {
		opt(neynarClient)
	}

	if neynarClient.apiKey == "" {
		return nil, errors.New("trying to initialise neynar without an api key")
	}

	return neynarClient, nil
}

// WithTimeout is a functional option to set the HTTP client timeout.
func WithApiKey(key string) Option {
	return func(c *NeynarClient) {
		c.apiKey = key
	}
}

func (nc *NeynarClient) ValidateFrameMessage(frameMessage []byte) (NeynarFrameValidation, error) {
	buttonAction := NeynarFrameValidation{}
	url := fmt.Sprintf("%v/v2/farcaster/frame/validate", nc.neynarUrl)

	resp, err := nc.makeRequest(http.MethodPost, url, "application/json", bytes.NewBuffer(frameMessage))
	if err != nil {
		log.Printf("failed to send POST request: %v", err)
		return NeynarFrameValidation{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		response, err := decodeFrameValidationResponse(resp.Body)
		if err != nil {
			log.Println("decodeing frame validation response failed", err)
			return NeynarFrameValidation{}, err
		}
		buttonAction = response.Action
	default:
		return NeynarFrameValidation{}, fmt.Errorf("call to validate frame message failed with HTTP status: %d", resp.StatusCode)
	}

	return buttonAction, nil
}

// func (nc *NeynarClient) ResolveVerifiedEthereumAddress(fid uint64) (string, error) {
// 	var ethereumAddress string
// 	url := fmt.Sprintf("%v/v2/farcaster/user/bulk?fids=%v", nc.neynarUrl, fid)

// 	resp, err := nc.makeRequest(http.MethodGet, url, "application/json", nil)
// 	if err != nil {
// 		log.Printf("executing http request failed: %v", err)
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	switch resp.StatusCode {
// 	case http.StatusOK:
// 		response, err := decodeVerifiedEthereumAddress(resp.Body)
// 		if err != nil {
// 			log.Println("failed to decode verified ethereum address", err)
// 			return "", err
// 		}
// 		ethereumAddress = response
// 	default:
// 		return "", fmt.Errorf("call to resolve verified ethereum address failed with HTTP status: %d", resp.StatusCode)
// 	}

// 	return ethereumAddress, nil
// }

func (nc *NeynarClient) makeRequest(method, url, contentType string, body io.Reader) (*http.Response, error) {
	httpRequest, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	httpRequest.Header.Set("api_key", nc.apiKey)
	httpRequest.Header.Set("Content-Type", contentType)
	httpRequest.Header.Set("accept", "appication/json")

	resp, err := nc.client.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func decodeFrameValidationResponse(response io.ReadCloser) (FrameValidationResponse, error) {
	var err error
	validationResponse := FrameValidationResponse{}

	if err = json.NewDecoder(response).Decode(&validationResponse); err != nil {
		err = fmt.Errorf("failed to decode response body: %v", err)
		return FrameValidationResponse{}, err
	}
	if !validationResponse.Valid {
		fmt.Println("message is invalid")
		return FrameValidationResponse{}, err
	}

	fmt.Println(validationResponse)
	return validationResponse, nil
}

// func decodeVerifiedEthereumAddress(response io.ReadCloser) (string, error) {
// 	var err error
// 	farcasterUserResponse := FarcasterUserResponse{}

// 	if err = json.NewDecoder(response).Decode(&farcasterUserResponse); err != nil {
// 		err = fmt.Errorf("failed to decode response body: %v", err)
// 		return "", err
// 	}

// 	if len(farcasterUserResponse.Users) == 0 {
// 		err = fmt.Errorf("length of farcasterUserResponse.Users = 0")
// 		return "", err
// 	}

// 	if len(farcasterUserResponse.Users[0].Verifications) == 0 {
// 		return "", ErrNoVerifications
// 	}
// 	return farcasterUserResponse.Users[0].Verifications[0], nil
// }
