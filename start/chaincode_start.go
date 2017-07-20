/*
Copyright IBM Corp 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/op/go-logging"
)

var myLogger = logging.MustGetLogger("blockInABrick_mgm")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	//calling from main
	err := shim.Start(new(SimpleChaincode))
	fmt.Println("checking in first chain code")
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	myLogger.Debug("Init Chaincode...")
	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}
	// Create Stock table
	err := stub.CreateTable("Brick_Item", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Item_ID", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "Unit_Price", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "Item_Quantity", Type: shim.ColumnDefinition_STRING, Key: false},
	})
	if err != nil {
		return nil, errors.New("Failed creating Brick_Item table.")
	}
	myLogger.Debug(" Init Done.")
	return nil, nil
}

// Invoke is our entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	myLogger.Debug("Invoke  Chaincode...")
	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "insertItem" { //generic read ledger
		return t.insertItem(stub, args)
	}
	myLogger.Debug(" Invoke Done.")

	return nil, errors.New("Received unknown function invocation: " + function)
}

//insert item in the Item table
func (t *SimpleChaincode) insertItem(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	Item_ID := args[0]
	Unit_Price := args[1]
	Item_Quantity := args[2]

	var columns []shim.Column
	idCol := shim.Column{Value: &shim.Column_String_{String_: Item_ID}}
	columns = append(columns, idCol)
	row, err := stub.GetRow("Brick_Item", columns)
	if err != nil {
		return nil, fmt.Errorf("Failed retriving associated row [%s]", err)
	}

	if len(row.Columns) == 0 {
		// Insert row
		ok, err := stub.InsertRow("Brick_Item", shim.Row{
			Columns: []*shim.Column{
				&shim.Column{Value: &shim.Column_String_{String_: Item_ID}},
				&shim.Column{Value: &shim.Column_String_{String_: Unit_Price}},
				&shim.Column{Value: &shim.Column_String_{String_: Item_Quantity}},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("Failed inserting row [%s]", err)
		}
		if !ok {
			return nil, errors.New("Failed inserting row.")
		}

	} else {
		// Update row
		ok, err := stub.ReplaceRow("Brick_Item", shim.Row{
			Columns: []*shim.Column{
				&shim.Column{Value: &shim.Column_String_{String_: row.Columns[0].GetString_() + " " + Item_ID}},
				&shim.Column{Value: &shim.Column_String_{String_: row.Columns[1].GetString_() + " " + Unit_Price}},
				&shim.Column{Value: &shim.Column_String_{String_: row.Columns[2].GetString_() + " " + Item_Quantity}},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("Failed replacing row [%s]", err)
		}
		if !ok {
			return nil, errors.New("Failed replacing row.")
		}
	}

	return nil, err
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "readItem" {
		return t.readItem(stub, args)
	}

	fmt.Println("query did not find func: " + function) //error

	return nil, errors.New("Received unknown function query: " + function)
}

//read from stub.
func (t *SimpleChaincode) readItem(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	if len(args) != 1 {
		return nil, errors.New("Expecting one argument as key")
	}
	item_code := args[0]
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: item_code}}
	columns = append(columns, col1)

	row, err := stub.GetRow("Brick_Item", columns)
	if err != nil {
		myLogger.Debugf("Failed retriving item_code [%s]: [%s]", string(item_code), err)
		return nil, fmt.Errorf("Failed retriving item_code [%s]: [%s]", string(item_code), err)
	}

	myLogger.Debugf("Query done [% x]", row.Columns[1].GetBytes())

	return row.Columns[1].GetBytes(), nil
}
