// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avm

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/gorilla/rpc/v2"

	"github.com/ava-labs/gecko/cache"
	"github.com/ava-labs/gecko/database"
	"github.com/ava-labs/gecko/database/versiondb"
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/snow"
	"github.com/ava-labs/gecko/snow/choices"
	"github.com/ava-labs/gecko/snow/consensus/snowstorm"
	"github.com/ava-labs/gecko/snow/engine/common"
	"github.com/ava-labs/gecko/utils/codec"
	"github.com/ava-labs/gecko/utils/constants"
	"github.com/ava-labs/gecko/utils/crypto"
	"github.com/ava-labs/gecko/utils/formatting"
	"github.com/ava-labs/gecko/utils/hashing"
	"github.com/ava-labs/gecko/utils/logging"
	"github.com/ava-labs/gecko/utils/timer"
	"github.com/ava-labs/gecko/utils/wrappers"
	"github.com/ava-labs/gecko/vms/components/avax"
	"github.com/ava-labs/gecko/vms/nftfx"
	"github.com/ava-labs/gecko/vms/secp256k1fx"

	cjson "github.com/ava-labs/gecko/utils/json"
	safemath "github.com/ava-labs/gecko/utils/math"
)

const (
	batchTimeout   = time.Second
	batchSize      = 30
	stateCacheSize = 30000
	idCacheSize    = 30000
	txCacheSize    = 30000
	addressSep     = "-"
)

var (
	errIncompatibleFx            = errors.New("incompatible feature extension")
	errUnknownFx                 = errors.New("unknown feature extension")
	errGenesisAssetMustHaveState = errors.New("genesis asset must have non-empty state")
	errInvalidAddress            = errors.New("invalid address")
	errWrongBlockchainID         = errors.New("wrong blockchain ID")
	errBootstrapping             = errors.New("chain is currently bootstrapping")
)

// VM implements the avalanche.DAGVM interface
type VM struct {
	metrics
	ids.Aliaser

	avax     ids.ID
	platform ids.ID

	// Contains information of where this VM is executing
	ctx *snow.Context

	// Used to check local time
	clock timer.Clock

	codec codec.Codec

	pubsub *cjson.PubSubServer

	// State management
	state *prefixedState

	// Set to true once this VM is marked as `Bootstrapped` by the engine
	bootstrapped bool

	// fee that must be burned by every transaction
	txFee uint64

	// Transaction issuing
	timer        *timer.Timer
	batchTimeout time.Duration
	txs          []snowstorm.Tx
	toEngine     chan<- common.Message

	baseDB database.Database
	db     *versiondb.Database

	typeToFxIndex map[reflect.Type]int
	fxs           []*parsedFx
}

type codecRegistry struct {
	index         int
	typeToFxIndex map[reflect.Type]int
	codec         codec.Codec
}

func (cr *codecRegistry) RegisterType(val interface{}) error {
	valType := reflect.TypeOf(val)
	cr.typeToFxIndex[valType] = cr.index
	return cr.codec.RegisterType(val)
}
func (cr *codecRegistry) Marshal(val interface{}) ([]byte, error) { return cr.codec.Marshal(val) }
func (cr *codecRegistry) Unmarshal(b []byte, val interface{}) error {
	return cr.codec.Unmarshal(b, val)
}

/*
 ******************************************************************************
 ******************************** Avalanche API *******************************
 ******************************************************************************
 */

// Initialize implements the avalanche.DAGVM interface
func (vm *VM) Initialize(
	ctx *snow.Context,
	db database.Database,
	genesisBytes []byte,
	toEngine chan<- common.Message,
	fxs []*common.Fx,
) error {
	vm.ctx = ctx
	vm.toEngine = toEngine
	vm.baseDB = db
	vm.db = versiondb.New(db)
	vm.typeToFxIndex = map[reflect.Type]int{}
	vm.Aliaser.Initialize()

	vm.pubsub = cjson.NewPubSubServer(ctx)
	c := codec.NewDefault()

	errs := wrappers.Errs{}
	errs.Add(
		vm.metrics.Initialize(ctx.Namespace, ctx.Metrics),

		vm.pubsub.Register("accepted"),
		vm.pubsub.Register("rejected"),
		vm.pubsub.Register("verified"),

		c.RegisterType(&BaseTx{}),
		c.RegisterType(&CreateAssetTx{}),
		c.RegisterType(&OperationTx{}),
		c.RegisterType(&ImportTx{}),
		c.RegisterType(&ExportTx{}),
	)
	if errs.Errored() {
		return errs.Err
	}

	vm.fxs = make([]*parsedFx, len(fxs))
	for i, fxContainer := range fxs {
		if fxContainer == nil {
			return errIncompatibleFx
		}
		fx, ok := fxContainer.Fx.(Fx)
		if !ok {
			return errIncompatibleFx
		}
		vm.fxs[i] = &parsedFx{
			ID: fxContainer.ID,
			Fx: fx,
		}
		vm.codec = &codecRegistry{
			index:         i,
			typeToFxIndex: vm.typeToFxIndex,
			codec:         c,
		}
		if err := fx.Initialize(vm); err != nil {
			return err
		}
	}

	vm.codec = c

	vm.state = &prefixedState{
		state: &state{State: avax.State{
			Cache: &cache.LRU{Size: stateCacheSize},
			DB:    vm.db,
			Codec: vm.codec,
		}},

		tx:       &cache.LRU{Size: idCacheSize},
		utxo:     &cache.LRU{Size: idCacheSize},
		txStatus: &cache.LRU{Size: idCacheSize},

		uniqueTx: &cache.EvictableLRU{Size: txCacheSize},
	}

	if err := vm.initAliases(genesisBytes); err != nil {
		return err
	}

	if dbStatus, err := vm.state.DBInitialized(); err != nil || dbStatus == choices.Unknown {
		if err := vm.initState(genesisBytes); err != nil {
			return err
		}
	}

	vm.timer = timer.NewTimer(func() {
		ctx.Lock.Lock()
		defer ctx.Lock.Unlock()

		vm.FlushTxs()
	})
	go ctx.Log.RecoverAndPanic(vm.timer.Dispatch)
	vm.batchTimeout = batchTimeout

	return vm.db.Commit()
}

// Bootstrapping is called by the consensus engine when it starts bootstrapping
// this chain
func (vm *VM) Bootstrapping() error {
	vm.metrics.numBootstrappingCalls.Inc()

	for _, fx := range vm.fxs {
		if err := fx.Fx.Bootstrapping(); err != nil {
			return err
		}
	}
	return nil
}

// Bootstrapped is called by the consensus engine when it is done bootstrapping
// this chain
func (vm *VM) Bootstrapped() error {
	vm.metrics.numBootstrappedCalls.Inc()

	for _, fx := range vm.fxs {
		if err := fx.Fx.Bootstrapped(); err != nil {
			return err
		}
	}
	vm.bootstrapped = true
	return nil
}

// Shutdown implements the avalanche.DAGVM interface
func (vm *VM) Shutdown() error {
	if vm.timer == nil {
		return nil
	}

	// There is a potential deadlock if the timer is about to execute a timeout.
	// So, the lock must be released before stopping the timer.
	vm.ctx.Lock.Unlock()
	vm.timer.Stop()
	vm.ctx.Lock.Lock()

	return vm.baseDB.Close()
}

// CreateHandlers implements the avalanche.DAGVM interface
func (vm *VM) CreateHandlers() map[string]*common.HTTPHandler {
	vm.metrics.numCreateHandlersCalls.Inc()

	rpcServer := rpc.NewServer()
	codec := cjson.NewCodec()
	rpcServer.RegisterCodec(codec, "application/json")
	rpcServer.RegisterCodec(codec, "application/json;charset=UTF-8")
	rpcServer.RegisterService(&Service{vm: vm}, "avm") // name this service "avm"

	return map[string]*common.HTTPHandler{
		"":        {Handler: rpcServer},
		"/pubsub": {LockOptions: common.NoLock, Handler: vm.pubsub},
	}
}

// CreateStaticHandlers implements the avalanche.DAGVM interface
func (vm *VM) CreateStaticHandlers() map[string]*common.HTTPHandler {
	newServer := rpc.NewServer()
	codec := cjson.NewCodec()
	newServer.RegisterCodec(codec, "application/json")
	newServer.RegisterCodec(codec, "application/json;charset=UTF-8")
	newServer.RegisterService(&StaticService{}, "avm") // name this service "avm"
	return map[string]*common.HTTPHandler{
		"": {LockOptions: common.WriteLock, Handler: newServer},
	}
}

// PendingTxs implements the avalanche.DAGVM interface
func (vm *VM) PendingTxs() []snowstorm.Tx {
	vm.metrics.numPendingTxsCalls.Inc()

	vm.timer.Cancel()

	txs := vm.txs
	vm.txs = nil
	return txs
}

// ParseTx implements the avalanche.DAGVM interface
func (vm *VM) ParseTx(b []byte) (snowstorm.Tx, error) {
	vm.metrics.numParseTxCalls.Inc()

	return vm.parseTx(b)
}

// GetTx implements the avalanche.DAGVM interface
func (vm *VM) GetTx(txID ids.ID) (snowstorm.Tx, error) {
	vm.metrics.numGetTxCalls.Inc()

	tx := &UniqueTx{
		vm:   vm,
		txID: txID,
	}
	// Verify must be called in the case the that tx was flushed from the unique
	// cache.
	return tx, tx.Verify()
}

/*
 ******************************************************************************
 ********************************** JSON API **********************************
 ******************************************************************************
 */

// IssueTx attempts to send a transaction to consensus.
// If onDecide is specified, the function will be called when the transaction is
// either accepted or rejected with the appropriate status. This function will
// go out of scope when the transaction is removed from memory.
func (vm *VM) IssueTx(b []byte, onDecide func(choices.Status)) (ids.ID, error) {
	if !vm.bootstrapped {
		return ids.ID{}, errBootstrapping
	}
	tx, err := vm.parseTx(b)
	if err != nil {
		return ids.ID{}, err
	}
	if err := tx.Verify(); err != nil {
		return ids.ID{}, err
	}
	vm.issueTx(tx)
	tx.onDecide = onDecide
	return tx.ID(), nil
}

// GetAtomicUTXOs returns the utxos that at least one of the provided addresses is
// referenced in.
func (vm *VM) GetAtomicUTXOs(addrs ids.Set) ([]*avax.UTXO, error) {
	smDB := vm.ctx.SharedMemory.GetDatabase(vm.platform)
	defer vm.ctx.SharedMemory.ReleaseDatabase(vm.platform)

	state := avax.NewPrefixedState(smDB, vm.codec)

	utxoIDs := ids.Set{}
	for _, addr := range addrs.List() {
		utxos, _ := state.PlatformFunds(addr)
		utxoIDs.Add(utxos...)
	}

	utxos := []*avax.UTXO{}
	for _, utxoID := range utxoIDs.List() {
		utxo, err := state.PlatformUTXO(utxoID)
		if err != nil {
			return nil, err
		}
		utxos = append(utxos, utxo)
	}
	return utxos, nil
}

// GetUTXOs returns the utxos that at least one of the provided addresses is
// referenced in.
func (vm *VM) GetUTXOs(addrs ids.Set) ([]*avax.UTXO, error) {
	utxoIDs := ids.Set{}
	for _, addr := range addrs.List() {
		utxos, _ := vm.state.Funds(addr)
		utxoIDs.Add(utxos...)
	}

	utxos := []*avax.UTXO{}
	for _, utxoID := range utxoIDs.List() {
		utxo, err := vm.state.UTXO(utxoID)
		if err != nil {
			return nil, err
		}
		utxos = append(utxos, utxo)
	}
	return utxos, nil
}

/*
 ******************************************************************************
 *********************************** Fx API ***********************************
 ******************************************************************************
 */

// Clock returns a reference to the internal clock of this VM
func (vm *VM) Clock() *timer.Clock { return &vm.clock }

// Codec returns a reference to the internal codec of this VM
func (vm *VM) Codec() codec.Codec { return vm.codec }

// Logger returns a reference to the internal logger of this VM
func (vm *VM) Logger() logging.Logger { return vm.ctx.Log }

/*
 ******************************************************************************
 ********************************** Timer API *********************************
 ******************************************************************************
 */

// FlushTxs into consensus
func (vm *VM) FlushTxs() {
	vm.timer.Cancel()
	if len(vm.txs) != 0 {
		select {
		case vm.toEngine <- common.PendingTxs:
		default:
			vm.ctx.Log.Warn("Delaying issuance of transactions due to contention")
			vm.timer.SetTimeoutIn(vm.batchTimeout)
		}
	}
}

/*
 ******************************************************************************
 ********************************** Helpers ***********************************
 ******************************************************************************
 */

func (vm *VM) initAliases(genesisBytes []byte) error {
	genesis := Genesis{}
	if err := vm.codec.Unmarshal(genesisBytes, &genesis); err != nil {
		return err
	}

	for _, genesisTx := range genesis.Txs {
		if len(genesisTx.Outs) != 0 {
			return errGenesisAssetMustHaveState
		}

		tx := Tx{
			UnsignedTx: &genesisTx.CreateAssetTx,
		}
		txBytes, err := vm.codec.Marshal(&tx)
		if err != nil {
			return err
		}
		tx.Initialize(txBytes)

		txID := tx.ID()

		if err = vm.Alias(txID, genesisTx.Alias); err != nil {
			return err
		}
	}

	return nil
}

func (vm *VM) initState(genesisBytes []byte) error {
	genesis := Genesis{}
	if err := vm.codec.Unmarshal(genesisBytes, &genesis); err != nil {
		return err
	}

	for _, genesisTx := range genesis.Txs {
		if len(genesisTx.Outs) != 0 {
			return errGenesisAssetMustHaveState
		}

		tx := Tx{
			UnsignedTx: &genesisTx.CreateAssetTx,
		}
		txBytes, err := vm.codec.Marshal(&tx)
		if err != nil {
			return err
		}
		tx.Initialize(txBytes)

		txID := tx.ID()

		vm.ctx.Log.Info("Initializing with AssetID %s", txID)

		if err := vm.state.SetTx(txID, &tx); err != nil {
			return err
		}
		if err := vm.state.SetStatus(txID, choices.Accepted); err != nil {
			return err
		}
		for _, utxo := range tx.UTXOs() {
			if err := vm.state.FundUTXO(utxo); err != nil {
				return err
			}
		}
	}

	return vm.state.SetDBInitialized(choices.Processing)
}

func (vm *VM) parseTx(b []byte) (*UniqueTx, error) {
	rawTx := &Tx{}
	err := vm.codec.Unmarshal(b, rawTx)
	if err != nil {
		return nil, err
	}
	rawTx.Initialize(b)

	tx := &UniqueTx{
		TxState: &TxState{
			Tx: rawTx,
		},
		vm:   vm,
		txID: rawTx.ID(),
	}
	if err := tx.SyntacticVerify(); err != nil {
		return nil, err
	}

	if tx.Status() == choices.Unknown {
		if err := vm.state.SetTx(tx.ID(), tx.Tx); err != nil {
			return nil, err
		}
		if err := tx.setStatus(choices.Processing); err != nil {
			return nil, err
		}
		return tx, vm.db.Commit()
	}

	return tx, nil
}

func (vm *VM) issueTx(tx snowstorm.Tx) {
	vm.txs = append(vm.txs, tx)
	switch {
	case len(vm.txs) == batchSize:
		vm.FlushTxs()
	case len(vm.txs) == 1:
		vm.timer.SetTimeoutIn(vm.batchTimeout)
	}
}

func (vm *VM) getUTXO(utxoID *avax.UTXOID) (*avax.UTXO, error) {
	inputID := utxoID.InputID()
	utxo, err := vm.state.UTXO(inputID)
	if err == nil {
		return utxo, nil
	}

	inputTx, inputIndex := utxoID.InputSource()
	parent := UniqueTx{
		vm:   vm,
		txID: inputTx,
	}

	if err := parent.Verify(); err != nil {
		return nil, errMissingUTXO
	} else if status := parent.Status(); status.Decided() {
		return nil, errMissingUTXO
	}

	parentUTXOs := parent.UTXOs()
	if uint32(len(parentUTXOs)) <= inputIndex || int(inputIndex) < 0 {
		return nil, errInvalidUTXO
	}
	return parentUTXOs[int(inputIndex)], nil
}

func (vm *VM) getFx(val interface{}) (int, error) {
	valType := reflect.TypeOf(val)
	fx, exists := vm.typeToFxIndex[valType]
	if !exists {
		return 0, errUnknownFx
	}
	return fx, nil
}

func (vm *VM) verifyFxUsage(fxID int, assetID ids.ID) bool {
	tx := &UniqueTx{
		vm:   vm,
		txID: assetID,
	}
	if status := tx.Status(); !status.Fetched() {
		return false
	}
	createAssetTx, ok := tx.UnsignedTx.(*CreateAssetTx)
	if !ok {
		return false
	}
	// TODO: This could be a binary search to improve performance... Or perhaps
	// make a map
	for _, state := range createAssetTx.States {
		if state.FxID == uint32(fxID) {
			return true
		}
	}
	return false
}

// GetHRP returns the Human-Readable-Part of addresses for this VM
func (vm *VM) GetHRP() string {
	networkID := vm.ctx.NetworkID
	hrp := constants.FallbackHRP
	if _, ok := constants.NetworkIDToHRP[networkID]; ok {
		hrp = constants.NetworkIDToHRP[networkID]
	}
	return hrp
}

// ParseAddress takes in an address string and produces bytes for the address
func (vm *VM) ParseAddress(addrStr string) ([]byte, error) {
	hrp := vm.GetHRP()

	chainPrefixes := []string{vm.ctx.ChainID.String()}
	if alias, err := vm.ctx.BCLookup.PrimaryAlias(vm.ctx.ChainID); err == nil {
		chainPrefixes = append(chainPrefixes, alias)
	}
	addr, err := formatting.ParseAddress(addrStr, chainPrefixes, addressSep, hrp)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// FormatAddress takes in a 20-byte slice and produces a string for an address
func (vm *VM) FormatAddress(b []byte) (string, error) {
	hrp := vm.GetHRP()

	chainPrefix := vm.ctx.ChainID.String()
	if alias, err := vm.ctx.BCLookup.PrimaryAlias(vm.ctx.ChainID); err == nil {
		chainPrefix = alias
	}
	addrstr, err := formatting.FormatAddress(b, chainPrefix, addressSep, hrp)
	if err != nil {
		return "", err
	}
	return addrstr, nil
}

// LoadUser ...
func (vm *VM) LoadUser(
	username string,
	password string,
) (
	[]*avax.UTXO,
	*secp256k1fx.Keychain,
	error,
) {
	db, err := vm.ctx.Keystore.GetDatabase(username, password)
	if err != nil {
		return nil, nil, fmt.Errorf("problem retrieving user: %w", err)
	}

	user := userState{vm: vm}

	// The error is explicitly dropped, as it may just mean that there are no
	// addresses.
	addresses, _ := user.Addresses(db)

	addrs := ids.Set{}
	for _, addr := range addresses {
		addrs.Add(ids.NewID(hashing.ComputeHash256Array(addr.Bytes())))
	}
	utxos, err := vm.GetUTXOs(addrs)
	if err != nil {
		return nil, nil, fmt.Errorf("problem retrieving user's UTXOs: %w", err)
	}

	kc := secp256k1fx.NewKeychain()
	for _, addr := range addresses {
		sk, err := user.Key(db, addr)
		if err != nil {
			return nil, nil, fmt.Errorf("problem retrieving private key: %w", err)
		}
		kc.Add(sk)
	}

	return utxos, kc, nil
}

// Spend ...
func (vm *VM) Spend(
	utxos []*avax.UTXO,
	kc *secp256k1fx.Keychain,
	amounts map[[32]byte]uint64,
) (
	map[[32]byte]uint64,
	[]*avax.TransferableInput,
	[][]*crypto.PrivateKeySECP256K1R,
	error,
) {
	amountsSpent := make(map[[32]byte]uint64, len(amounts))
	time := vm.clock.Unix()

	ins := []*avax.TransferableInput{}
	keys := [][]*crypto.PrivateKeySECP256K1R{}
	for _, utxo := range utxos {
		assetID := utxo.AssetID()
		assetKey := assetID.Key()
		amount := amounts[assetKey]
		amountSpent := amountsSpent[assetKey]

		if amountSpent >= amount {
			// we already have enough inputs allocated to this asset
			continue
		}

		inputIntf, signers, err := kc.Spend(utxo.Out, time)
		if err != nil {
			// this utxo can't be spent with the current keys right now
			continue
		}
		input, ok := inputIntf.(avax.TransferableIn)
		if !ok {
			// this input doesn't have an amount, so I don't care about it here
			continue
		}
		newAmountSpent, err := safemath.Add64(amountSpent, input.Amount())
		if err != nil {
			// there was an error calculating the consumed amount, just error
			return nil, nil, nil, errSpendOverflow
		}
		amountsSpent[assetKey] = newAmountSpent

		// add the new input to the array
		ins = append(ins, &avax.TransferableInput{
			UTXOID: utxo.UTXOID,
			Asset:  avax.Asset{ID: assetID},
			In:     input,
		})
		// add the required keys to the array
		keys = append(keys, signers)
	}

	for asset, amount := range amounts {
		if amountsSpent[asset] < amount {
			return nil, nil, nil, errInsufficientFunds
		}
	}

	avax.SortTransferableInputsWithSigners(ins, keys)
	return amountsSpent, ins, keys, nil
}

// SpendNFT ...
func (vm *VM) SpendNFT(
	utxos []*avax.UTXO,
	kc *secp256k1fx.Keychain,
	assetID ids.ID,
	groupID uint32,
	to ids.ShortID,
) (
	[]*Operation,
	[][]*crypto.PrivateKeySECP256K1R,
	error,
) {
	time := vm.clock.Unix()

	ops := []*Operation{}
	keys := [][]*crypto.PrivateKeySECP256K1R{}

	for _, utxo := range utxos {
		// makes sure that the variable isn't overwritten with the next iteration
		utxo := utxo

		if len(ops) > 0 {
			// we have already been able to create the operation needed
			break
		}

		if !utxo.AssetID().Equals(assetID) {
			// wrong asset ID
			continue
		}
		out, ok := utxo.Out.(*nftfx.TransferOutput)
		if !ok {
			// wrong output type
			continue
		}
		if out.GroupID != groupID {
			// wrong group id
			continue
		}
		indices, signers, ok := kc.Match(&out.OutputOwners, time)
		if !ok {
			// unable to spend the output
			continue
		}

		// add the new operation to the array
		ops = append(ops, &Operation{
			Asset:   utxo.Asset,
			UTXOIDs: []*avax.UTXOID{&utxo.UTXOID},
			Op: &nftfx.TransferOperation{
				Input: secp256k1fx.Input{
					SigIndices: indices,
				},
				Output: nftfx.TransferOutput{
					GroupID: out.GroupID,
					Payload: out.Payload,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{to},
					},
				},
			},
		})
		// add the required keys to the array
		keys = append(keys, signers)
	}

	if len(ops) == 0 {
		return nil, nil, errInsufficientFunds
	}

	sortOperationsWithSigners(ops, keys, vm.codec)
	return ops, keys, nil
}

// SpendAll ...
func (vm *VM) SpendAll(
	utxos []*avax.UTXO,
	kc *secp256k1fx.Keychain,
) (
	map[[32]byte]uint64,
	[]*avax.TransferableInput,
	[][]*crypto.PrivateKeySECP256K1R,
	error,
) {
	amountsSpent := make(map[[32]byte]uint64)
	time := vm.clock.Unix()

	ins := []*avax.TransferableInput{}
	keys := [][]*crypto.PrivateKeySECP256K1R{}
	for _, utxo := range utxos {
		assetID := utxo.AssetID()
		assetKey := assetID.Key()
		amountSpent := amountsSpent[assetKey]

		inputIntf, signers, err := kc.Spend(utxo.Out, time)
		if err != nil {
			// this utxo can't be spent with the current keys right now
			continue
		}
		input, ok := inputIntf.(avax.TransferableIn)
		if !ok {
			// this input doesn't have an amount, so I don't care about it here
			continue
		}
		newAmountSpent, err := safemath.Add64(amountSpent, input.Amount())
		if err != nil {
			// there was an error calculating the consumed amount, just error
			return nil, nil, nil, errSpendOverflow
		}
		amountsSpent[assetKey] = newAmountSpent

		// add the new input to the array
		ins = append(ins, &avax.TransferableInput{
			UTXOID: utxo.UTXOID,
			Asset:  avax.Asset{ID: assetID},
			In:     input,
		})
		// add the required keys to the array
		keys = append(keys, signers)
	}

	avax.SortTransferableInputsWithSigners(ins, keys)
	return amountsSpent, ins, keys, nil
}

// Mint ...
func (vm *VM) Mint(
	utxos []*avax.UTXO,
	kc *secp256k1fx.Keychain,
	amounts map[[32]byte]uint64,
	to ids.ShortID,
) (
	[]*Operation,
	[][]*crypto.PrivateKeySECP256K1R,
	error,
) {
	time := vm.clock.Unix()

	ops := []*Operation{}
	keys := [][]*crypto.PrivateKeySECP256K1R{}

	for _, utxo := range utxos {
		// makes sure that the variable isn't overwritten with the next iteration
		utxo := utxo

		assetID := utxo.AssetID()
		assetKey := assetID.Key()
		amount := amounts[assetKey]
		if amount == 0 {
			continue
		}

		out, ok := utxo.Out.(*secp256k1fx.MintOutput)
		if !ok {
			continue
		}

		inIntf, signers, err := kc.Spend(out, time)
		if err != nil {
			continue
		}

		in, ok := inIntf.(*secp256k1fx.Input)
		if !ok {
			continue
		}

		// add the operation to the array
		ops = append(ops, &Operation{
			Asset:   utxo.Asset,
			UTXOIDs: []*avax.UTXOID{&utxo.UTXOID},
			Op: &secp256k1fx.MintOperation{
				MintInput:  *in,
				MintOutput: *out,
				TransferOutput: secp256k1fx.TransferOutput{
					Amt: amount,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{to},
					},
				},
			},
		})
		// add the required keys to the array
		keys = append(keys, signers)

		// remove the asset from the required amounts to mint
		delete(amounts, assetKey)
	}

	for _, amount := range amounts {
		if amount > 0 {
			return nil, nil, errAddressesCantMintAsset
		}
	}

	sortOperationsWithSigners(ops, keys, vm.codec)
	return ops, keys, nil
}

// MintNFT ...
func (vm *VM) MintNFT(
	utxos []*avax.UTXO,
	kc *secp256k1fx.Keychain,
	assetID ids.ID,
	payload []byte,
	to ids.ShortID,
) (
	[]*Operation,
	[][]*crypto.PrivateKeySECP256K1R,
	error,
) {
	time := vm.clock.Unix()

	ops := []*Operation{}
	keys := [][]*crypto.PrivateKeySECP256K1R{}

	for _, utxo := range utxos {
		// makes sure that the variable isn't overwritten with the next iteration
		utxo := utxo

		if len(ops) > 0 {
			// we have already been able to create the operation needed
			break
		}

		if !utxo.AssetID().Equals(assetID) {
			// wrong asset id
			continue
		}
		out, ok := utxo.Out.(*nftfx.MintOutput)
		if !ok {
			// wrong output type
			continue
		}

		indices, signers, ok := kc.Match(&out.OutputOwners, time)
		if !ok {
			// unable to spend the output
			continue
		}

		// add the operation to the array
		ops = append(ops, &Operation{
			Asset: avax.Asset{ID: assetID},
			UTXOIDs: []*avax.UTXOID{
				&utxo.UTXOID,
			},
			Op: &nftfx.MintOperation{
				MintInput: secp256k1fx.Input{
					SigIndices: indices,
				},
				GroupID: out.GroupID,
				Payload: payload,
				Outputs: []*secp256k1fx.OutputOwners{&secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{to},
				}},
			},
		})
		// add the required keys to the array
		keys = append(keys, signers)
	}

	if len(ops) == 0 {
		return nil, nil, errAddressesCantMintAsset
	}

	sortOperationsWithSigners(ops, keys, vm.codec)
	return ops, keys, nil
}

// SignSECP256K1Fx ...
func (vm *VM) SignSECP256K1Fx(tx *Tx, keys [][]*crypto.PrivateKeySECP256K1R) error {
	unsignedBytes, err := vm.codec.Marshal(&tx.UnsignedTx)
	if err != nil {
		return fmt.Errorf("problem creating transaction: %w", err)
	}
	hash := hashing.ComputeHash256(unsignedBytes)

	for _, credKeys := range keys {
		cred := &secp256k1fx.Credential{}
		for _, key := range credKeys {
			sig, err := key.SignHash(hash)
			if err != nil {
				return fmt.Errorf("problem creating transaction: %w", err)
			}
			fixedSig := [crypto.SECP256K1RSigLen]byte{}
			copy(fixedSig[:], sig)

			cred.Sigs = append(cred.Sigs, fixedSig)
		}
		tx.Creds = append(tx.Creds, cred)
	}

	b, err := vm.codec.Marshal(tx)
	if err != nil {
		return fmt.Errorf("problem creating transaction: %w", err)
	}
	tx.Initialize(b)
	return nil
}

// SignNFTFx ...
func (vm *VM) SignNFTFx(tx *Tx, keys [][]*crypto.PrivateKeySECP256K1R) error {
	unsignedBytes, err := vm.codec.Marshal(&tx.UnsignedTx)
	if err != nil {
		return fmt.Errorf("problem creating transaction: %w", err)
	}
	hash := hashing.ComputeHash256(unsignedBytes)

	for _, credKeys := range keys {
		cred := &nftfx.Credential{}
		for _, key := range credKeys {
			sig, err := key.SignHash(hash)
			if err != nil {
				return fmt.Errorf("problem creating transaction: %w", err)
			}
			fixedSig := [crypto.SECP256K1RSigLen]byte{}
			copy(fixedSig[:], sig)

			cred.Sigs = append(cred.Sigs, fixedSig)
		}
		tx.Creds = append(tx.Creds, cred)
	}

	b, err := vm.codec.Marshal(tx)
	if err != nil {
		return fmt.Errorf("problem creating transaction: %w", err)
	}
	tx.Initialize(b)
	return nil
}
