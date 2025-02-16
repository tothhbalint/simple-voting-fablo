package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Vote struct {
	VoterID   string `json:"voterID"`
	Candidate string `json:"candidate"`
}

type Candidate struct {
	ID   string `json:"candidateID"`
	Name string `json:"name"`
}

type VotingContract struct {
	contractapi.Contract
}

func (v *VotingContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// example data to fill the ledger with candidates
	candidates := []Candidate{
		{ID: "1", Name: "Alice"},
		{ID: "2", Name: "Bob"},
		{ID: "3", Name: "Charlie"},
	}

	for _, candidate := range candidates {
		candidateJSON, err := json.Marshal(candidate)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState("CANDIDATE_"+candidate.ID, candidateJSON)
		if err != nil {
			return fmt.Errorf("failed to put candidate to world state: %v", err)
		}
	}

	return nil
}

// maybe getCandidate instead, check if candidate is real
func (v *VotingContract) validateCandidate(ctx contractapi.TransactionContextInterface, candidateID string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState("CANDIDATE_" + candidateID)
	if err != nil {
		return false, err
	}
	return assetJSON != nil, nil
}

func (v *VotingContract) GetCandidates(ctx contractapi.TransactionContextInterface) ([]Candidate, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("CANDIDATE_", "CANDIDATE_~")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var candidates []Candidate
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var candidate Candidate
		err = json.Unmarshal(queryResponse.Value, &candidate)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

func (v *VotingContract) CastVote(ctx contractapi.TransactionContextInterface, voterID string, candidateID string) error {
	// check if voter has voted
	existingVote, err := ctx.GetStub().GetState(voterID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if existingVote != nil {
		return fmt.Errorf("this voter: %s has already cast a vote", voterID)
	}

	// cast the vote if not
	exists, err := v.validateCandidate(ctx, candidateID)
	if err != nil {
		return fmt.Errorf("failed to read from world state:%v", err)
	}
	if !exists {
		return fmt.Errorf("the candidate %s does not exist", candidateID)
	}

	vote := Vote{
		VoterID:   voterID,
		Candidate: candidateID,
	}
	voteJSON, _ := json.Marshal(vote)
	//insert the vote
	return ctx.GetStub().PutState(voterID, voteJSON)
}

func (v *VotingContract) CountVotes(ctx contractapi.TransactionContextInterface) (map[string]int, error) {
	//	candidates, err := v.GetCandidates(ctx)
	//	if err != nil {
	//		return nil, fmt.Errorf("failed to read from world state:%v", err)
	//	}
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	vote_count := make(map[string]int)
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(queryResponse.Key, "CANDIDATE_") {
			continue
		}

		var vote Vote
		err = json.Unmarshal(queryResponse.Value, &vote)
		if err != nil {
			return nil, err
		}
		candidateData, err := ctx.GetStub().GetState("CANDIDATE_" + vote.Candidate)
		if err != nil {
			return nil, err
		}

		var candidateMap map[string]interface{}

		err = json.Unmarshal(candidateData, &candidateMap)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal candidate")
		}

		name, ok := candidateMap["name"].(string)
		if !ok {
			return nil, fmt.Errorf("name field is missing")
		}

		vote_count[name]++
	}
	return vote_count, nil
}
