// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package info

import (
	"net/http"

	"github.com/gorilla/rpc/v2"

	"github.com/ava-labs/gecko/chains"
	"github.com/ava-labs/gecko/genesis"
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/network"
	"github.com/ava-labs/gecko/snow/engine/common"
	"github.com/ava-labs/gecko/utils/logging"
	"github.com/ava-labs/gecko/version"

	cjson "github.com/ava-labs/gecko/utils/json"
)

// Info is the API service for unprivileged info on a node
type Info struct {
	version      version.Version
	nodeID       ids.ShortID
	networkID    uint32
	log          logging.Logger
	networking   network.Network
	chainManager chains.Manager
}

// NewService returns a new admin API service
func NewService(log logging.Logger, version version.Version, nodeID ids.ShortID, networkID uint32, chainManager chains.Manager, peers network.Network) *common.HTTPHandler {
	newServer := rpc.NewServer()
	codec := cjson.NewCodec()
	newServer.RegisterCodec(codec, "application/json")
	newServer.RegisterCodec(codec, "application/json;charset=UTF-8")
	newServer.RegisterService(&Info{
		version:      version,
		nodeID:       nodeID,
		networkID:    networkID,
		log:          log,
		chainManager: chainManager,
		networking:   peers,
	}, "info")
	return &common.HTTPHandler{Handler: newServer}
}

// GetNodeVersionReply are the results from calling GetNodeVersion
type GetNodeVersionReply struct {
	Version string `json:"version"`
}

// GetNodeVersion returns the version this node is running
func (service *Info) GetNodeVersion(_ *http.Request, _ *struct{}, reply *GetNodeVersionReply) error {
	service.log.Info("Info: GetNodeVersion called")

	reply.Version = service.version.String()
	return nil
}

// GetNodeIDReply are the results from calling GetNodeID
type GetNodeIDReply struct {
	NodeID ids.ShortID `json:"nodeID"`
}

// GetNodeID returns the node ID of this node
func (service *Info) GetNodeID(_ *http.Request, _ *struct{}, reply *GetNodeIDReply) error {
	service.log.Info("Info: GetNodeID called")

	reply.NodeID = service.nodeID
	return nil
}

// GetNetworkIDReply are the results from calling GetNetworkID
type GetNetworkIDReply struct {
	NetworkID cjson.Uint32 `json:"networkID"`
}

// GetNetworkID returns the network ID this node is running on
func (service *Info) GetNetworkID(_ *http.Request, _ *struct{}, reply *GetNetworkIDReply) error {
	service.log.Info("Info: GetNetworkID called")

	reply.NetworkID = cjson.Uint32(service.networkID)
	return nil
}

// GetNetworkNameReply is the result from calling GetNetworkName
type GetNetworkNameReply struct {
	NetworkName string `json:"networkName"`
}

// GetNetworkName returns the network name this node is running on
func (service *Info) GetNetworkName(_ *http.Request, _ *struct{}, reply *GetNetworkNameReply) error {
	service.log.Info("Info: GetNetworkName called")

	reply.NetworkName = genesis.NetworkName(service.networkID)
	return nil
}

// GetBlockchainIDArgs are the arguments for calling GetBlockchainID
type GetBlockchainIDArgs struct {
	Alias string `json:"alias"`
}

// GetBlockchainIDReply are the results from calling GetBlockchainID
type GetBlockchainIDReply struct {
	BlockchainID string `json:"blockchainID"`
}

// GetBlockchainID returns the blockchain ID that resolves the alias that was supplied
func (service *Info) GetBlockchainID(_ *http.Request, args *GetBlockchainIDArgs, reply *GetBlockchainIDReply) error {
	service.log.Info("Info: GetBlockchainID called")

	bID, err := service.chainManager.Lookup(args.Alias)
	reply.BlockchainID = bID.String()
	return err
}

// PeersReply are the results from calling Peers
type PeersReply struct {
	Peers []network.PeerID `json:"peers"`
}

// Peers returns the list of current validators
func (service *Info) Peers(_ *http.Request, _ *struct{}, reply *PeersReply) error {
	service.log.Info("Info: Peers called")

	reply.Peers = service.networking.Peers()
	return nil
}
