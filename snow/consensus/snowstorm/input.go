// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowstorm

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/snow"
	"github.com/ava-labs/gecko/snow/consensus/snowball"
	"github.com/ava-labs/gecko/snow/events"
	"github.com/ava-labs/gecko/utils/formatting"
	"github.com/ava-labs/gecko/utils/wrappers"
)

// InputFactory implements Factory by returning an input struct
type InputFactory struct{}

// New implements Factory
func (InputFactory) New() Consensus { return &Input{} }

// Input is an implementation of a multi-color, non-transitive, snowball
// instance
type Input struct {
	metrics

	ctx    *snow.Context
	params snowball.Parameters

	// preferences is the set of consumerIDs that have only in edges
	// virtuous is the set of consumerIDs that have no edges
	preferences, virtuous, virtuousVoting ids.Set

	txs    map[[32]byte]txNode    // Map consumerID -> consumerNode
	inputs map[[32]byte]inputNode // Map inputID -> inputNode

	pendingAccept, pendingReject events.Blocker

	time uint64

	// Number of times RecordPoll has been called
	currentVote int

	errs wrappers.Errs
}

type txNode struct {
	bias int
	tx   Tx

	timestamp uint64
}

type inputNode struct {
	bias, confidence, lastVote int
	rogue                      bool
	preference                 ids.ID
	color                      ids.ID
	conflicts                  ids.Set
}

// Initialize implements the ConflictGraph interface
func (ig *Input) Initialize(ctx *snow.Context, params snowball.Parameters) {
	ctx.Log.AssertDeferredNoError(params.Valid)

	ig.ctx = ctx
	ig.params = params

	if err := ig.metrics.Initialize(ctx.Log, params.Namespace, params.Metrics); err != nil {
		ig.ctx.Log.Error("%s", err)
	}

	ig.txs = make(map[[32]byte]txNode)
	ig.inputs = make(map[[32]byte]inputNode)
}

// Parameters implements the Snowstorm interface
func (ig *Input) Parameters() snowball.Parameters { return ig.params }

// IsVirtuous implements the ConflictGraph interface
func (ig *Input) IsVirtuous(tx Tx) bool {
	id := tx.ID()
	for _, consumption := range tx.InputIDs().List() {
		input := ig.inputs[consumption.Key()]
		if input.rogue ||
			(input.conflicts.Len() > 0 && !input.conflicts.Contains(id)) {
			return false
		}
	}
	return true
}

// Add implements the ConflictGraph interface
func (ig *Input) Add(tx Tx) error {
	if ig.Issued(tx) {
		return nil // Already inserted
	}

	txID := tx.ID()
	bytes := tx.Bytes()

	ig.ctx.DecisionDispatcher.Issue(ig.ctx.ChainID, txID, bytes)
	inputs := tx.InputIDs()
	// If there are no inputs, they are vacuously accepted
	if inputs.Len() == 0 {
		if err := tx.Accept(); err != nil {
			return err
		}
		ig.ctx.DecisionDispatcher.Accept(ig.ctx.ChainID, txID, bytes)
		ig.metrics.Issued(txID)
		ig.metrics.Accepted(txID)
		return nil
	}

	cn := txNode{tx: tx}
	virtuous := true
	// If there are inputs, they must be voted on
	for _, consumption := range inputs.List() {
		consumptionKey := consumption.Key()
		input, exists := ig.inputs[consumptionKey]
		input.rogue = exists // If the input exists for a conflict
		if exists {
			for _, conflictID := range input.conflicts.List() {
				ig.virtuous.Remove(conflictID)
				ig.virtuousVoting.Remove(conflictID)
			}
		} else {
			input.preference = txID // If there isn't a conflict, I'm preferred
		}
		input.conflicts.Add(txID)
		ig.inputs[consumptionKey] = input

		virtuous = virtuous && !exists
	}

	// Add the node to the set
	ig.txs[txID.Key()] = cn
	if virtuous {
		// If I'm preferred in all my conflict sets, I'm preferred.
		// Because the preference graph is a DAG, there will always be at least
		// one preferred consumer, if there is a consumer
		ig.preferences.Add(txID)
		ig.virtuous.Add(txID)
		ig.virtuousVoting.Add(txID)
	}
	ig.metrics.Issued(txID)

	toReject := &inputRejector{
		ig: ig,
		tn: cn,
	}

	for _, dependency := range tx.Dependencies() {
		if !dependency.Status().Decided() {
			toReject.deps.Add(dependency.ID())
		}
	}
	ig.pendingReject.Register(toReject)
	return ig.errs.Err
}

// Issued implements the ConflictGraph interface
func (ig *Input) Issued(tx Tx) bool {
	if tx.Status().Decided() {
		return true
	}
	_, ok := ig.txs[tx.ID().Key()]
	return ok
}

// Virtuous implements the ConflictGraph interface
func (ig *Input) Virtuous() ids.Set { return ig.virtuous }

// Preferences implements the ConflictGraph interface
func (ig *Input) Preferences() ids.Set { return ig.preferences }

// Conflicts implements the ConflictGraph interface
func (ig *Input) Conflicts(tx Tx) ids.Set {
	id := tx.ID()
	conflicts := ids.Set{}

	for _, input := range tx.InputIDs().List() {
		inputNode := ig.inputs[input.Key()]
		conflicts.Union(inputNode.conflicts)
	}

	conflicts.Remove(id)
	return conflicts
}

// RecordPoll implements the ConflictGraph interface
func (ig *Input) RecordPoll(votes ids.Bag) error {
	ig.currentVote++

	votes.SetThreshold(ig.params.Alpha)
	threshold := votes.Threshold()
	for _, toInc := range threshold.List() {
		incKey := toInc.Key()
		tx, exist := ig.txs[incKey]
		if !exist {
			// Votes for decided consumptions are ignored
			continue
		}

		tx.bias++

		// The timestamp is needed to ensure correctness in the case that a
		// consumer was rejected from a conflict set, when it was preferred in
		// this conflict set, when there is a tie for the second highest
		// confidence.
		ig.time++
		tx.timestamp = ig.time

		preferred := true
		rogue := false
		confidence := ig.params.BetaRogue

		consumptions := tx.tx.InputIDs().List()
		for _, inputID := range consumptions {
			inputKey := inputID.Key()
			input := ig.inputs[inputKey]

			// If I did not receive a vote in the last vote, reset my confidence to 0
			if input.lastVote+1 != ig.currentVote {
				input.confidence = 0
			}
			input.lastVote = ig.currentVote

			// check the snowflake preference
			if !toInc.Equals(input.color) {
				input.confidence = 0
			}
			// update the snowball preference
			if tx.bias > input.bias {
				// if the previous preference lost it's preference in this
				// input, it can't be preferred in all the inputs
				ig.preferences.Remove(input.preference)

				input.bias = tx.bias
				input.preference = toInc
			}

			// update snowflake vars
			input.color = toInc
			input.confidence++

			ig.inputs[inputKey] = input

			// track cumulative statistics
			preferred = preferred && toInc.Equals(input.preference)
			rogue = rogue || input.rogue
			if confidence > input.confidence {
				confidence = input.confidence
			}
		}

		// If the node wasn't accepted, but was preferred, make sure it is
		// marked as preferred
		if preferred {
			ig.preferences.Add(toInc)
		}

		if (!rogue && confidence >= ig.params.BetaVirtuous) ||
			confidence >= ig.params.BetaRogue {
			ig.deferAcceptance(tx)
			if ig.errs.Errored() {
				return ig.errs.Err
			}
			continue
		}

		ig.txs[incKey] = tx
	}
	return ig.errs.Err
}

func (ig *Input) deferAcceptance(tn txNode) {
	toAccept := &inputAccepter{
		ig: ig,
		tn: tn,
	}

	for _, dependency := range tn.tx.Dependencies() {
		if !dependency.Status().Decided() {
			toAccept.deps.Add(dependency.ID())
		}
	}

	ig.virtuousVoting.Remove(tn.tx.ID())
	ig.pendingAccept.Register(toAccept)
}

// reject all the ids and remove them from their conflict sets
func (ig *Input) reject(ids ...ids.ID) error {
	for _, conflict := range ids {
		conflictKey := conflict.Key()
		cn := ig.txs[conflictKey]
		delete(ig.txs, conflictKey)
		ig.preferences.Remove(conflict) // A rejected value isn't preferred

		// Remove from all conflict sets
		ig.removeConflict(conflict, cn.tx.InputIDs().List()...)

		// Mark it as rejected
		if err := cn.tx.Reject(); err != nil {
			return err
		}
		ig.ctx.DecisionDispatcher.Reject(ig.ctx.ChainID, cn.tx.ID(), cn.tx.Bytes())
		ig.metrics.Rejected(conflict)
		ig.pendingAccept.Abandon(conflict)
		ig.pendingReject.Fulfill(conflict)
	}
	return nil
}

// Remove id from all of its conflict sets
func (ig *Input) removeConflict(id ids.ID, inputIDs ...ids.ID) {
	for _, inputID := range inputIDs {
		inputKey := inputID.Key()
		// if the input doesn't exists, it was already decided
		if input, exists := ig.inputs[inputKey]; exists {
			input.conflicts.Remove(id)

			// If there is nothing attempting to consume the input, remove it
			// from memory
			if input.conflicts.Len() == 0 {
				delete(ig.inputs, inputKey)
				continue
			}

			// If I was previously preferred, I must find who should now be
			// preferred. This shouldn't normally happen, therefore it is okay
			// to be fairly slow here
			if input.preference.Equals(id) {
				newPreference := ids.ID{}
				newBias := -1
				newBiasTime := uint64(0)

				// Find the highest bias conflict
				for _, spend := range input.conflicts.List() {
					tx := ig.txs[spend.Key()]
					if tx.bias > newBias ||
						(tx.bias == newBias &&
							newBiasTime < tx.timestamp) {
						newPreference = spend
						newBias = tx.bias
						newBiasTime = tx.timestamp
					}
				}

				// Set the preferences to the highest bias
				input.preference = newPreference
				input.bias = newBias

				ig.inputs[inputKey] = input

				// We need to check if this node is now preferred
				preferenceNode, exist := ig.txs[newPreference.Key()]
				if exist {
					isPreferred := true
					inputIDs := preferenceNode.tx.InputIDs().List()
					for _, inputID := range inputIDs {
						inputKey := inputID.Key()
						input := ig.inputs[inputKey]

						if !newPreference.Equals(input.preference) {
							// If this preference isn't the preferred color, it
							// isn't preferred. Input might not exist, in which
							// case this still isn't the preferred color
							isPreferred = false
							break
						}
					}
					if isPreferred {
						// If I'm preferred in all my conflict sets, I'm
						// preferred
						ig.preferences.Add(newPreference)
					}
				}
			} else {
				// If i'm rejecting the non-preference, do nothing
				ig.inputs[inputKey] = input
			}
		}
	}
}

// Quiesce implements the ConflictGraph interface
func (ig *Input) Quiesce() bool {
	numVirtuous := ig.virtuousVoting.Len()
	ig.ctx.Log.Verbo("Conflict graph has %d voting virtuous transactions and %d transactions", numVirtuous, len(ig.txs))
	return numVirtuous == 0
}

// Finalized implements the ConflictGraph interface
func (ig *Input) Finalized() bool {
	numTxs := len(ig.txs)
	ig.ctx.Log.Verbo("Conflict graph has %d pending transactions", numTxs)
	return numTxs == 0
}

func (ig *Input) String() string {
	nodes := []tempNode{}
	for _, tx := range ig.txs {
		id := tx.tx.ID()

		confidence := ig.params.BetaRogue
		for _, inputID := range tx.tx.InputIDs().List() {
			input := ig.inputs[inputID.Key()]
			if input.lastVote != ig.currentVote {
				confidence = 0
				break
			}

			if input.confidence < confidence {
				confidence = input.confidence
			}
			if !id.Equals(input.color) {
				confidence = 0
				break
			}
		}

		nodes = append(nodes, tempNode{
			id:         id,
			bias:       tx.bias,
			confidence: confidence,
		})
	}
	sortTempNodes(nodes)

	sb := strings.Builder{}

	sb.WriteString("IG(")

	format := fmt.Sprintf(
		"\n    Choice[%s] = ID: %%50s Confidence: %s Bias: %%d",
		formatting.IntFormat(len(nodes)-1),
		formatting.IntFormat(ig.params.BetaRogue-1))

	for i, cn := range nodes {
		sb.WriteString(fmt.Sprintf(format, i, cn.id, cn.confidence, cn.bias))
	}

	if len(nodes) > 0 {
		sb.WriteString("\n")
	}
	sb.WriteString(")")

	return sb.String()
}

type inputAccepter struct {
	ig       *Input
	deps     ids.Set
	rejected bool
	tn       txNode
}

func (a *inputAccepter) Dependencies() ids.Set { return a.deps }

func (a *inputAccepter) Fulfill(id ids.ID) {
	a.deps.Remove(id)
	a.Update()
}

func (a *inputAccepter) Abandon(id ids.ID) { a.rejected = true }

func (a *inputAccepter) Update() {
	if a.rejected || a.deps.Len() != 0 || a.ig.errs.Errored() {
		return
	}

	id := a.tn.tx.ID()
	delete(a.ig.txs, id.Key())

	// Remove Tx from all of its conflicts
	inputIDs := a.tn.tx.InputIDs()
	a.ig.removeConflict(id, inputIDs.List()...)

	a.ig.virtuous.Remove(id)
	a.ig.preferences.Remove(id)

	// Reject the conflicts
	conflicts := ids.Set{}
	for inputKey, exists := range inputIDs {
		if exists {
			inputNode := a.ig.inputs[inputKey]
			conflicts.Union(inputNode.conflicts)
		}
	}
	if err := a.ig.reject(conflicts.List()...); err != nil {
		a.ig.errs.Add(err)
		return
	}

	// Mark it as accepted
	if err := a.tn.tx.Accept(); err != nil {
		a.ig.errs.Add(err)
		return
	}
	a.ig.ctx.DecisionDispatcher.Accept(a.ig.ctx.ChainID, id, a.tn.tx.Bytes())
	a.ig.metrics.Accepted(id)

	a.ig.pendingAccept.Fulfill(id)
	a.ig.pendingReject.Abandon(id)
}

// inputRejector implements Blockable
type inputRejector struct {
	ig       *Input
	deps     ids.Set
	rejected bool // true if the transaction represented by fn has been rejected
	tn       txNode
}

func (r *inputRejector) Dependencies() ids.Set { return r.deps }

func (r *inputRejector) Fulfill(id ids.ID) {
	if r.rejected || r.ig.errs.Errored() {
		return
	}
	r.rejected = true
	r.ig.errs.Add(r.ig.reject(r.tn.tx.ID()))
}

func (*inputRejector) Abandon(id ids.ID) {}

func (*inputRejector) Update() {}

type tempNode struct {
	id               ids.ID
	bias, confidence int
}

type sortTempNodeData []tempNode

func (tnd sortTempNodeData) Less(i, j int) bool {
	return bytes.Compare(tnd[i].id.Bytes(), tnd[j].id.Bytes()) == -1
}
func (tnd sortTempNodeData) Len() int      { return len(tnd) }
func (tnd sortTempNodeData) Swap(i, j int) { tnd[j], tnd[i] = tnd[i], tnd[j] }

func sortTempNodes(nodes []tempNode) { sort.Sort(sortTempNodeData(nodes)) }
