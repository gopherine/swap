package geth

import (
	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/ethclient"
)

//Client declared
var Client *ethclient.Client
//Init initiaizes eth infura client based on selection
func Init(network string) *ethclient.Client {
	var err error

	if network=="testnet" {
		Client, err = ethclient.Dial("https://rinkeby.infura.io/v3/9269365f61d246f2864880ed5bf2b610")
	} else  {
	
		//Should be replaced by mainnet endpoint
		Client, err = ethclient.Dial("https://mainnet.infura.io/v3/9269365f61d246f2864880ed5bf2b610")
	}
	if err!=nil {
		log.Fatal(err)
	}
	
	log.Info("We have a connection ! on "+network)
	_=Client
	return Client
}