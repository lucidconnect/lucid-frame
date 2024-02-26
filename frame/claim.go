package frame

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

type MintPassResponse struct {
	Valid  bool   `json:"valid"`
	PassID string `json:"passID"`
}

func ClaimItem(itemId, claimAddress string) error {
	inverseBaseUrl := os.Getenv("INVERSE_URL")
	mintPassUrl, _ := url.Parse(fmt.Sprintf("%v/mintPass", inverseBaseUrl))
	claimUrl, _ := url.Parse(fmt.Sprintf("%v/claim", inverseBaseUrl))

	mq := mintPassUrl.Query()
	mq.Add("itemId", itemId)

	mintPassUrl.RawQuery = mq.Encode()
	httpRequest, err := http.NewRequest(http.MethodPost, mintPassUrl.String(), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	var mintPass MintPassResponse
	json.NewDecoder(resp.Body).Decode(&mintPass)

	// claim
	cq := claimUrl.Query()
	cq.Add("passId", mintPass.PassID)
	cq.Add("claimingAddress", claimAddress)

	claimUrl.RawQuery = cq.Encode()
	claimHttpRequest, _ := http.NewRequest(http.MethodPost, claimUrl.String(), nil)
	claimResp, err := http.DefaultClient.Do(claimHttpRequest)
	if err != nil {
		log.Println(err)
		return err
	}
	defer claimResp.Body.Close()

	return nil
}
