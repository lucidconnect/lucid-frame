package neynar

import (
	"errors"
	"time"
)

var (
	ErrNoVerifications = errors.New("no verifications found")
)

type FrameValidationResponse struct {
	Valid  bool                  `json:"valid"`
	Action NeynarFrameValidation `json:"action"`
}

// type Message struct {
// 	Data            MessageData `json:"data"`
// 	Hash            string      `json:"hash"`
// 	HashScheme      string      `json:"hashScheme"`
// 	Signature       string      `json:"signature"`
// 	SignatureScheme string      `json:"signatureScheme"`
// 	Signer          string      `json:"signer"`
// }

// type MessageData struct {
// 	Type            string      `json:"type"`
// 	Fid             uint64      `json:"fid"`
// 	Timestamp       uint32      `json:"timestamp"`
// 	Network         string      `json:"network"`
// 	FrameActionBody FrameAction `json:"frameActionBody"`
// }

// type FrameAction struct {
// 	Url         string `json:"url"`
// 	ButtonIndex uint32 `json:"buttonIndex"`
// 	CastId      struct {
// 		Fid  uint64 `json:"fid"`
// 		Hash string `json:"hash"`
// 	} `json:"castId"`
// 	InputText string `json:"inputText"`
// }

// type FarcasterUserResponse struct {
// 	Users []User
// }

// // User represents a Farcaster user as returned by Neynar's api
// type User struct {
// 	Object         string   `json:"object"`
// 	Fid            uint64   `json:"fid"`
// 	CustodyAddress string   `json:"custody_address"`
// 	Verifications  []string `json:"verifications"`
// }

/*
  - "verified_addresses": {
    "eth_addresses": [],
    "sol_addresses": []
    },
*/

type VerifiedAddresses struct {
	EthAddresses []string `json:"eth_addresses"`
	SolAddresses []string `json:"sol_addresses"`
}

type Interactor struct {
	Fid              int               `json:"fid"`
	VerifiedAdresses VerifiedAddresses `json:"verified_addresses"`
}
type Button struct {
	Title      string `json:"title"`
	Index      int    `json:"index"`
	ActionType string `json:"action_type"`
}
type NeynarFrameValidation struct {
	Object       string     `json:"object"`
	Interactor   Interactor `json:"interactor"`
	TappedButton Button     `json:"tapped_button"`
	Input        struct {
		Text string `json:"text"`
	} `json:"input"`

	Cast      Cast      `json:"cast"`
	Timestamp time.Time `json:"timestamp"`
}

type Cast struct {
	Frames []Frame `json:"frames"`
}

type Frame struct {
	Version string        `json:"version"`
	Image   string        `json:"image"`
	Buttons []FrameButton `json:"buttons"`
}

type FrameButton struct {
	Index      int    `json:"index"`
	Title      string `json:"title"`
	ActionType string `json:"action_type"`
}
