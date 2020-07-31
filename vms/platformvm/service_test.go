// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ava-labs/gecko/ids"

	"github.com/ava-labs/gecko/vms/avm"

	"github.com/ava-labs/gecko/api/keystore"
	"github.com/ava-labs/gecko/utils/crypto"
)

var (
	// Test user username
	testUsername string = "ScoobyUser"

	// Test user password, must meet minimum complexity/length requirements
	testPassword string = "ShaggyPassword1Zoinks!"

	// Bytes docoded from CB58 "ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN"
	testPrivateKey []byte = []byte{
		0x56, 0x28, 0x9e, 0x99, 0xc9, 0x4b, 0x69, 0x12,
		0xbf, 0xc1, 0x2a, 0xdc, 0x09, 0x3c, 0x9b, 0x51,
		0x12, 0x4f, 0x0d, 0xc5, 0x4a, 0xc7, 0xa7, 0x66,
		0xb2, 0xbc, 0x5c, 0xcf, 0x55, 0x8d, 0x80, 0x27,
	}

	// Platform address resulting from the above private key
	testAddress string = "P-6Y3kysjF9jnHnYkdS9yGAuoHyae2eNmeV"
)

func defaultService(t *testing.T) *Service {
	vm := defaultVM()
	vm.Ctx.Lock.Lock()
	defer vm.Ctx.Lock.Unlock()
	ks := keystore.CreateTestKeystore(t)
	if err := ks.AddUser(testUsername, testPassword); err != nil {
		t.Fatal(err)
	}
	vm.SnowmanVM.Ctx.Keystore = ks.NewBlockchainKeyStore(vm.SnowmanVM.Ctx.ChainID)
	return &Service{vm: vm}
}

// Give user [testUsername] control of [testPrivateKey] and keys[0] (which is funded)
func defaultAddress(t *testing.T, service *Service) {
	service.vm.Ctx.Lock.Lock()
	defer service.vm.Ctx.Lock.Unlock()
	userDB, err := service.vm.SnowmanVM.Ctx.Keystore.GetDatabase(testUsername, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	user := user{db: userDB}
	pk, err := service.vm.factory.ToPrivateKey(testPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	privKey := pk.(*crypto.PrivateKeySECP256K1R)
	if err := user.putAddress(privKey); err != nil {
		t.Fatal(err)
	} else if err := user.putAddress(keys[0]); err != nil {
		t.Fatal(err)
	}
}

func TestAddDefaultSubnetValidator(t *testing.T) {
	expectedJSONString := `{"startTime":"0","endTime":"0","id":"","destination":"","delegationFeeRate":"0","username":"","password":""}`
	args := AddDefaultSubnetValidatorArgs{}
	bytes, err := json.Marshal(&args)
	if err != nil {
		t.Fatal(err)
	}
	jsonString := string(bytes)
	if jsonString != expectedJSONString {
		t.Fatalf("Expected: %s\nResult: %s", expectedJSONString, jsonString)
	}
}

func TestCreateBlockchainArgsParsing(t *testing.T) {
	jsonString := `{"vmID":"lol","fxIDs":["secp256k1"], "name":"awesome", "username":"bob loblaw", "password":"yeet", "genesisData":"SkB92YpWm4Q2iPnLGCuDPZPgUQMxajqQQuz91oi3xD984f8r"}`
	args := CreateBlockchainArgs{}
	err := json.Unmarshal([]byte(jsonString), &args)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = json.Marshal(args.GenesisData); err != nil {
		t.Fatal(err)
	}
}

func TestExportKey(t *testing.T) {
	jsonString := `{"username":"ScoobyUser","password":"ShaggyPassword1Zoinks!","address":"P-6Y3kysjF9jnHnYkdS9yGAuoHyae2eNmeV"}`
	args := ExportKeyArgs{}
	err := json.Unmarshal([]byte(jsonString), &args)
	if err != nil {
		t.Fatal(err)
	}

	service := defaultService(t)
	defaultAddress(t, service)
	service.vm.Ctx.Lock.Lock()
	defer func() { service.vm.Shutdown(); service.vm.Ctx.Lock.Unlock() }()

	reply := ExportKeyReply{}
	if err := service.ExportKey(nil, &args, &reply); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(testPrivateKey, reply.PrivateKey.Bytes) {
		t.Fatalf("Expected %v, got %v", testPrivateKey, reply.PrivateKey)
	}
}

func TestImportKey(t *testing.T) {
	jsonString := `{"username":"ScoobyUser","password":"ShaggyPassword1Zoinks!","privateKey":"ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN"}`
	args := ImportKeyArgs{}
	err := json.Unmarshal([]byte(jsonString), &args)
	if err != nil {
		t.Fatal(err)
	}

	service := defaultService(t)
	service.vm.Ctx.Lock.Lock()
	defer func() { service.vm.Shutdown(); service.vm.Ctx.Lock.Unlock() }()

	reply := ImportKeyReply{}
	if err := service.ImportKey(nil, &args, &reply); err != nil {
		t.Fatal(err)
	}
	if testAddress != reply.Address {
		t.Fatalf("Expected %q, got %q", testAddress, reply.Address)
	}
}

// Test issuing a tx, having it be dropped, and then re-issued and accepted
func TestGetTxStatus(t *testing.T) {
	service := defaultService(t)
	defaultAddress(t, service)
	service.vm.Ctx.Lock.Lock()
	defer func() { service.vm.Shutdown(); service.vm.Ctx.Lock.Unlock() }()

	// create a tx
	tx, err := service.vm.newCreateChainTx(
		testSubnet1.id,
		nil,
		avm.ID,
		nil,
		"chain name",
		[]*crypto.PrivateKeySECP256K1R{testSubnet1ControlKeys[0], testSubnet1ControlKeys[1]},
	)
	if err != nil {
		t.Fatal(err)
	}
	arg := &GetTxStatusArgs{TxID: tx.ID()}
	var status Status
	if err := service.GetTxStatus(nil, arg, &status); err != nil {
		t.Fatal(err)
	} else if status != Unknown {
		t.Fatalf("status should be unknown but is %s", status)
		// put the chain in existing chain list
	} else if err := service.vm.issueTx(tx); err != nil {
		t.Fatal(err)
	} else if err := service.vm.putChains(service.vm.DB, []*DecisionTx{tx}); err != nil {
		t.Fatal(err)
	} else if _, err := service.vm.BuildBlock(); err == nil {
		t.Fatal("should have errored because chain already exists")
	} else if err := service.GetTxStatus(nil, arg, &status); err != nil {
		t.Fatal(err)
	} else if status != Dropped {
		t.Fatalf("status should be Dropped but is %s", status)
		// remove the chain from existing chain list
	} else if err := service.vm.putChains(service.vm.DB, []*DecisionTx{}); err != nil {
		t.Fatal(err)
	} else if err := service.vm.issueTx(tx); err != nil {
		t.Fatal(err)
	} else if block, err := service.vm.BuildBlock(); err != nil {
		t.Fatal(err)
	} else if blk, ok := block.(*StandardBlock); !ok {
		t.Fatalf("should be *StandardBlock but it %T", blk)
	} else if err := blk.Verify(); err != nil {
		t.Fatal(err)
	} else if err := blk.Accept(); err != nil {
		t.Fatal(err)
	} else if err := service.GetTxStatus(nil, arg, &status); err != nil {
		t.Fatal(err)
	} else if status != Committed {
		t.Fatalf("status should be Committed but is %s", status)
	}
}

// Test issuing and then retrieving a transaction
func TestGetTx(t *testing.T) {
	service := defaultService(t)
	defaultAddress(t, service)
	service.vm.Ctx.Lock.Lock()
	defer func() { service.vm.Shutdown(); service.vm.Ctx.Lock.Unlock() }()

	type test struct {
		description string
		createTx    func() (interface{}, error)
		ID          func(interface{}) ids.ID
		toBytes     func(interface{}) []byte
	}

	tests := []test{
		test{
			"standard block",
			func() (interface{}, error) {
				return service.vm.newCreateChainTx( // Test GetTx works for standard blocks
					testSubnet1.id,
					nil,
					avm.ID,
					nil,
					"chain name",
					[]*crypto.PrivateKeySECP256K1R{testSubnet1ControlKeys[0], testSubnet1ControlKeys[1]},
				)
			},
			func(tx interface{}) ids.ID { return tx.(*DecisionTx).ID() },
			func(tx interface{}) []byte {
				return tx.(*DecisionTx).UnsignedDecisionTx.(*UnsignedCreateChainTx).Bytes()
			},
		},
		test{
			"proposal block",
			func() (interface{}, error) {
				return service.vm.newAddDefaultSubnetValidatorTx( // Test GetTx works for proposal blocks
					MinimumStakeAmount,
					uint64(service.vm.clock.Time().Add(Delta).Unix()),
					uint64(service.vm.clock.Time().Add(Delta).Add(MinimumStakingDuration).Unix()),
					ids.GenerateTestShortID(),
					ids.GenerateTestShortID(),
					0,
					[]*crypto.PrivateKeySECP256K1R{keys[0]},
				)
			},
			func(tx interface{}) ids.ID { return tx.(*ProposalTx).ID() },
			func(tx interface{}) []byte {
				return tx.(*ProposalTx).UnsignedProposalTx.(*UnsignedAddDefaultSubnetValidatorTx).Bytes()
			},
		},
		test{
			"atomic block",
			func() (interface{}, error) {
				return service.vm.newExportTx( // Test GetTx works for proposal blocks
					100,
					ids.GenerateTestShortID(),
					[]*crypto.PrivateKeySECP256K1R{keys[0]},
				)
			},
			func(tx interface{}) ids.ID { return tx.(*AtomicTx).ID() },
			func(tx interface{}) []byte {
				return tx.(*AtomicTx).UnsignedAtomicTx.(*UnsignedExportTx).Bytes()
			},
		},
	}

	for _, test := range tests {
		tx, err := test.createTx()
		if err != nil {
			t.Fatalf("failed test '%s': %s", test.description, err)
		}
		arg := &GetTxArgs{TxID: test.ID(tx)}
		var response GetTxResponse
		if err := service.GetTx(nil, arg, &response); err == nil {
			t.Fatalf("failed test '%s': haven't issued tx yet so shouldn't be able to get it", test.description)
		} else if err := service.vm.issueTx(tx); err != nil {
			t.Fatalf("failed test '%s': %s", test.description, err)
		} else if block, err := service.vm.BuildBlock(); err != nil {
			t.Fatalf("failed test '%s': %s", test.description, err)
		} else if err := block.Verify(); err != nil {
			t.Fatalf("failed test '%s': %s", test.description, err)
		} else if err := block.Accept(); err != nil {
			t.Fatalf("failed test '%s': %s", test.description, err)
		} else if blk, ok := block.(*ProposalBlock); ok { // For proposal blocks, commit them
			if options, err := blk.Options(); err != nil {
				t.Fatalf("failed test '%s': %s", test.description, err)
			} else if commit, ok := options[0].(*Commit); !ok {
				t.Fatalf("failed test '%s': should prefer to commit", test.description)
			} else if err := commit.Verify(); err != nil {
				t.Fatalf("failed test '%s': %s", test.description, err)
			} else if err := commit.Accept(); err != nil {
				t.Fatalf("failed test '%s': %s", test.description, err)
			}
		} else if err := service.GetTx(nil, arg, &response); err != nil {
			t.Fatalf("failed test '%s': %s", test.description, err)
		} else if !bytes.Equal(response.RawTx.Bytes, test.toBytes(tx)) {
			t.Fatalf("failed test '%s': byte representation of tx in response is incorrect", test.description)
		}
	}
}
