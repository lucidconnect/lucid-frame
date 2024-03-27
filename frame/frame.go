package frame

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
	MintButton        Button = "mint"
	ClaimButton       Button = "claim"
	RefreshBotton     Button = "refresh"
	TransactionButton Button = "view tx"
	PromptButton      Button = "make your own"
)

type ClaimFrame struct {
	ID                uuid.UUID `gorm:"primaryKey"`
	ItemId            string    // TODO: itemId will now be dropId
	DropId            string
	ImageUrl          string
	CollectionAddress string
}

type Drop struct {
	Base
	CreatorID              uuid.UUID
	CreatorAddress         string
	Name                   string
	Image                  string `json:"image"`
	Thumbnail              string `json:"thumbnail"`
	Description            string `json:"description"`
	AAContractAddress      *string
	TransactionHash        *string
	AAWalletDeploymentHash *string
	BlockchainNetwork      *string
	Featured               bool `gorm:"default:false"`
	MintUrl                string
	MintPrice              *float64
	GasIsCreatorSponsored  bool
}

type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (base *Base) BeforeCreate(scope *gorm.DB) error {
	base.ID = uuid.New()
	return nil
}

func CreateClaimFrame(dropId, imageUrl, collectionAddr string, db *gorm.DB) (string, error) {
	// check if a frame already exists
	frameDetail, err := GetFrameByDropId(dropId, db)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			id := uuid.New()
			frame := ClaimFrame{
				ID:                id,
				ItemId:            "",
				DropId:            dropId,
				ImageUrl:          imageUrl,
				CollectionAddress: collectionAddr,
			}
			err := db.Create(frame).Error
			if err != nil {
				log.Println(err)
				return "", err
			}
			return id.String(), nil
		}
	}

	return frameDetail.ID.String(), nil
}

func GetFrameDetails(id string, db *gorm.DB) (*ClaimFrame, error) {
	var frameDetails *ClaimFrame
	uid, err := uuid.Parse(id)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if err := db.Where("id = ?", uid).First(&frameDetails).Error; err != nil {
		return nil, err
	}

	return frameDetails, nil
}

func GetDropDetails(id string, db *gorm.DB) (*Drop, error) {
	var drop *Drop
	fmt.Println("uuid ", id)
	uid, err := uuid.Parse(id)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if err := db.Where("id = ?", uid).First(&drop).Error; err != nil {
		return nil, err
	}

	return drop, nil
}

func GetFrameByDropId(dropId string, db *gorm.DB) (*ClaimFrame, error) {
	var frameDetails *ClaimFrame
	if err := db.Where("drop_id = ?", dropId).First(&frameDetails).Error; err != nil {
		return nil, err
	}

	return frameDetails, nil
}

func FrameToExternalClaim(w http.ResponseWriter, imageUrl, id string) {
	frame := fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<meta name="description" content="luciddrops.xyz">
				<meta property="og:image" content="%v">
				<meta property="fc:frame" content="vNext" />
				<meta property="fc:frame:image" content="%v" />
				<meta property="fc:frame:button:1" content="%v" />
				<meta property="fc:frame:button:1:action" content="post_redirect" />
				<title></title>
			</head>
			<body>
				<h1>Lucid Drops</h1>
			</body>
			</html>
			`, imageUrl, imageUrl, MintButton)
	fmt.Fprint(w, frame)
}

func returnClaimFrame(imageUrl string) string {
	frame := fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<meta name="description" content="luciddrops.xyz">
				<meta property="og:image" content="%v">
				<meta property="fc:frame" content="vNext" />
				<meta property="fc:frame:image" content="%v" />
				<meta property="fc:frame:button:1" content="%v" />
				<title></title>
			</head>
			<body>
				<h1>Lucid Drops</h1>
			</body>
			</html>
			`, imageUrl, imageUrl, ClaimButton)
	return frame
}

func returnTransactionSuccessFrame(imageUrl, frameId, tx string) string {
	baseUrl := os.Getenv("BASE_URL")
	landingPage := os.Getenv("LUCID_LANDING_PAGE")

	txUrl := fmt.Sprintf("%v/frame/%v?tx=%v", baseUrl, frameId, tx)
	landingPageButton := Button(fmt.Sprintf("%v - %v", PromptButton, landingPage))
	url := fmt.Sprintf("%v/frame/%v?claimed=true", baseUrl, frameId)

	frame := fmt.Sprintf(`
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
				<meta property="fc:frame:button:2" content="%v" />
				<meta property="fc:frame:button:2:action" content="post_redirect" />
				<meta property="fc:frame:post_url" content="%v" />
				<title></title>
			</head>
			<body>
				<h1>Inverse</h1>
			</body>
			</html>
			`, imageUrl, imageUrl, TransactionButton, txUrl, landingPageButton, url)
	return frame
}

func returnMintLimitFrame(imageUrl, frameId string) string {
	baseUrl := os.Getenv("BASE_URL")
	url := fmt.Sprintf("%v/frame/%v?claimed=true", baseUrl, frameId)
	frame := fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<meta name="description" content="luciddrops.xyz">
				<meta property="og:image" content="%v">
				<meta property="fc:frame" content="vNext" />
				<meta property="fc:frame:image" content="%v" />
				<meta property="fc:frame:button:1" content="%v" />
				<meta property="fc:frame:button:1:action" content="post_redirect" />
				<meta property="fc:frame:post_url" content="%v" />
				<title></title>
			</head>
			<body>
				<h1>Lucid Drops</h1>
			</body>
			</html>
			`, imageUrl, imageUrl, PromptButton, url)
	return frame
}
func ParseFrame(imageUrl, frameId string, tx string, buttons ...Button) string {
	var frame string
	// for i, title := range buttons {
	switch buttons[0] {
	case ClaimButton:
		frame = fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<meta name="description" content="luciddrops.xyz">
				<meta property="og:image" content="%v">
				<meta property="fc:frame" content="vNext" />
				<meta property="fc:frame:image" content="%v" />
				<meta property="fc:frame:button:1" content="%v" />
				<title></title>
			</head>
			<body>
				<h1>Lucid Drops</h1>
			</body>
			</html>
			`, imageUrl, imageUrl, buttons[0])
	case TransactionButton:
		// on redirect, server should respond with a 302 and redirect to a set url
		landingPage := os.Getenv("LUCID_LANDING_PAGE")
		txButton := buttons[0]
		landingPageButton := Button(fmt.Sprintf("%v - %v", buttons[1], landingPage))
		baseUrl := os.Getenv("BASE_URL")
		url := fmt.Sprintf("%v/frame/%v?claimed=true", baseUrl, frameId)
		txUrl := fmt.Sprintf("%v/frame/%v?tx=%v", baseUrl, frameId, tx)
		// "https://7806-2a09-bac5-4dd6-d2-00-15-36d.ngrok-free.app/f4a76b5e-6616-491f-a846-b1a811a3de94?claimed=true"
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
				<meta property="fc:frame:button:2" content="%v" />
				<meta property="fc:frame:button:2:action" content="post_redirect" />
				<meta property="fc:frame:post_url" content="%v" />
				<title></title>
			</head>
			<body>
				<h1>Inverse</h1>
			</body>
			</html>
			`, imageUrl, imageUrl, txButton, txUrl, landingPageButton, url)
	case PromptButton:
		baseUrl := os.Getenv("BASE_URL")
		url := fmt.Sprintf("%v/frame/%v?claimed=true", baseUrl, frameId)
		frame = fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<meta name="description" content="luciddrops.xyz">
				<meta property="og:image" content="%v">
				<meta property="fc:frame" content="vNext" />
				<meta property="fc:frame:image" content="%v" />
				<meta property="fc:frame:button:1" content="%v" />
				<meta property="fc:frame:button:1:action" content="post_redirect" />
				<meta property="fc:frame:post_url" content="%v" />
				<title></title>
			</head>
			<body>
				<h1>Lucid Drops</h1>
			</body>
			</html>
			`, imageUrl, imageUrl, buttons[0], url)
	}
	return frame
}

func ParseFrameAction(btn Button, drop, verifiedAddress string) (string, error) {
	var response string
	fmt.Println(btn)
	switch btn {
	case ClaimButton:
		tx, err := ClaimItem(drop, verifiedAddress)
		if err != nil {
			return "", err
		}
		response = tx
	case RefreshBotton:
		// refresh
	case TransactionButton:
		// view transaction
		response = os.Getenv("BLOCK_EXPLORER")
	case PromptButton:
		// return a 302
		response = os.Getenv("LUCID_LANDING_PAGE")
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
