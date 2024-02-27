package frame

import (
	"fmt"

	"github.com/google/uuid"

	// "github.com/lucidconnect/silver-arrow/abi/KernelFactory"
	// "github.com/lucidconnect/silver-arrow/core/service/erc4337"
	"gorm.io/gorm"
)

/**
Frame variables:
	item-id
	image

Frame constants:
	button-id

*/

type Button string

var (
	ClaimButton       Button = "claim"
	RefreshBotton     Button = "refresh"
	TransactionButton Button = "view tx"
	PromptButton      Button = "make your own @ inverse.xyz"
)

type ClaimFrame struct {
	ID                uuid.UUID `gorm:"primaryKey"`
	ItemId            string
	ImageUrl          string
	CollectionAddress string
}

func CreateClaimFrame(itemId, imageUrl, collectionAddr string, db *gorm.DB) (string, error) {
	id := uuid.New()
	frame := ClaimFrame{
		ID:                id,
		ItemId:            itemId,
		ImageUrl:          imageUrl,
		CollectionAddress: collectionAddr,
	}

	if err := db.Create(frame).Error; err != nil {
		return "", err
	}

	return id.String(), nil
}

func GetFrameDetails(id string, db *gorm.DB) (*ClaimFrame, error) {
	var frameDetails *ClaimFrame
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	if err := db.Where("id = ?", uid).First(&frameDetails).Error; err != nil {
		return nil, err
	}

	return frameDetails, nil
}

func ParseFrame(imageUrl string, title Button) string {
	var frame string
	switch title {
	case ClaimButton:
		frame = fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<meta name="description" content="This is a simple Go web application that returns HTML meta tags.">
			<meta property="og:image" content="%v">
			<meta property="fc:frame" content="vNext" />
			<meta property="fc:frame:image" content="%v" />
			<meta property="fc:frame:button:1" content="%v" />
			<title></title>
		</head>
		<body>
			<h1>Inverse</h1>
		</body>
		</html>
		`, imageUrl, imageUrl, title)
	case "refresh":
	case "view tx":
	case PromptButton:
		// on redirect, server should respond with a 302 and redirect to a set url
		url := "https://7806-2a09-bac5-4dd6-d2-00-15-36d.ngrok-free.app/f4a76b5e-6616-491f-a846-b1a811a3de94?claimed=true"
		frame = fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<meta name="description" content="This is a simple Go web application that returns HTML meta tags.">
			<meta property="og:image" content="%v">
			<meta property="fc:frame" content="vNext" />
			<meta property="fc:frame:image" content="%v" />
			<meta property="fc:frame:button:1" content="%v" />
			<meta property="fc:frame:button:1:action" content="post_redirect" />
			<meta property="fc:frame:post_url" content="%v" />
			<title></title>
		</head>
		<body>
			<h1>Inverse</h1>
		</body>
		</html>
		`, imageUrl, imageUrl, title, url)

	}

	return frame
}

func ParseFrameAction(btn Button, item, verifiedAddress string) (string, error) {
	var response string
	fmt.Println(btn)
	switch btn {
	case ClaimButton:
		// err := ClaimItem(item, verifiedAddress)
		// if err != nil {
		// 	return "", err
		// }
		response = "claim succesful"
	case RefreshBotton:
		// refresh
	case TransactionButton:
		// view transaction
	case PromptButton:
		// return a 302
		response = "https://inverse.wtf"
	}
	return response, nil
}

// func CalculateCounterFactualAddress(farcasterSigner, kernelFactoryAddress, rpc string) (string, error) {
// 	ownerBytes, err := hexutil.Decode(farcasterSigner)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	initCode, err := getContractInitCode(ownerBytes, common.Big0, kernelFactoryAddress)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	fmt.Println("account initcode", hexutil.Encode(initCode))

// 	kernelFactory := common.HexToAddress(kernelFactoryAddress)
// 	backend := getEthBackend(rpc)

// 	accountFactoryCaller, err := KernelFactory.NewKernelFactoryCaller(kernelFactory, backend)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	walletAddress, err := accountFactoryCaller.GetAccountAddress(nil, initCode, common.Big0)
// 	if err != nil {
// 		return "", err
// 	}
// 	fmt.Println("account address", walletAddress)

// 	return walletAddress.Hex(), nil
// }

// func getEthBackend(rpc string) *ethclient.Client {
// 	conn, err := ethclient.Dial(rpc)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
// 	}
// 	return conn
// }

// func getContractInitCode(owner []byte, index *big.Int, factoryAddress string) ([]byte, error) {
// 	initCode := []byte{}
// 	implementation := os.Getenv("KERNEL_IMPLEMENTATION_ADDRESS")
// 	defaultValidator := os.Getenv("DEFAULT_VALIDATOR")

// 	kernelImplementation := common.HexToAddress(implementation)
// 	// fmt.Println("accountAddress ", accountAddress)

// 	/** inputs to createAddress
// 		- account implementation
// 		- calldata:abi.encodeWithSelector(
// 	 		KernelStorage.initialize.selector, defaultValidator, abi.encodePacked(owner)),
// 		- index
// 	*/

// 	callData, err := erc4337.EncodeKernelStorageWithSelector("initialize", common.HexToAddress(defaultValidator), owner)

// 	fmt.Println("callData", hexutil.Encode(callData))
// 	if err != nil {
// 		return nil, err
// 	}

// 	data := owner
// 	fmt.Println("enable data ", hexutil.Encode(data))
// 	code, err := erc4337.GetCreateAccountFnData(kernelImplementation, callData, index)
// 	if err != nil {
// 		return nil, err
// 	}
// 	factoryAddressToBytes := common.FromHex(factoryAddress)
// 	initCode = append(initCode, factoryAddressToBytes...)
// 	initCode = append(initCode, code...)

// 	return initCode, nil
// }
