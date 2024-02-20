package wallet

import (
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/lucidconnect/silver-arrow/abi/KernelFactory"
	"github.com/lucidconnect/silver-arrow/core/service/erc4337"
)

func CalculateCounterFactualAddress(farcasterSigner, kernelFactoryAddress, rpc string) (string, error) {
	ownerBytes, err := hexutil.Decode(farcasterSigner)
	if err != nil {
		log.Println(err)
	}

	initCode, err := getContractInitCode(ownerBytes, common.Big0, kernelFactoryAddress)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("account initcode", hexutil.Encode(initCode))

	kernelFactory := common.HexToAddress(kernelFactoryAddress)
	backend := getEthBackend(rpc)

	accountFactoryCaller, err := KernelFactory.NewKernelFactoryCaller(kernelFactory, backend)
	if err != nil {
		log.Println(err)
	}

	walletAddress, err := accountFactoryCaller.GetAccountAddress(nil, initCode, common.Big0)
	if err != nil {
		return "", err
	}
	fmt.Println("account address", walletAddress)

	return walletAddress.Hex(), nil
}

func getEthBackend(rpc string) *ethclient.Client {
	conn, err := ethclient.Dial(rpc)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	return conn
}

func getContractInitCode(owner []byte, index *big.Int, factoryAddress string) ([]byte, error) {
	initCode := []byte{}
	implementation := os.Getenv("KERNEL_IMPLEMENTATION_ADDRESS")
	defaultValidator := os.Getenv("DEFAULT_VALIDATOR")

	kernelImplementation := common.HexToAddress(implementation)
	// fmt.Println("accountAddress ", accountAddress)

	/** inputs to createAddress
		- account implementation
		- calldata:abi.encodeWithSelector(
	 		KernelStorage.initialize.selector, defaultValidator, abi.encodePacked(owner)),
		- index
	*/

	callData, err := erc4337.EncodeKernelStorageWithSelector("initialize", common.HexToAddress(defaultValidator), owner)

	fmt.Println("callData", hexutil.Encode(callData))
	if err != nil {
		return nil, err
	}

	data := owner
	fmt.Println("enable data ", hexutil.Encode(data))
	code, err := erc4337.GetCreateAccountFnData(kernelImplementation, callData, index)
	if err != nil {
		return nil, err
	}
	factoryAddressToBytes := common.FromHex(factoryAddress)
	initCode = append(initCode, factoryAddressToBytes...)
	initCode = append(initCode, code...)

	return initCode, nil
}
