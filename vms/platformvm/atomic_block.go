// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

import (
	"errors"

	"github.com/ava-labs/gecko/database/versiondb"
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/snow/choices"
	"github.com/ava-labs/gecko/vms/components/core"
)

var (
	errConflictingParentTxs = errors.New("block contains a transaction that conflicts with a transaction in a parent block")
)

// AtomicBlock being accepted results in the transaction contained in the
// block to be accepted and committed to the chain.
type AtomicBlock struct {
	CommonDecisionBlock `serialize:"true"`

	Tx AtomicTx `serialize:"true"`

	inputs ids.Set
}

// initialize this block
func (ab *AtomicBlock) initialize(vm *VM, bytes []byte) error {
	if err := ab.CommonDecisionBlock.initialize(vm, bytes); err != nil {
		return err
	}
	txBytes, err := ab.vm.codec.Marshal(&ab.Tx)
	if err != nil {
		return err
	}
	return ab.Tx.initialize(vm, txBytes)
}

// Reject implements the snowman.Block interface
func (ab *AtomicBlock) conflicts(s ids.Set) bool {
	if ab.Status() == choices.Accepted {
		return false
	}
	if ab.inputs.Overlaps(s) {
		return true
	}
	return ab.parentBlock().conflicts(s)
}

// Verify this block performs a valid state transition.
//
// The parent block must be a proposal
//
// This function also sets onAcceptDB database if the verification passes.
func (ab *AtomicBlock) Verify() error {
	parentBlock := ab.parentBlock()

	ab.inputs = ab.Tx.InputUTXOs()

	if parentBlock.conflicts(ab.inputs) {
		return errConflictingParentTxs
	}

	// AtomicBlock is not a modifier on a proposal block, so its parent must be
	// a decision.
	parent, ok := parentBlock.(decision)
	if !ok {
		return errInvalidBlockType
	}

	pdb := parent.onAccept()

	ab.onAcceptDB = versiondb.New(pdb)
	if err := ab.Tx.SemanticVerify(ab.onAcceptDB, ab.Tx.Credentials); err != nil {
		return err
	} else if txBytes, err := ab.vm.codec.Marshal(ab.Tx); err != nil {
		return err
	} else if err := ab.vm.putTx(ab.onAcceptDB, ab.Tx.ID(), txBytes); err != nil {
		return err
	} else if err := ab.vm.putStatus(ab.onAcceptDB, ab.Tx.ID(), Committed); err != nil {
		return err
	}

	ab.vm.currentBlocks[ab.ID().Key()] = ab
	ab.parentBlock().addChild(ab)
	return nil
}

// Accept implements the snowman.Block interface
func (ab *AtomicBlock) Accept() error {
	ab.vm.Ctx.Log.Verbo("Accepting block with ID %s", ab.ID())

	if err := ab.CommonBlock.Accept(); err != nil {
		return err
	}

	// Update the state of the chain in the database
	if err := ab.onAcceptDB.Commit(); err != nil {
		ab.vm.Ctx.Log.Error("unable to commit onAcceptDB")
		return err
	}

	batch, err := ab.vm.DB.CommitBatch()
	if err != nil {
		ab.vm.Ctx.Log.Fatal("unable to commit vm's DB")
		return err
	}
	defer ab.vm.DB.Abort()

	if err := ab.Tx.Accept(batch); err != nil {
		ab.vm.Ctx.Log.Error("unable to atomically commit block")
		return err
	}

	for _, child := range ab.children {
		child.setBaseDatabase(ab.vm.DB)
	}
	if ab.onAcceptFunc != nil {
		ab.onAcceptFunc()
	}

	ab.free()
	return nil
}

// newAtomicBlock returns a new *AtomicBlock where the block's parent, a
// decision block, has ID [parentID].
func (vm *VM) newAtomicBlock(parentID ids.ID, height uint64, tx AtomicTx) (*AtomicBlock, error) {
	ab := &AtomicBlock{
		CommonDecisionBlock: CommonDecisionBlock{
			CommonBlock: CommonBlock{
				Block: core.NewBlock(parentID, height),
				vm:    vm,
			},
		},
		Tx: tx,
	}

	// We serialize this block as a Block so that it can be deserialized into a
	// Block
	blk := Block(ab)
	bytes, err := Codec.Marshal(&blk)
	if err != nil {
		return nil, err
	}
	ab.Block.Initialize(bytes, vm.SnowmanVM)
	return ab, nil
}
