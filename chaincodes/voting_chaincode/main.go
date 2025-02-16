package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	votingCode, err := contractapi.NewChaincode(&VotingContract{})

	if err != nil {
		log.Panicf("Error creating voting-chaincode: %v", err)
	}

	if err := votingCode.Start(); err != nil {
		log.Panicf("Error starting voting-chaincode: %v", err)
	}
}
