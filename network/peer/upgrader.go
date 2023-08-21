// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package peer

import (
	"crypto/tls"
	"errors"
	"net"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/staking"
)

var (
	errNoCert = errors.New("tls handshake finished with no peer certificate")

	_ Upgrader = (*tlsServerUpgrader)(nil)
	_ Upgrader = (*tlsClientUpgrader)(nil)
)

type Upgrader interface {
	// Must be thread safe
	Upgrade(net.Conn) (ids.NodeID, net.Conn, *staking.Certificate, error)
}

type tlsServerUpgrader struct {
	config *tls.Config
}

func NewTLSServerUpgrader(config *tls.Config) Upgrader {
	return tlsServerUpgrader{
		config: config,
	}
}

func (t tlsServerUpgrader) Upgrade(conn net.Conn) (ids.NodeID, net.Conn, *staking.Certificate, error) {
	return connToIDAndCert(tls.Server(conn, t.config))
}

type tlsClientUpgrader struct {
	config *tls.Config
}

func NewTLSClientUpgrader(config *tls.Config) Upgrader {
	return tlsClientUpgrader{
		config: config,
	}
}

func (t tlsClientUpgrader) Upgrade(conn net.Conn) (ids.NodeID, net.Conn, *staking.Certificate, error) {
	return connToIDAndCert(tls.Client(conn, t.config))
}

func connToIDAndCert(conn *tls.Conn) (ids.NodeID, net.Conn, *staking.Certificate, error) {
	if err := conn.Handshake(); err != nil {
		return ids.NodeID{}, nil, nil, err
	}

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return ids.NodeID{}, nil, nil, errNoCert
	}

	tlsCert := state.PeerCertificates[0]
	// Invariant: ParseCertificate is used rather than CertificateFromX509 to
	// ensure that signature verification can assume the certificate was
	// parseable according the staking package's parser.
	peerCert, err := staking.ParseCertificate(tlsCert.Raw)
	if err != nil {
		return ids.NodeID{}, nil, nil, err
	}

	// We validate the certificate here to attempt to make the validity of the
	// peer certificate as clear as possible. Specifically, a node running a
	// prior version using an invalid certificate should not be able to report
	// healthy.
	if err := staking.ValidateCertificate(peerCert); err != nil {
		return ids.NodeID{}, nil, nil, err
	}

	nodeID := ids.NodeIDFromCert(peerCert)
	return nodeID, conn, peerCert, nil
}
