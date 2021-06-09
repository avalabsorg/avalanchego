// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avm

import (
	"errors"
	"fmt"
	"github.com/ava-labs/avalanchego/database/linkeddb"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"

	"github.com/ava-labs/avalanchego/cache"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowstorm"
	"github.com/ava-labs/avalanchego/vms/components/avax"
)

var (
	errAssetIDMismatch = errors.New("asset IDs in the input don't match the utxo")
	errWrongAssetID    = errors.New("asset ID must be AVAX in the atomic tx")
	errMissingUTXO     = errors.New("missing utxo")
	errUnknownTx       = errors.New("transaction is unknown")
	errRejectedTx      = errors.New("transaction is rejected")
)

var (
	_ snowstorm.Tx    = &UniqueTx{}
	_ cache.Evictable = &UniqueTx{}
)

// UniqueTx provides a de-duplication service for txs. This only provides a
// performance boost
type UniqueTx struct {
	*TxCachedState

	vm   *VM
	txID ids.ID
}

type TxCachedState struct {
	*Tx

	unique, verifiedTx, verifiedState bool
	validity                          error

	inputs     []ids.ID
	inputUTXOs []*avax.UTXOID
	utxos      []*avax.UTXO
	deps       []snowstorm.Tx

	status choices.Status
}

func (tx *UniqueTx) refresh() {
	tx.vm.numTxRefreshes.Inc()

	if tx.TxCachedState == nil {
		tx.TxCachedState = &TxCachedState{}
	}
	if tx.unique {
		return
	}
	unique := tx.vm.state.DeduplicateTx(tx)
	prevTx := tx.Tx
	if unique == tx {
		tx.vm.numTxRefreshMisses.Inc()

		// If no one was in the cache, make sure that there wasn't an
		// intermediate object whose state I must reflect
		if status, err := tx.vm.state.GetStatus(tx.ID()); err == nil {
			tx.status = status
		}
		tx.unique = true
	} else {
		tx.vm.numTxRefreshHits.Inc()

		// If someone is in the cache, they must be up to date

		// This ensures that every unique tx object points to the same tx state
		tx.TxCachedState = unique.TxCachedState
	}

	if tx.Tx != nil {
		return
	}

	if prevTx == nil {
		if innerTx, err := tx.vm.state.GetTx(tx.ID()); err == nil {
			tx.Tx = innerTx
		}
	} else {
		tx.Tx = prevTx
	}
}

// Evict is called when this UniqueTx will no longer be returned from a cache
// lookup
func (tx *UniqueTx) Evict() {
	// Lock is already held here
	tx.unique = false
	tx.deps = nil
}

func (tx *UniqueTx) setStatus(status choices.Status) error {
	tx.refresh()
	if tx.status == status {
		return nil
	}
	tx.status = status
	return tx.vm.state.PutStatus(tx.ID(), status)
}

// ID returns the wrapped txID
func (tx *UniqueTx) ID() ids.ID       { return tx.txID }
func (tx *UniqueTx) Key() interface{} { return tx.txID }

// getAddress returns a ids.ShortID address given a transaction. This function returns either a
// ids.ShortEmpty ID or an address from the transaction object. An address is returned if the
// tx object has UTXO with a singular receiver.
// Should we check if the value of funds is > 0 too? 🤔❓
func getAddress(tx *UniqueTx) ids.ShortID {
	// go through all transfer outputs, assert they are secp transfer outputs
	// filter by threshold == 1 and address in the transfer output == 1
	for _, utxo := range tx.UTXOs() {
		out, ok := utxo.Out.(*secp256k1fx.TransferOutput)
		if !ok {
			continue
		}

		notSingularOutputReceiver := len(out.OutputOwners.Addrs) != 1 && out.OutputOwners.Threshold != 1
		if notSingularOutputReceiver {
			continue
		}

		return out.OutputOwners.Addrs[0]
	}
	// return empty
	return ids.ShortEmpty
}

// Accept is called when the transaction was finalized as accepted by consensus
func (tx *UniqueTx) Accept() error {
	fmt.Println("Transaction got accepted!")

	if s := tx.Status(); s != choices.Processing {
		tx.vm.ctx.Log.Error("Failed to accept tx %s because the tx is in state %s", tx.txID, s)
		return fmt.Errorf("transaction has invalid status: %s", s)
	}

	defer tx.vm.db.Abort()

	// Remove spent utxos
	for _, utxo := range tx.InputUTXOs() {
		if utxo.Symbolic() {
			// If the UTXO is symbolic, it can't be spent
			continue
		}
		utxoID := utxo.InputID()
		if err := tx.vm.state.DeleteUTXO(utxoID); err != nil {
			tx.vm.ctx.Log.Error("Failed to spend utxo %s due to %s", utxoID, err)
			return err
		}
	}

	// Add new utxos
	for _, utxo := range tx.UTXOs() {
		if err := tx.vm.state.PutUTXO(utxo.InputID(), utxo); err != nil {
			tx.vm.ctx.Log.Error("Failed to fund utxo %s due to %s", utxo.InputID(), err)
			return err
		}
	}

	if err := tx.setStatus(choices.Accepted); err != nil {
		tx.vm.ctx.Log.Error("Failed to accept tx %s due to %s", tx.txID, err)
		return err
	}

	txID := tx.ID()

	commitBatch, err := tx.vm.db.CommitBatch()
	if err != nil {
		tx.vm.ctx.Log.Error("Failed to calculate CommitBatch for %s due to %s", txID, err)
		return err
	}

	if err := tx.ExecuteWithSideEffects(tx.vm, commitBatch); err != nil {
		tx.vm.ctx.Log.Error("Failed to commit accept %s due to %s", txID, err)
		return err
	}

	tx.vm.ctx.Log.Verbo("Accepted Tx: %s", txID)

	// Get transaction address and proceed with indexing the transaction against the prefix DB
	// associated with the address if the returned address is not empty
	// should this be enabled on a config flag? Like indexing? 🤔❓
	address := getAddress(tx)
	if address != ids.ShortEmpty {
		addressPrefixDb := linkeddb.NewDefault(prefixdb.New(address.Bytes(), tx.vm.db))
		// []byte(txID.String()) as key vs []byte(txID.Hex()) 🤔❓
		e := addressPrefixDb.Put([]byte(txID.String()), tx.Bytes())
		if e != nil {
			tx.vm.ctx.Log.Error("Failed to save transaction to the address DB", e)
		}
	}

	tx.vm.pubsub.Publish(txID, NewPubSubFilterer(tx.Tx))
	tx.vm.walletService.decided(txID)

	tx.deps = nil // Needed to prevent a memory leak

	return nil
}

// Reject is called when the transaction was finalized as rejected by consensus
func (tx *UniqueTx) Reject() error {
	defer tx.vm.db.Abort()

	if err := tx.setStatus(choices.Rejected); err != nil {
		tx.vm.ctx.Log.Error("Failed to reject tx %s due to %s", tx.txID, err)
		return err
	}

	txID := tx.ID()
	tx.vm.ctx.Log.Debug("Rejecting Tx: %s", txID)

	if err := tx.vm.db.Commit(); err != nil {
		tx.vm.ctx.Log.Error("Failed to commit reject %s due to %s", tx.txID, err)
		return err
	}

	tx.vm.walletService.decided(txID)

	tx.deps = nil // Needed to prevent a memory leak

	return nil
}

// Status returns the current status of this transaction
func (tx *UniqueTx) Status() choices.Status {
	tx.refresh()
	return tx.status
}

// Dependencies returns the set of transactions this transaction builds on
func (tx *UniqueTx) Dependencies() []snowstorm.Tx {
	tx.refresh()
	if tx.Tx == nil || len(tx.deps) != 0 {
		return tx.deps
	}

	txIDs := ids.Set{}
	for _, in := range tx.InputUTXOs() {
		if in.Symbolic() {
			continue
		}
		txID, _ := in.InputSource()
		if txIDs.Contains(txID) {
			continue
		}
		txIDs.Add(txID)
		tx.deps = append(tx.deps, &UniqueTx{
			vm:   tx.vm,
			txID: txID,
		})
	}
	consumedIDs := tx.Tx.ConsumedAssetIDs()
	for assetID := range tx.Tx.AssetIDs() {
		if consumedIDs.Contains(assetID) || txIDs.Contains(assetID) {
			continue
		}
		txIDs.Add(assetID)
		tx.deps = append(tx.deps, &UniqueTx{
			vm:   tx.vm,
			txID: assetID,
		})
	}
	return tx.deps
}

// InputIDs returns the set of utxoIDs this transaction consumes
func (tx *UniqueTx) InputIDs() []ids.ID {
	tx.refresh()
	if tx.Tx == nil || len(tx.inputs) != 0 {
		return tx.inputs
	}

	inputUTXOs := tx.InputUTXOs()
	tx.inputs = make([]ids.ID, len(inputUTXOs))
	for i, utxo := range inputUTXOs {
		tx.inputs[i] = utxo.InputID()
	}
	return tx.inputs
}

// InputUTXOs returns the utxos that will be consumed on tx acceptance
func (tx *UniqueTx) InputUTXOs() []*avax.UTXOID {
	tx.refresh()
	if tx.Tx == nil || len(tx.inputUTXOs) != 0 {
		return tx.inputUTXOs
	}
	tx.inputUTXOs = tx.Tx.InputUTXOs()
	return tx.inputUTXOs
}

// UTXOs returns the utxos that will be added to the UTXO set on tx acceptance
func (tx *UniqueTx) UTXOs() []*avax.UTXO {
	tx.refresh()
	if tx.Tx == nil || len(tx.utxos) != 0 {
		return tx.utxos
	}
	tx.utxos = tx.Tx.UTXOs()
	return tx.utxos
}

// Bytes returns the binary representation of this transaction
func (tx *UniqueTx) Bytes() []byte {
	tx.refresh()
	return tx.Tx.Bytes()
}

func (tx *UniqueTx) verifyWithoutCacheWrites() error {
	switch status := tx.Status(); status {
	case choices.Unknown:
		return errUnknownTx
	case choices.Accepted:
		return nil
	case choices.Rejected:
		return errRejectedTx
	default:
		return tx.SemanticVerify()
	}
}

// Verify the validity of this transaction
func (tx *UniqueTx) Verify() error {
	if err := tx.verifyWithoutCacheWrites(); err != nil {
		return err
	}

	tx.verifiedState = true
	return nil
}

// SyntacticVerify verifies that this transaction is well formed
func (tx *UniqueTx) SyntacticVerify() error {
	tx.refresh()

	if tx.Tx == nil {
		return errUnknownTx
	}

	if tx.verifiedTx {
		return tx.validity
	}

	tx.verifiedTx = true
	tx.validity = tx.Tx.SyntacticVerify(
		tx.vm.ctx,
		tx.vm.codec,
		tx.vm.feeAssetID,
		tx.vm.txFee,
		tx.vm.creationTxFee,
		len(tx.vm.fxs),
	)
	return tx.validity
}

// SemanticVerify the validity of this transaction
func (tx *UniqueTx) SemanticVerify() error {
	// SyntacticVerify sets the error on validity and is checked in the next
	// statement
	_ = tx.SyntacticVerify()

	if tx.validity != nil || tx.verifiedState {
		return tx.validity
	}

	return tx.Tx.SemanticVerify(tx.vm, tx.UnsignedTx)
}
