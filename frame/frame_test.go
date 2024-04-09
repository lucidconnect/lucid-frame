package frame

import (
	"fmt"
	"net/url"
	"testing"
	"time"
)

// "image/jpeg"

// func Test_GenerateCounterFactualAddress(t *testing.T) {
// 	os.Setenv("KERNEL_IMPLEMENTATION_ADDRESS", "0xf048AD83CB2dfd6037A43902a2A5Be04e53cd2Eb")
// 	os.Setenv("DEFAULT_VALIDATOR", "0x20d0f856eF56F5A9fA5AE27Fcf698388de85e4c1")
// 	kernelFactory := "0x5de4839a76cf55d0c90e2061ef4386d962E15ae3"
// 	rpc := "https://base-mainnet.g.alchemy.com/v2/vzwdbWEEKGUWa9lHQJ4RnKaBGNcvniN9"
// 	ed25519Signer := "0x6349813e5362a6d500b9a41bd066e92ccfe492d5a27892e028f7adcdbf920116"

// 	_, err := frame.CalculateCounterFactualAddress(ed25519Signer, kernelFactory, rpc)
// 	assert.NoError(t, err, "error calculating counterfactual address")
// }

// func Test_GenerateImage(t *testing.T) {
// 	b, err := generateImage("test image", "#000000", "#FFFFFF", 32)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// image.
// 	out, err := os.Create("./sample.jpg")
// 	assert.NoError(t, err)
// 	// jpeg.Encode(out)
// }

func Test_ImageUrls(t *testing.T) {
	parsedURL, _ := url.Parse("https://jade-historic-pony-314.mypinata.cloud/ipfs/QmZnC193fR8C1qZyLf8XgTn7UzTTYNnyva81sJUeeMWd74")
	timestamp := time.Now().Unix() // Get current Unix timestamp
	query := parsedURL.Query()
	query.Set("_", fmt.Sprintf("%d", timestamp)) // Append timestamp as query parameter
	parsedURL.RawQuery = query.Encode()

	newImageUrl := parsedURL.String()

	fmt.Println(newImageUrl)
}
