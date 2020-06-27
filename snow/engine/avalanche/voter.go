// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

import (
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/snow/consensus/snowstorm"
	"github.com/ava-labs/gecko/snow/engine/avalanche/vertex"
)

type voter struct {
	t         *Transitive
	vdr       ids.ShortID
	requestID uint32
	response  ids.Set
	deps      ids.Set
}

func (v *voter) Dependencies() ids.Set { return v.deps }

func (v *voter) Fulfill(id ids.ID) {
	v.deps.Remove(id)
	v.Update()
}

func (v *voter) Abandon(id ids.ID) { v.Fulfill(id) }

func (v *voter) Update() {
	if v.deps.Len() != 0 || v.t.errs.Errored() {
		return
	}

	results, finished := v.t.polls.Vote(v.requestID, v.vdr, v.response.List())
	if !finished {
		return
	}
	results = v.bubbleVotes(results)

	v.t.Config.Context.Log.Debug("Finishing poll with:\n%s", &results)
	if err := v.t.Consensus.RecordPoll(results); err != nil {
		v.t.errs.Add(err)
		return
	}

	txs := []snowstorm.Tx(nil)
	for _, orphanID := range v.t.Consensus.Orphans().List() {
		if tx, err := v.t.Config.VM.GetTx(orphanID); err == nil {
			txs = append(txs, tx)
		} else {
			v.t.Config.Context.Log.Warn("Failed to fetch %s during attempted re-issuance", orphanID)
		}
	}
	if len(txs) > 0 {
		v.t.Config.Context.Log.Debug("Re-issuing %d transactions", len(txs))
	}
	if err := v.t.batch(txs, true /*=force*/, false /*empty*/); err != nil {
		v.t.errs.Add(err)
		return
	}

	if v.t.Consensus.Quiesce() {
		v.t.Config.Context.Log.Debug("Avalanche engine can quiesce")
		return
	}

	v.t.Config.Context.Log.Debug("Avalanche engine can't quiesce")
	v.t.errs.Add(v.t.repoll())
}

func (v *voter) bubbleVotes(votes ids.UniqueBag) ids.UniqueBag {
	bubbledVotes := ids.UniqueBag{}
	vertexHeap := vertex.NewHeap()
	for _, vote := range votes.List() {
		vtx, err := v.t.Config.State.GetVertex(vote)
		if err != nil {
			continue
		}

		vertexHeap.Push(vtx)
	}

	for vertexHeap.Len() > 0 {
		vtx := vertexHeap.Pop()
		vtxID := vtx.ID()
		set := votes.GetSet(vtxID)
		status := vtx.Status()

		if !status.Fetched() {
			v.t.Config.Context.Log.Verbo("Dropping %d vote(s) for %s because the vertex is unknown", set.Len(), vtxID)
			bubbledVotes.RemoveSet(vtx.ID())
			continue
		}

		if status.Decided() {
			v.t.Config.Context.Log.Verbo("Dropping %d vote(s) for %s because the vertex is decided", set.Len(), vtxID)
			bubbledVotes.RemoveSet(vtx.ID())
			continue
		}

		if v.t.Consensus.VertexIssued(vtx) {
			v.t.Config.Context.Log.Verbo("Applying %d vote(s) for %s", set.Len(), vtx.ID())
			bubbledVotes.UnionSet(vtx.ID(), set)
		} else {
			v.t.Config.Context.Log.Verbo("Bubbling %d vote(s) for %s because the vertex isn't issued", set.Len(), vtx.ID())
			bubbledVotes.RemoveSet(vtx.ID()) // Remove votes for this vertex because it hasn't been issued
			for _, parentVtx := range vtx.Parents() {
				bubbledVotes.UnionSet(parentVtx.ID(), set)
				vertexHeap.Push(parentVtx)
			}
		}
	}

	return bubbledVotes
}
