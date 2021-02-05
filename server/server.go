package main

import (
	"os"

	"github.com/gopherine/ethuni/geth"
	log "github.com/sirupsen/logrus"

	"github.com/gofiber/fiber"
)

//! TODO: Capitalizing for future file spiltting

//Key for parsing Swap
type Key struct {
	Asset string `json:"asset"`
	PrivateKey string `json:"privatekey"`
	Address string `json:"address"`
}

//GenerateAddress ERC20
func GenerateAddress(c *fiber.Ctx) {
	acc:=geth.Account{}
	account:=acc.GenerateAddress()
	c.Status(201).JSON(account)
}

//GetBalance of Ether/ERC20
func GetBalance(c *fiber.Ctx) {
	acc:=geth.Account{Address: c.Params("address")}
	asset := c.Params("asset")
	bal,err:=acc.GetBalance(asset)
	if err!=nil {
		c.Status(422).JSON(&fiber.Map{
			"error":err.Error(),
		})
		return
	} 
	c.Status(200).JSON(&fiber.Map{asset:bal})
}

//Swap token
func Swap(c *fiber.Ctx){
	s:= new(Key)
	if err := c.BodyParser(s); err != nil {
		log.Error(err)
		c.Status(422).JSON(&fiber.Map{
			"error":err.Error(),
		})	
		return
	}

	hash,err:=geth.Swap(s.Asset,s.PrivateKey,s.Address)
	if err!=nil {
		log.Error(err)
		c.Status(422).JSON(&fiber.Map{
			"error":err.Error(),
		})
		return	
	}
	
	c.Status(201).JSON(&fiber.Map{"txhash": hash.Hash().Hex()})
}

//pass testnet or mainnet as argument - testnet will initialize the program in rinkeby
//TODO: add middleware for body-validation
func main() {
	app:= fiber.New();
	connect:=geth.Init(os.Args[1])
	defer connect.Close()
	
	//SECTION - Routes localhost:3000
	app.Get("/address",GenerateAddress)
	//? /balance/ethereum/0xC33E7409705e46390344A356560bE9d1b744015A -> please use full name of the asset
	app.Get("/balance/:asset/:address",GetBalance)
	//? body sample for /swap
	/*
	* {
	*	"privatekey":"4b3daa2c6791297826d642ba2162254a79e56f17790abcdfae8d73959f65f62b",
	*	"address":"0xC33E7409705e46390344A356560bE9d1b744015A",
	*	"asset":"chainlink"
	* }
	*/
	app.Post("/swap",Swap)

	//go run server/server.go testnet 
	app.Listen(3000)
}