package geth

import (
	"context"
	"crypto/ecdsa"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/gopherine/ethuni/geth/abi"

	log "github.com/sirupsen/logrus"
)

//ContractAddress for ERC20
var ContractAddress= map[string]string {
	"dai": "0xc7ad46e0b8a400bb3c915120d284aafba8fc4735",
    "maker":  "0xF9bA5210F91D0474bd1e1DcDAeC4C58E359AaD85",
	"chainlink": "0x01BE23585060835E02B77ef475b0Cc51aA1e0709",
	"uniswap_router":"0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D",
}
//Account information
type Account struct {
	Address string `json:"address"`
	PublicKey string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}
//GenerateAddress in ERC20 format
//TODO: Make the function generic to accept multiple layer1 assets
func (acc Account)GenerateAddress() Account{

	privateKey,err := crypto.GenerateKey()
	if err!=nil {
		log.Fatal(err)
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	acc.PrivateKey=hexutil.Encode(privateKeyBytes)

	publicKey:= privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	acc.PublicKey=hexutil.Encode(publicKeyBytes)[4:]

	acc.Address = crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	return acc
}

//GetBalance of the given address along with its respective token address
//TODO: return appropriate error
func (acc Account) GetBalance (asset string) (*big.Float,error){
	//For sorting decimal places
	fbalance := new(big.Float)
	account:= common.HexToAddress(acc.Address)
	if asset == "ethereum" {
		ethValue, err := Client.BalanceAt(context.Background(),account,nil)
		if err != nil {
			log.Fatal(err)
			return nil,err
		}
		
		fbalance.SetString(ethValue.String())
	} else {
		token:=common.HexToAddress(ContractAddress[asset])

		instance, err := abi.NewIERC20(token,Client)
		if err != nil {
			log.Error(err)
			return nil,err
		}
		value,err:=instance.BalanceOf(&bind.CallOpts{},account)
		//BalanceOf(&bind.CallOpts{},account
		if err != nil {
			log.Error(err)
			return nil,err
		}

		fbalance.SetString(value.String())	
	}
	
	//TODO: Fix decimal values for ERC20 if uneven
	balance := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18))) //Converting from wei to eth
	return balance,nil
}


//Swap the token pair currently only Eth to token is supported
//private key must be sent without 0x prefix
//for testing: pk:"4b3daa2c6791297826d642ba2162254a79e56f17790abcdfae8d73959f65f62b", address:"0xC33E7409705e46390344A356560bE9d1b744015A"
func Swap(asset string,privateKey string,address string) (*types.Transaction,error){
	token:=common.HexToAddress(ContractAddress["uniswap_router"])
	instance, err := abi.NewIUniswapV2Router02(token,Client)
	if err != nil {
		log.Error(err)
		return nil,err
	}

	//TODO: https://ethereum.stackexchange.com/questions/90324/what-is-uniswaps-amountoutminuint256 explore more
	Ibalance := new(big.Int).SetInt64(500)
	//TODO: Check what is the optimum delay time
	IDeadline := new(big.Int).SetInt64(20000000000000)

	weth,_:=instance.WETH(&bind.CallOpts{})
	//path pair for uniswap eg: eth->dai [eth,dai]
	pair:=[]common.Address{weth,common.HexToAddress(ContractAddress[asset])}
	chainid,err:=Client.ChainID(context.Background())
	if err!=nil{
		log.Error(err)
		return nil,err
	}

	//Generating signer - using privatekey of the address to be used to for tx
	key,err:=crypto.HexToECDSA(privateKey)
	if err!=nil{
		log.Error("Failed to create new transactor ",err)
		return nil,err
	}
	auth,err:=bind.NewKeyedTransactorWithChainID(key,chainid)
	if err !=nil {
		log.Error(err)
		return nil,err
	}

	opts := &bind.TransactOpts{
		Value:    big.NewInt(20000000000),
		Signer:   auth.Signer,
		From:     auth.From,
		GasLimit: 500000,
		GasPrice: big.NewInt(20000000000),
	}
	
	swap,err:=instance.SwapExactETHForTokens(
		opts,
		Ibalance,
		pair,
		common.HexToAddress(address),
		IDeadline)
	if err != nil {
		log.Error(err)
		return nil,err
	}
	
	return swap, nil
}


