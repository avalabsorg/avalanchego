// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package message

import (
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/wrappers"
)

var (
	ErrParsingContainerID    = errors.New("could not parse container ID")
	ErrDuplicatedContainerID = errors.New("inbound message contains duplicated container ID")
)

// Field that may be packed into a message
type Field uint32

// Fields that may be packed. These values are not sent over the wire.
const (
	VersionStr          Field = iota // Used in handshake
	NetworkID                        // Used in handshake
	NodeID                           // Used in handshake
	MyTime                           // Used in handshake
	IP                               // Used in handshake
	Peers                            // Used in handshake
	ChainID                          // Used for dispatching
	RequestID                        // Used for all messages
	Deadline                         // Used for request messages
	ContainerID                      // Used for querying
	ContainerBytes                   // Used for gossiping
	ContainerIDs                     // Used for querying
	MultiContainerBytes              // Used in MultiPut
	SigBytes                         // Used in handshake / peer gossiping
	VersionTime                      // Used in handshake / peer gossiping
	SignedPeers                      // Used in peer gossiping
	TrackedSubnets                   // Used in handshake / peer gossiping
	AppRequestBytes                  // Used at application level
	AppResponseBytes                 // Used at application level
	AppGossipBytes                   // Used at application level
)

// Packer returns the packer function that can be used to pack this field.
func (f Field) Packer() func(*wrappers.Packer, interface{}) {
	switch f {
	case VersionStr:
		return wrappers.TryPackStr
	case NetworkID:
		return wrappers.TryPackInt
	case NodeID:
		return wrappers.TryPackInt
	case MyTime:
		return wrappers.TryPackLong
	case IP:
		return wrappers.TryPackIP
	case Peers:
		return wrappers.TryPackIPList
	case ChainID: // TODO: This will be shortened to use a modified varint spec
		return wrappers.TryPackHash
	case RequestID:
		return wrappers.TryPackInt
	case Deadline:
		return wrappers.TryPackLong
	case ContainerID:
		return wrappers.TryPackHash
	case ContainerBytes:
		return wrappers.TryPackBytes
	case ContainerIDs:
		return wrappers.TryPackHashes
	case MultiContainerBytes:
		return wrappers.TryPack2DBytes
	case AppRequestBytes, AppResponseBytes, AppGossipBytes:
		return wrappers.TryPackBytes
	case SigBytes:
		return wrappers.TryPackBytes
	case VersionTime:
		return wrappers.TryPackLong
	case SignedPeers:
		return wrappers.TryPackIPCertList
	case TrackedSubnets:
		return wrappers.TryPackHashes
	default:
		return nil
	}
}

// Unpacker returns the unpacker function that can be used to unpack this field.
func (f Field) Unpacker() func(*wrappers.Packer) interface{} {
	switch f {
	case VersionStr:
		return wrappers.TryUnpackStr
	case NetworkID:
		return wrappers.TryUnpackInt
	case NodeID:
		return wrappers.TryUnpackInt
	case MyTime:
		return wrappers.TryUnpackLong
	case IP:
		return wrappers.TryUnpackIP
	case Peers:
		return wrappers.TryUnpackIPList
	case ChainID: // TODO: This will be shortened to use a modified varint spec
		return wrappers.TryUnpackHash
	case RequestID:
		return wrappers.TryUnpackInt
	case Deadline:
		return wrappers.TryUnpackLong
	case ContainerID:
		return wrappers.TryUnpackHash
	case ContainerBytes:
		return wrappers.TryUnpackBytes
	case ContainerIDs:
		return wrappers.TryUnpackHashes
	case MultiContainerBytes:
		return wrappers.TryUnpack2DBytes
	case AppRequestBytes, AppResponseBytes, AppGossipBytes:
		return wrappers.TryUnpackBytes
	case SigBytes:
		return wrappers.TryUnpackBytes
	case VersionTime:
		return wrappers.TryUnpackLong
	case SignedPeers:
		return wrappers.TryUnpackIPCertList
	case TrackedSubnets:
		return wrappers.TryUnpackHashes
	default:
		return nil
	}
}

func (f Field) String() string {
	switch f {
	case VersionStr:
		return "VersionStr"
	case NetworkID:
		return "NetworkID"
	case NodeID:
		return "NodeID"
	case MyTime:
		return "MyTime"
	case IP:
		return "IP"
	case Peers:
		return "Peers"
	case ChainID:
		return "ChainID"
	case RequestID:
		return "RequestID"
	case Deadline:
		return "Deadline"
	case ContainerID:
		return "ContainerID"
	case ContainerBytes:
		return "Container Bytes"
	case ContainerIDs:
		return "Container IDs"
	case MultiContainerBytes:
		return "MultiContainerBytes"
	case AppRequestBytes:
		return "AppRequestBytes"
	case AppResponseBytes:
		return "AppResponseBytes"
	case AppGossipBytes:
		return "AppGossipBytes"
	case SigBytes:
		return "SigBytes"
	case VersionTime:
		return "VersionTime"
	case SignedPeers:
		return "SignedPeers"
	case TrackedSubnets:
		return "TrackedSubnets"
	default:
		return "Unknown Field"
	}
}

func encodeContainerIDs(containerIDs []ids.ID) [][]byte {
	containerIDBytes := make([][]byte, len(containerIDs))
	for i, containerID := range containerIDs {
		copy := containerID
		containerIDBytes[i] = copy[:]
	}
	return containerIDBytes
}

func DecodeContainerIDs(inMsg InboundMessage) ([]ids.ID, error) {
	containerIDsBytes := inMsg.Get(ContainerIDs).([][]byte)
	res := make([]ids.ID, len(containerIDsBytes))
	idSet := ids.NewSet(len(containerIDsBytes))
	for i, containerIDBytes := range containerIDsBytes {
		containerID, err := ids.ToID(containerIDBytes)
		if err != nil {
			return nil, ErrParsingContainerID
		}
		if idSet.Contains(containerID) {
			return nil, ErrDuplicatedContainerID
		}
		res[i] = containerID
		idSet.Add(containerID)
	}
	return res, nil
}
