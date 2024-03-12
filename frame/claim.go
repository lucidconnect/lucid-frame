package frame

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"lucid.frame/lucidNft"
)

type MintPassResponse struct {
	Valid  bool   `json:"valid"`
	PassID string `json:"passID"`
}

type MintAuthorizationResponse struct {
	Amount               string `json:"amount"`
	TokenID              string `json:"tokenId"`
	Nonce                string `json:"nonce"`
	Chain                int    `json:"chain"`
	PackedData           string `json:"packedData"`
	MintingAbi           string `json:"mintingABI"`
	MintingSignature     string `json:"mintingSignature"`
	SmartContractAddress string `json:"smartContractAddress"`
}

func ClaimItem(itemId, claimAddress string) (string, error) {
	// blockExplorer := os.Getenv("BLOCK_EXPLORER")
	inverseBaseUrl := os.Getenv("INVERSE_URL")
	mintPassUrl, _ := url.Parse(fmt.Sprintf("%v/mintPass", inverseBaseUrl))
	claimUrl, _ := url.Parse(fmt.Sprintf("%v/claim", inverseBaseUrl))

	mq := mintPassUrl.Query()
	mq.Add("itemId", itemId)
	mq.Add("wallet", claimAddress)

	mintPassUrl.RawQuery = mq.Encode()
	httpRequest, err := http.NewRequest(http.MethodPost, mintPassUrl.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		log.Println("http request error ", err)
		return "", err
	}
	defer resp.Body.Close()

	var mintPass MintPassResponse
	if err = json.NewDecoder(resp.Body).Decode(&mintPass); err != nil {
		log.Println("decoding mint pass failed ", err)
		return "", err
	}

	if !mintPass.Valid {
		return "mint limit reached", nil
	}
	// claim
	cq := claimUrl.Query()
	cq.Add("passId", mintPass.PassID)
	cq.Add("claimingAddress", claimAddress)

	claimUrl.RawQuery = cq.Encode()
	claimHttpRequest, _ := http.NewRequest(http.MethodPost, claimUrl.String(), nil)
	claimResp, err := http.DefaultClient.Do(claimHttpRequest)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer claimResp.Body.Close()

	var signatureResponse MintAuthorizationResponse
	if err = json.NewDecoder(claimResp.Body).Decode(&signatureResponse); err != nil {
		log.Println("getting mint authorization error", err)
		fmt.Println(claimResp.Body)
		return "", err
	}

	claimer := common.HexToAddress(claimAddress)
	contract := common.HexToAddress(signatureResponse.SmartContractAddress)

	amount, _ := new(big.Int).SetString(signatureResponse.Amount, 10)
	tokenId, _ := new(big.Int).SetString(signatureResponse.TokenID, 10)
	nonce, _ := new(big.Int).SetString(signatureResponse.Nonce, 10)
	// amount.SetString(signatureResponse.Amount, 10)
	// tokenId.SetString(signatureResponse.TokenID, 10)
	// nonce.SetString(signatureResponse.Nonce, 10)
	signature, err := hex.DecodeString(signatureResponse.MintingSignature)
	if err != nil {
		err = fmt.Errorf("decoding signature failed with error %v", err)
		log.Println("signature ", signatureResponse.MintingSignature)
		return "", err
	}

	txHash, err := MintNft(contract, claimer, amount, tokenId, nonce, signature, int64(signatureResponse.Chain))
	if err != nil {
		log.Println("minting nft failed", err)
		return "", err
	}

	// tx := fmt.Sprintf("%v/tx/%v", blockExplorer, txHash)
	return txHash, nil
}

func MintNft(contractAddress, claimAddress common.Address, amount, tokenId, nonce *big.Int, signature []byte, chain int64) (string, error) {
	rpc := os.Getenv("RPC")
	privateKey := os.Getenv("PRIVATE_KEY")
	backend := getEthBackend(rpc)
	lucidNftCaller, err := lucidNft.NewLucidNft(contractAddress, backend)
	if err != nil {
		log.Println(err)
		return "", err
	}

	transactOpt, err := getTransactor(privateKey, chain)
	if err != nil {
		log.Println(err)
		return "", err
	}
	tx, err := lucidNftCaller.Mint(transactOpt, claimAddress, amount, tokenId, nonce, signature)
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

func getTransactor(privateKey string, chain int64) (*bind.TransactOpts, error) {
	key, err := crypto.HexToECDSA(privateKey[2:])
	if err != nil {
		return nil, err
	}

	chainId := big.NewInt(chain)

	transactionOpts, err := bind.NewKeyedTransactorWithChainID(key, chainId)
	if err != nil {
		return nil, err
	}
	return transactionOpts, nil
}

func getEthBackend(rpc string) *ethclient.Client {
	conn, err := ethclient.Dial(rpc)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	return conn
}
