package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// CarInsuranceChaincode
type CarInsuranceChaincode struct {
}

//==============================================================================================================================
//	 Participants within the system
//==============================================================================================================================
const POLICY_HOLDER = "policy_holder"
const IDENTITY_INSPECTION = "identity_inspection"
const VEHICLE_INSPECTION = "vehicle_inspection"
const CLAIM_INSPECTION = "claim_inspection"
const SETTLEMENT = "settlement"

//==============================================================================================================================
//	 States within the system
//	 Helps to track the state and perform actions accordingly
//==============================================================================================================================
const STATE_INIT_CLAIM = 0
const STATE_IDENTITY_INSPECTION = 1
const STATE_VEHICLE_INSPECTION = 2
const STATE_CLAIM_INSPECTION = 3
const STATE_SETTLEMENT = 4

func main() {
	err := shim.Start(new(CarInsuranceChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	} else {
		fmt.Printf("Simple chaincode started...")
	}
}

//==============================================================================================================================
//	 Init method for chaincode
//==============================================================================================================================
func (t *CarInsuranceChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Init..")
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1 argument.")
	}

	return nil, nil
}

//==============================================================================================================================
//	 Invoke method to invoke a chaincode function
//==============================================================================================================================
func (t *CarInsuranceChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Invoke..")

	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "createClaim" {
		return t.createClaim(stub, args)
	}

	return nil, nil
}

//==============================================================================================================================
//	 Query method for queries on blockchain
//==============================================================================================================================
func (t *CarInsuranceChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Query..")

	if function == "getClaim" { //read a variable
		return t.getClaim(stub, args)
	}

	return nil, errors.New("Received unknown function query: " + function)
}

//=================================================================================================================================
//	 createClaim - Creates a new Claim object and saves it.
//   args - IncidentDate, FirstName, LastName, Email, SSN, BirthDate, PolicyId, VIN, LicencePlateNumber
//=================================================================================================================================
func (t *CarInsuranceChaincode) createClaim(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 9 {
		return nil, errors.New("Incorrect number of arguments. IncidentDate, FirstName, LastName, Email, SSN, BirthDate, PolicyId, VIN, LicencePlateNumber required.")
	}

	var newUser = NewUser(args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8])
	var newClaim = NewClaim("", args[0], newUser)

	bytes, err := json.Marshal(newClaim)

	if err != nil {
		return nil, errors.New("Error creating new claim")
	}

	err = stub.PutState("claim", bytes)

	bytes, err = json.Marshal(STATE_INIT_CLAIM)

	if err != nil {
		return nil, errors.New("Error setting init claim state.")
	}

	// Set the state when new claim is created.
	err = stub.PutState("current_state", bytes)

	return nil, nil
}

//=================================================================================================================================
//	 getClaim - Gets claim details.
//   args - key
//=================================================================================================================================
func (t *CarInsuranceChaincode) getClaim(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, jsonResp string
	//var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get value of " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

//=================================================================================================================================
//	 getClaimStatus - Gets current state of claim process.
//=================================================================================================================================
func (t *CarInsuranceChaincode) getClaimStatus(stub shim.ChaincodeStubInterface) (int, error) {
	var jsonResp string
	var claimData Claim

	valAsbytes, err := stub.GetState("claim")
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get claim status\"}"
		return -1, errors.New(jsonResp)
	}

	err = json.Unmarshal(valAsbytes, &claimData)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to UnMarshal claim data\"}"
		return -1, errors.New(jsonResp)
	}

	return claimData.Status, nil
}

//=================================================================================================================================
//	 verifyUserIdentity - Verifies user identity (first stage).
//=================================================================================================================================
func (t *CarInsuranceChaincode) verifyUserIdentity(stub shim.ChaincodeStubInterface) ([]byte, error) {
	var claimData Claim
	var jsonResp string
	var userData, claimUser User
	var log string = ""
	userData = GetUserData()
	data, err := stub.GetState("claim")

	if err != nil {
		jsonResp = "{\"Error\":\"Failed to retrieve claim details\"}"
		return nil, errors.New(jsonResp)
	}

	err = json.Unmarshal(data, &claimData)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to UnMarshal claim data\"}"
		return nil, errors.New(jsonResp)
	}

	claimUser = claimData.UserDetails

	if userData.FirstName == claimUser.FirstName && userData.LastName == claimUser.LastName && userData.BirthDate == claimUser.BirthDate && userData.Email == claimUser.Email && userData.LicencePlateNumber == claimUser.LicencePlateNumber && userData.PolicyId == claimUser.PolicyId && userData.SSN == claimUser.SSN && userData.VIN == claimUser.VIN {
		log = log + "User Details Verified!"
	} else {
		jsonResp = "{\"Error\":\"User Identity authentication failed\"}"
		return nil, errors.New(jsonResp)
	}

	data, err = json.Marshal(log)

	if err != nil {
		return nil, errors.New("Error creating log")
	}

	return data, nil

}
