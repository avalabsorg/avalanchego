// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avm

import (
	"context"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/formatting/addressconverter"
	"github.com/ava-labs/avalanchego/utils/rpc"

	cjson "github.com/ava-labs/avalanchego/utils/json"
)

const chainIDAlias = "X"

// Interface compliance
var _ Client = &client{}

// Client for interacting with an AVM (X-Chain) instance
type Client interface {
	WalletClient
	// GetTxStatus returns the status of [txID]
	GetTxStatus(ctx context.Context, txID ids.ID, options ...rpc.Option) (choices.Status, error)
	// ConfirmTx attempts to confirm [txID] by repeatedly checking its status.
	// Note: ConfirmTx will block until either the context is done or the client
	//       returns a decided status.
	ConfirmTx(ctx context.Context, txID ids.ID, freq time.Duration, options ...rpc.Option) (choices.Status, error)
	// GetTx returns the byte representation of [txID]
	GetTx(ctx context.Context, txID ids.ID, options ...rpc.Option) ([]byte, error)
	// IssueStopVertex issues a stop vertex.
	IssueStopVertex(ctx context.Context, options ...rpc.Option) error
	// GetUTXOs returns the byte representation of the UTXOs controlled by [addrs]
	GetUTXOs(
		ctx context.Context,
		addrs []ids.ShortID,
		limit uint32,
		startAddress ids.ShortID,
		startUTXOID ids.ID,
		options ...rpc.Option,
	) ([][]byte, ids.ShortID, ids.ID, error)
	// GetAtomicUTXOs returns the byte representation of the atomic UTXOs controlled by [addrs]
	// from [sourceChain]
	GetAtomicUTXOs(
		ctx context.Context,
		addrs []ids.ShortID,
		sourceChain string,
		limit uint32,
		startAddress ids.ShortID,
		startUTXOID ids.ID,
		options ...rpc.Option,
	) ([][]byte, ids.ShortID, ids.ID, error)
	// GetAssetDescription returns a description of [assetID]
	GetAssetDescription(ctx context.Context, assetID string, options ...rpc.Option) (*GetAssetDescriptionReply, error)
	// GetBalance returns the balance of [assetID] held by [addr].
	// If [includePartial], balance includes partial owned (i.e. in a multisig) funds.
	GetBalance(ctx context.Context, addr ids.ShortID, assetID string, includePartial bool, options ...rpc.Option) (*GetBalanceReply, error)
	// GetAllBalances returns all asset balances for [addr]
	// CreateAsset creates a new asset and returns its assetID
	GetAllBalances(context.Context, ids.ShortID, bool, ...rpc.Option) ([]Balance, error)
	CreateAsset(
		ctx context.Context,
		user api.UserPass,
		from []ids.ShortID,
		changeAddr ids.ShortID,
		name string,
		symbol string,
		denomination byte,
		holders []*ClientHolder,
		minters []ClientOwners,
		options ...rpc.Option,
	) (ids.ID, error)
	// CreateFixedCapAsset creates a new fixed cap asset and returns its assetID
	CreateFixedCapAsset(
		ctx context.Context,
		user api.UserPass,
		from []ids.ShortID,
		changeAddr ids.ShortID,
		name string,
		symbol string,
		denomination byte,
		holders []*ClientHolder,
		options ...rpc.Option,
	) (ids.ID, error)
	// CreateVariableCapAsset creates a new variable cap asset and returns its assetID
	CreateVariableCapAsset(
		ctx context.Context,
		user api.UserPass,
		from []ids.ShortID,
		changeAddr ids.ShortID,
		name string,
		symbol string,
		denomination byte,
		minters []ClientOwners,
		options ...rpc.Option,
	) (ids.ID, error)
	// CreateNFTAsset creates a new NFT asset and returns its assetID
	CreateNFTAsset(
		ctx context.Context,
		user api.UserPass,
		from []ids.ShortID,
		changeAddr ids.ShortID,
		name string,
		symbol string,
		minters []ClientOwners,
		options ...rpc.Option,
	) (ids.ID, error)
	// CreateAddress creates a new address controlled by [user]
	CreateAddress(ctx context.Context, user api.UserPass, options ...rpc.Option) (ids.ShortID, error)
	// ListAddresses returns all addresses on this chain controlled by [user]
	ListAddresses(ctx context.Context, user api.UserPass, options ...rpc.Option) ([]ids.ShortID, error)
	// ExportKey returns the private key corresponding to [addr] controlled by [user]
	ExportKey(ctx context.Context, user api.UserPass, addr ids.ShortID, options ...rpc.Option) (string, error)
	// ImportKey imports [privateKey] to [user]
	ImportKey(ctx context.Context, user api.UserPass, privateKey string, options ...rpc.Option) (ids.ShortID, error)
	// Mint [amount] of [assetID] to be owned by [to]
	Mint(
		ctx context.Context,
		user api.UserPass,
		from []ids.ShortID,
		changeAddr ids.ShortID,
		amount uint64,
		assetID string,
		to ids.ShortID,
		options ...rpc.Option,
	) (ids.ID, error)
	// SendNFT sends an NFT and returns the ID of the newly created transaction
	SendNFT(
		ctx context.Context,
		user api.UserPass,
		from []ids.ShortID,
		changeAddr ids.ShortID,
		assetID string,
		groupID uint32,
		to ids.ShortID,
		options ...rpc.Option,
	) (ids.ID, error)
	// MintNFT issues a MintNFT transaction and returns the ID of the newly created transaction
	MintNFT(
		ctx context.Context,
		user api.UserPass,
		from []ids.ShortID,
		changeAddr ids.ShortID,
		assetID string,
		payload []byte,
		to ids.ShortID,
		options ...rpc.Option,
	) (ids.ID, error)
	// Import sends an import transaction to import funds from [sourceChain] and
	// returns the ID of the newly created transaction
	Import(ctx context.Context, user api.UserPass, to ids.ShortID, sourceChain string, options ...rpc.Option) (ids.ID, error) // Export sends an asset from this chain to the P/C-Chain.
	// After this tx is accepted, the AVAX must be imported to the P/C-chain with an importTx.
	// Returns the ID of the newly created atomic transaction
	Export(
		ctx context.Context,
		user api.UserPass,
		from []ids.ShortID,
		changeAddr ids.ShortID,
		amount uint64,
		to ids.ShortID,
		toChainIDAlias string,
		assetID string,
		options ...rpc.Option,
	) (ids.ID, error)
}

// implementation for an AVM client for interacting with avm [chain]
type client struct {
	requester rpc.EndpointRequester
	// used for address ID -> string conversion
	hrp string
}

// NewClient returns an AVM client for interacting with avm [chain]
func NewClient(uri string, chain string, networkID uint32) Client {
	return &client{
		requester: rpc.NewEndpointRequester(uri, fmt.Sprintf("/ext/%s", constants.ChainAliasPrefix+chain), "avm"),
		hrp:       constants.GetHRP(networkID),
	}
}

func (c *client) IssueTx(ctx context.Context, txBytes []byte, options ...rpc.Option) (ids.ID, error) {
	txStr, err := formatting.EncodeWithChecksum(formatting.Hex, txBytes)
	if err != nil {
		return ids.ID{}, err
	}
	res := &api.JSONTxID{}
	err = c.requester.SendRequest(ctx, "issueTx", &api.FormattedTx{
		Tx:       txStr,
		Encoding: formatting.Hex,
	}, res, options...)
	return res.TxID, err
}

func (c *client) IssueStopVertex(ctx context.Context, options ...rpc.Option) error {
	return c.requester.SendRequest(ctx, "issueStopVertex", &struct{}{}, &struct{}{}, options...)
}

func (c *client) GetTxStatus(ctx context.Context, txID ids.ID, options ...rpc.Option) (choices.Status, error) {
	res := &GetTxStatusReply{}
	err := c.requester.SendRequest(ctx, "getTxStatus", &api.JSONTxID{
		TxID: txID,
	}, res, options...)
	return res.Status, err
}

func (c *client) ConfirmTx(ctx context.Context, txID ids.ID, freq time.Duration, options ...rpc.Option) (choices.Status, error) {
	ticker := time.NewTicker(freq)
	defer ticker.Stop()

	for {
		status, err := c.GetTxStatus(ctx, txID, options...)
		if err == nil {
			if status.Decided() {
				return status, nil
			}
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return status, ctx.Err()
		}
	}
}

func (c *client) GetTx(ctx context.Context, txID ids.ID, options ...rpc.Option) ([]byte, error) {
	res := &api.FormattedTx{}
	err := c.requester.SendRequest(ctx, "getTx", &api.GetTxArgs{
		TxID:     txID,
		Encoding: formatting.Hex,
	}, res, options...)
	if err != nil {
		return nil, err
	}

	txBytes, err := formatting.Decode(res.Encoding, res.Tx)
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}

func (c *client) GetUTXOs(
	ctx context.Context,
	addrs []ids.ShortID,
	limit uint32,
	startAddress ids.ShortID,
	startUTXOID ids.ID,
	options ...rpc.Option,
) ([][]byte, ids.ShortID, ids.ID, error) {
	return c.GetAtomicUTXOs(ctx, addrs, "", limit, startAddress, startUTXOID, options...)
}

func (c *client) GetAtomicUTXOs(
	ctx context.Context,
	addrs []ids.ShortID,
	sourceChain string,
	limit uint32,
	startAddress ids.ShortID,
	startUTXOID ids.ID,
	options ...rpc.Option,
) ([][]byte, ids.ShortID, ids.ID, error) {
	res := &api.GetUTXOsReply{}
	err := c.requester.SendRequest(ctx, "getUTXOs", &api.GetUTXOsArgs{
		Addresses:   ids.ShortIDSliceToStringSlice(addrs),
		SourceChain: sourceChain,
		Limit:       cjson.Uint32(limit),
		StartIndex: api.Index{
			Address: startAddress.String(),
			UTXO:    startUTXOID.String(),
		},
		Encoding: formatting.Hex,
	}, res, options...)
	if err != nil {
		return nil, ids.ShortID{}, ids.Empty, err
	}

	utxos := make([][]byte, len(res.UTXOs))
	for i, utxo := range res.UTXOs {
		utxoBytes, err := formatting.Decode(res.Encoding, utxo)
		if err != nil {
			return nil, ids.ShortID{}, ids.Empty, err
		}
		utxos[i] = utxoBytes
	}
	endAddr, err := addressconverter.ParseAddressToID(res.EndIndex.Address)
	if err != nil {
		return nil, ids.ShortID{}, ids.Empty, err
	}
	endUTXOID, err := ids.FromString(res.EndIndex.UTXO)
	if err != nil {
		return nil, ids.ShortID{}, ids.Empty, err
	}
	return utxos, endAddr, endUTXOID, nil
}

func (c *client) GetAssetDescription(ctx context.Context, assetID string, options ...rpc.Option) (*GetAssetDescriptionReply, error) {
	res := &GetAssetDescriptionReply{}
	err := c.requester.SendRequest(ctx, "getAssetDescription", &GetAssetDescriptionArgs{
		AssetID: assetID,
	}, res, options...)
	return res, err
}

func (c *client) GetBalance(
	ctx context.Context,
	addr ids.ShortID,
	assetID string,
	includePartial bool,
	options ...rpc.Option,
) (*GetBalanceReply, error) {
	res := &GetBalanceReply{}
	err := c.requester.SendRequest(ctx, "getBalance", &GetBalanceArgs{
		Address:        addr.String(),
		AssetID:        assetID,
		IncludePartial: includePartial,
	}, res, options...)
	return res, err
}

func (c *client) GetAllBalances(
	ctx context.Context,
	addr ids.ShortID,
	includePartial bool,
	options ...rpc.Option,
) ([]Balance, error) {
	res := &GetAllBalancesReply{}
	err := c.requester.SendRequest(ctx, "getAllBalances", &GetAllBalancesArgs{
		JSONAddress:    api.JSONAddress{Address: addr.String()},
		IncludePartial: includePartial,
	}, res, options...)
	return res.Balances, err
}

// ClientHolder describes how much an address owns of an asset
type ClientHolder struct {
	Amount  uint64
	Address ids.ShortID
}

// ClientOwners describes who can perform an action
type ClientOwners struct {
	Threshold uint32
	Minters   []ids.ShortID
}

func (c *client) CreateAsset(
	ctx context.Context,
	user api.UserPass,
	from []ids.ShortID,
	changeAddr ids.ShortID,
	name string,
	symbol string,
	denomination byte,
	clientHolders []*ClientHolder,
	clientMinters []ClientOwners,
	options ...rpc.Option,
) (ids.ID, error) {
	res := &FormattedAssetID{}
	var err error
	holders := make([]*Holder, len(clientHolders))
	for i, clientHolder := range clientHolders {
		holders[i].Amount = cjson.Uint64(clientHolder.Amount)
		holders[i].Address = clientHolder.Address.String()
	}
	minters := make([]Owners, len(clientMinters))
	for i, clientMinter := range clientMinters {
		minters[i].Threshold = cjson.Uint32(clientMinter.Threshold)
		minters[i].Minters = ids.ShortIDSliceToStringSlice(clientMinter.Minters)
	}
	err = c.requester.SendRequest(ctx, "createAsset", &CreateAssetArgs{
		JSONSpendHeader: api.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  api.JSONFromAddrs{From: ids.ShortIDSliceToStringSlice(from)},
			JSONChangeAddr: api.JSONChangeAddr{ChangeAddr: changeAddr.String()},
		},
		Name:           name,
		Symbol:         symbol,
		Denomination:   denomination,
		InitialHolders: holders,
		MinterSets:     minters,
	}, res, options...)
	return res.AssetID, err
}

func (c *client) CreateFixedCapAsset(
	ctx context.Context,
	user api.UserPass,
	from []ids.ShortID,
	changeAddr ids.ShortID,
	name string,
	symbol string,
	denomination byte,
	clientHolders []*ClientHolder,
	options ...rpc.Option,
) (ids.ID, error) {
	res := &FormattedAssetID{}
	fromStr, err := addressconverter.FormatAddressesFromID(chainIDAlias, c.hrp, from)
	if err != nil {
		return ids.Empty, err
	}
	changeAddrStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, changeAddr[:])
	if err != nil {
		return ids.Empty, err
	}
	holders := make([]*Holder, len(clientHolders))
	for i, clientHolder := range clientHolders {
		holders[i].Amount = cjson.Uint64(clientHolder.Amount)
		holders[i].Address, err = formatting.FormatAddress(chainIDAlias, c.hrp, clientHolder.Address[:])
		if err != nil {
			return ids.Empty, err
		}
	}
	err = c.requester.SendRequest(ctx, "createAsset", &CreateAssetArgs{
		JSONSpendHeader: api.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  api.JSONFromAddrs{From: fromStr},
			JSONChangeAddr: api.JSONChangeAddr{ChangeAddr: changeAddrStr},
		},
		Name:           name,
		Symbol:         symbol,
		Denomination:   denomination,
		InitialHolders: holders,
	}, res, options...)
	return res.AssetID, err
}

func (c *client) CreateVariableCapAsset(
	ctx context.Context,
	user api.UserPass,
	from []ids.ShortID,
	changeAddr ids.ShortID,
	name string,
	symbol string,
	denomination byte,
	clientMinters []ClientOwners,
	options ...rpc.Option,
) (ids.ID, error) {
	res := &FormattedAssetID{}
	fromStr, err := addressconverter.FormatAddressesFromID(chainIDAlias, c.hrp, from)
	if err != nil {
		return ids.Empty, err
	}
	changeAddrStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, changeAddr[:])
	if err != nil {
		return ids.Empty, err
	}
	minters := make([]Owners, len(clientMinters))
	for i, clientMinter := range clientMinters {
		minters[i].Threshold = cjson.Uint32(clientMinter.Threshold)
		minters[i].Minters, err = addressconverter.FormatAddressesFromID(chainIDAlias, c.hrp, clientMinter.Minters)
		if err != nil {
			return ids.Empty, err
		}
	}
	err = c.requester.SendRequest(ctx, "createAsset", &CreateAssetArgs{
		JSONSpendHeader: api.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  api.JSONFromAddrs{From: fromStr},
			JSONChangeAddr: api.JSONChangeAddr{ChangeAddr: changeAddrStr},
		},
		Name:         name,
		Symbol:       symbol,
		Denomination: denomination,
		MinterSets:   minters,
	}, res, options...)
	return res.AssetID, err
}

func (c *client) CreateNFTAsset(
	ctx context.Context,
	user api.UserPass,
	from []ids.ShortID,
	changeAddr ids.ShortID,
	name string,
	symbol string,
	clientMinters []ClientOwners,
	options ...rpc.Option,
) (ids.ID, error) {
	res := &FormattedAssetID{}
	fromStr, err := addressconverter.FormatAddressesFromID(chainIDAlias, c.hrp, from)
	if err != nil {
		return ids.Empty, err
	}
	changeAddrStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, changeAddr[:])
	if err != nil {
		return ids.Empty, err
	}
	minters := make([]Owners, len(clientMinters))
	for i, clientMinter := range clientMinters {
		minters[i].Threshold = cjson.Uint32(clientMinter.Threshold)
		minters[i].Minters, err = addressconverter.FormatAddressesFromID(chainIDAlias, c.hrp, clientMinter.Minters)
		if err != nil {
			return ids.Empty, err
		}
	}
	err = c.requester.SendRequest(ctx, "createNFTAsset", &CreateNFTAssetArgs{
		JSONSpendHeader: api.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  api.JSONFromAddrs{From: fromStr},
			JSONChangeAddr: api.JSONChangeAddr{ChangeAddr: changeAddrStr},
		},
		Name:       name,
		Symbol:     symbol,
		MinterSets: minters,
	}, res, options...)
	return res.AssetID, err
}

func (c *client) CreateAddress(ctx context.Context, user api.UserPass, options ...rpc.Option) (ids.ShortID, error) {
	res := &api.JSONAddress{}
	err := c.requester.SendRequest(ctx, "createAddress", &user, res, options...)
	if err != nil {
		return ids.ShortID{}, err
	}
	return addressconverter.ParseAddressToID(res.Address)
}

func (c *client) ListAddresses(ctx context.Context, user api.UserPass, options ...rpc.Option) ([]ids.ShortID, error) {
	res := &api.JSONAddresses{}
	err := c.requester.SendRequest(ctx, "listAddresses", &user, res, options...)
	if err != nil {
		return nil, err
	}
	return addressconverter.ParseAddressesToID(res.Addresses)
}

func (c *client) ExportKey(ctx context.Context, user api.UserPass, addr ids.ShortID, options ...rpc.Option) (string, error) {
	res := &ExportKeyReply{}
	addrStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, addr[:])
	if err != nil {
		return "", err
	}
	err = c.requester.SendRequest(ctx, "exportKey", &ExportKeyArgs{
		UserPass: user,
		Address:  addrStr,
	}, res, options...)
	return res.PrivateKey, err
}

func (c *client) ImportKey(ctx context.Context, user api.UserPass, privateKey string, options ...rpc.Option) (ids.ShortID, error) {
	res := &api.JSONAddress{}
	err := c.requester.SendRequest(ctx, "importKey", &ImportKeyArgs{
		UserPass:   user,
		PrivateKey: privateKey,
	}, res, options...)
	if err != nil {
		return ids.ShortID{}, err
	}
	return addressconverter.ParseAddressToID(res.Address)
}

func (c *client) Send(
	ctx context.Context,
	user api.UserPass,
	from []ids.ShortID,
	changeAddr ids.ShortID,
	amount uint64,
	assetID string,
	to ids.ShortID,
	memo string,
	options ...rpc.Option,
) (ids.ID, error) {
	res := &api.JSONTxID{}
	fromStr, err := addressconverter.FormatAddressesFromID(chainIDAlias, c.hrp, from)
	if err != nil {
		return ids.Empty, err
	}
	changeAddrStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, changeAddr[:])
	if err != nil {
		return ids.Empty, err
	}
	toStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, to[:])
	if err != nil {
		return ids.Empty, err
	}
	err = c.requester.SendRequest(ctx, "send", &SendArgs{
		JSONSpendHeader: api.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  api.JSONFromAddrs{From: fromStr},
			JSONChangeAddr: api.JSONChangeAddr{ChangeAddr: changeAddrStr},
		},
		SendOutput: SendOutput{
			Amount:  cjson.Uint64(amount),
			AssetID: assetID,
			To:      toStr,
		},
		Memo: memo,
	}, res, options...)
	return res.TxID, err
}

func (c *client) SendMultiple(
	ctx context.Context,
	user api.UserPass,
	from []ids.ShortID,
	changeAddr ids.ShortID,
	clientOutputs []ClientSendOutput,
	memo string,
	options ...rpc.Option,
) (ids.ID, error) {
	res := &api.JSONTxID{}
	fromStr, err := addressconverter.FormatAddressesFromID(chainIDAlias, c.hrp, from)
	if err != nil {
		return ids.Empty, err
	}
	changeAddrStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, changeAddr[:])
	if err != nil {
		return ids.Empty, err
	}
	outputs := make([]SendOutput, len(clientOutputs))
	for i, clientOutput := range clientOutputs {
		outputs[i].Amount = cjson.Uint64(clientOutput.Amount)
		outputs[i].AssetID = clientOutput.AssetID
		outputs[i].To, err = formatting.FormatAddress(chainIDAlias, c.hrp, clientOutput.To[:])
		if err != nil {
			return ids.Empty, err
		}
	}
	err = c.requester.SendRequest(ctx, "sendMultiple", &SendMultipleArgs{
		JSONSpendHeader: api.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  api.JSONFromAddrs{From: fromStr},
			JSONChangeAddr: api.JSONChangeAddr{ChangeAddr: changeAddrStr},
		},
		Outputs: outputs,
		Memo:    memo,
	}, res, options...)
	return res.TxID, err
}

func (c *client) Mint(
	ctx context.Context,
	user api.UserPass,
	from []ids.ShortID,
	changeAddr ids.ShortID,
	amount uint64,
	assetID string,
	to ids.ShortID,
	options ...rpc.Option,
) (ids.ID, error) {
	res := &api.JSONTxID{}
	fromStr, err := addressconverter.FormatAddressesFromID(chainIDAlias, c.hrp, from)
	if err != nil {
		return ids.Empty, err
	}
	changeAddrStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, changeAddr[:])
	if err != nil {
		return ids.Empty, err
	}
	toStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, to[:])
	if err != nil {
		return ids.Empty, err
	}
	err = c.requester.SendRequest(ctx, "mint", &MintArgs{
		JSONSpendHeader: api.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  api.JSONFromAddrs{From: fromStr},
			JSONChangeAddr: api.JSONChangeAddr{ChangeAddr: changeAddrStr},
		},
		Amount:  cjson.Uint64(amount),
		AssetID: assetID,
		To:      toStr,
	}, res, options...)
	return res.TxID, err
}

func (c *client) SendNFT(
	ctx context.Context,
	user api.UserPass,
	from []ids.ShortID,
	changeAddr ids.ShortID,
	assetID string,
	groupID uint32,
	to ids.ShortID,
	options ...rpc.Option,
) (ids.ID, error) {
	res := &api.JSONTxID{}
	fromStr, err := addressconverter.FormatAddressesFromID(chainIDAlias, c.hrp, from)
	if err != nil {
		return ids.Empty, err
	}
	changeAddrStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, changeAddr[:])
	if err != nil {
		return ids.Empty, err
	}
	toStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, to[:])
	if err != nil {
		return ids.Empty, err
	}
	err = c.requester.SendRequest(ctx, "sendNFT", &SendNFTArgs{
		JSONSpendHeader: api.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  api.JSONFromAddrs{From: fromStr},
			JSONChangeAddr: api.JSONChangeAddr{ChangeAddr: changeAddrStr},
		},
		AssetID: assetID,
		GroupID: cjson.Uint32(groupID),
		To:      toStr,
	}, res, options...)
	return res.TxID, err
}

func (c *client) MintNFT(
	ctx context.Context,
	user api.UserPass,
	from []ids.ShortID,
	changeAddr ids.ShortID,
	assetID string,
	payload []byte,
	to ids.ShortID,
	options ...rpc.Option,
) (ids.ID, error) {
	payloadStr, err := formatting.EncodeWithChecksum(formatting.Hex, payload)
	if err != nil {
		return ids.ID{}, err
	}
	res := &api.JSONTxID{}
	fromStr, err := addressconverter.FormatAddressesFromID(chainIDAlias, c.hrp, from)
	if err != nil {
		return ids.Empty, err
	}
	changeAddrStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, changeAddr[:])
	if err != nil {
		return ids.Empty, err
	}
	toStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, to[:])
	if err != nil {
		return ids.Empty, err
	}
	err = c.requester.SendRequest(ctx, "mintNFT", &MintNFTArgs{
		JSONSpendHeader: api.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  api.JSONFromAddrs{From: fromStr},
			JSONChangeAddr: api.JSONChangeAddr{ChangeAddr: changeAddrStr},
		},
		AssetID:  assetID,
		Payload:  payloadStr,
		To:       toStr,
		Encoding: formatting.Hex,
	}, res, options...)
	return res.TxID, err
}

func (c *client) Import(ctx context.Context, user api.UserPass, to ids.ShortID, sourceChain string, options ...rpc.Option) (ids.ID, error) {
	res := &api.JSONTxID{}
	toStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, to[:])
	if err != nil {
		return ids.Empty, err
	}
	err = c.requester.SendRequest(ctx, "import", &ImportArgs{
		UserPass:    user,
		To:          toStr,
		SourceChain: sourceChain,
	}, res, options...)
	return res.TxID, err
}

func (c *client) Export(
	ctx context.Context,
	user api.UserPass,
	from []ids.ShortID,
	changeAddr ids.ShortID,
	amount uint64,
	to ids.ShortID,
	toChainIDAlias string,
	assetID string,
	options ...rpc.Option,
) (ids.ID, error) {
	res := &api.JSONTxID{}
	fromStr, err := addressconverter.FormatAddressesFromID(chainIDAlias, c.hrp, from)
	if err != nil {
		return ids.Empty, err
	}
	changeAddrStr, err := formatting.FormatAddress(chainIDAlias, c.hrp, changeAddr[:])
	if err != nil {
		return ids.Empty, err
	}
	toStr, err := formatting.FormatAddress(toChainIDAlias, c.hrp, to[:])
	if err != nil {
		return ids.Empty, err
	}
	err = c.requester.SendRequest(ctx, "export", &ExportArgs{
		JSONSpendHeader: api.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  api.JSONFromAddrs{From: fromStr},
			JSONChangeAddr: api.JSONChangeAddr{ChangeAddr: changeAddrStr},
		},
		Amount:  cjson.Uint64(amount),
		To:      toStr,
		AssetID: assetID,
	}, res, options...)
	return res.TxID, err
}
