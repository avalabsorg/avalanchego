// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

// func TestBuildGenesisInvalidUTXOBalance(t *testing.T) {
// 	id, _ := ids.ShortFromString("8CrVPQZ4VSqgL8zTdvL14G8HqAfrBr4z")
// 	utxo := APIUTXO{
// 		Address: id,
// 		Amount:  0,
// 	}
// 	weight := json.Uint64(987654321)
// 	validator := APIDefaultSubnetValidator{
// 		APIValidator: APIValidator{
// 			EndTime: 15,
// 			Weight:  &weight,
// 			ID:      id,
// 		},
// 		Destination: id,
// 	}

// 	args := BuildGenesisArgs{
// 		UTXOs: []APIUTXO{
// 			utxo,
// 		},
// 		Validators: []APIDefaultSubnetValidator{
// 			validator,
// 		},
// 		Time: 5,
// 	}
// 	reply := BuildGenesisReply{}

// 	ss := StaticService{}
// 	if err := ss.BuildGenesis(nil, &args, &reply); err == nil {
// 		t.Fatalf("Should have errored due to an invalid balance")
// 	}
// }

// func TestBuildGenesisInvalidAmount(t *testing.T) {
// 	id, _ := ids.ShortFromString("8CrVPQZ4VSqgL8zTdvL14G8HqAfrBr4z")
// 	utxo := APIUTXO{
// 		Address: id,
// 		Amount:  123456789,
// 	}
// 	weight := json.Uint64(0)
// 	validator := APIDefaultSubnetValidator{
// 		APIValidator: APIValidator{
// 			StartTime: 0,
// 			EndTime:   15,
// 			Weight:    &weight,
// 			ID:        id,
// 		},
// 		Destination: id,
// 	}

// 	args := BuildGenesisArgs{
// 		UTXOs: []APIUTXO{
// 			utxo,
// 		},
// 		Validators: []APIDefaultSubnetValidator{
// 			validator,
// 		},
// 		Time: 5,
// 	}
// 	reply := BuildGenesisReply{}

// 	ss := StaticService{}
// 	if err := ss.BuildGenesis(nil, &args, &reply); err == nil {
// 		t.Fatalf("Should have errored due to an invalid amount")
// 	}
// }

// func TestBuildGenesisInvalidEndtime(t *testing.T) {
// 	id, _ := ids.ShortFromString("8CrVPQZ4VSqgL8zTdvL14G8HqAfrBr4z")
// 	utxo := APIUTXO{
// 		Address: id,
// 		Amount:  123456789,
// 	}

// 	weight := json.Uint64(987654321)
// 	validator := APIDefaultSubnetValidator{
// 		APIValidator: APIValidator{
// 			StartTime: 0,
// 			EndTime:   5,
// 			Weight:    &weight,
// 			ID:        id,
// 		},
// 		Destination: id,
// 	}

// 	args := BuildGenesisArgs{
// 		UTXOs: []APIUTXO{
// 			utxo,
// 		},
// 		Validators: []APIDefaultSubnetValidator{
// 			validator,
// 		},
// 		Time: 5,
// 	}
// 	reply := BuildGenesisReply{}

// 	ss := StaticService{}
// 	if err := ss.BuildGenesis(nil, &args, &reply); err == nil {
// 		t.Fatalf("Should have errored due to an invalid end time")
// 	}
// }

// func TestBuildGenesisReturnsSortedValidators(t *testing.T) {
// 	id := ids.NewShortID([20]byte{1})
// 	utxo := APIUTXO{
// 		Address: id,
// 		Amount:  123456789,
// 	}

// 	weight := json.Uint64(987654321)
// 	validator1 := APIDefaultSubnetValidator{
// 		APIValidator: APIValidator{
// 			StartTime: 0,
// 			EndTime:   20,
// 			Weight:    &weight,
// 			ID:        id,
// 		},
// 		Destination: id,
// 	}

// 	validator2 := APIDefaultSubnetValidator{
// 		APIValidator: APIValidator{
// 			StartTime: 3,
// 			EndTime:   15,
// 			Weight:    &weight,
// 			ID:        id,
// 		},
// 		Destination: id,
// 	}

// 	validator3 := APIDefaultSubnetValidator{
// 		APIValidator: APIValidator{
// 			StartTime: 1,
// 			EndTime:   10,
// 			Weight:    &weight,
// 			ID:        id,
// 		},
// 		Destination: id,
// 	}

// 	args := BuildGenesisArgs{
// 		AvaxAssetID: ids.NewID([32]byte{'d', 'u', 'm', 'm', 'y', ' ', 'I', 'D'}),
// 		UTXOs: []APIUTXO{
// 			utxo,
// 		},
// 		Validators: []APIDefaultSubnetValidator{
// 			validator1,
// 			validator2,
// 			validator3,
// 		},
// 		Time: 5,
// 	}
// 	reply := BuildGenesisReply{}

// 	ss := StaticService{}
// 	if err := ss.BuildGenesis(nil, &args, &reply); err != nil {
// 		t.Fatalf("BuildGenesis should not have errored but got error: %s", err)
// 	}

// 	genesis := &Genesis{}
// 	if err := Codec.Unmarshal(reply.Bytes.Bytes, genesis); err != nil {
// 		t.Fatal(err)
// 	}
// 	validators := genesis.Validators
// 	if validators.Len() == 0 {
// 		t.Fatal("Validators should contain 3 validators")
// 	}
// 	currentValidator := validators.Remove()
// 	for validators.Len() > 0 {
// 		nextValidator := validators.Remove()
// 		if currentValidator.EndTime().Unix() > nextValidator.EndTime().Unix() {
// 			t.Fatalf("Validators returned by genesis should be a min heap sorted by end time")
// 		}
// 		currentValidator = nextValidator
// 	}
// }
