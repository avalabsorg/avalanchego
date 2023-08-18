// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package staking

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/utils/units"
)

// MaxRSAKeyBitLen is the maximum RSA key size in bits that we are willing to
// parse.
//
// https://github.com/golang/go/blob/go1.19.12/src/crypto/tls/handshake_client.go#L860-L862
const (
	MaxRSAKeyByteLen = units.KiB
	MaxRSAKeyBitLen  = 8 * MaxRSAKeyByteLen
)

var (
	ErrUnsupportedAlgorithm       = errors.New("staking: cannot verify signature: unsupported algorithm")
	ErrPublicKeyAlgoMismatch      = errors.New("staking: signature algorithm specified different public key type")
	ErrInvalidRSAPublicKey        = errors.New("staking: invalid RSA public key")
	ErrInvalidECDSAPublicKey      = errors.New("staking: invalid ECDSA public key")
	ErrECDSAVerificationFailure   = errors.New("staking: ECDSA verification failure")
	ErrED25519VerificationFailure = errors.New("staking: Ed25519 verification failure")
)

// CheckSignature verifies that the signature is a valid signature over signed
// from the certificate.
//
// Ref: https://github.com/golang/go/blob/go1.19.12/src/crypto/x509/x509.go#L793-L797
// Ref: https://github.com/golang/go/blob/go1.19.12/src/crypto/x509/x509.go#L816-L879
func CheckSignature(cert *Certificate, msg []byte, signature []byte) error {
	pubkeyAlgo, ok := signatureAlgorithmVerificationDetails[cert.SignatureAlgorithm]
	if !ok {
		return ErrUnsupportedAlgorithm
	}

	hasher := crypto.SHA256.New()
	_, err := hasher.Write(msg)
	if err != nil {
		return err
	}
	hashed := hasher.Sum(nil)

	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		if pubkeyAlgo != x509.RSA {
			return signaturePublicKeyAlgoMismatchError(pubkeyAlgo, pub)
		}
		if bitLen := pub.N.BitLen(); bitLen > MaxRSAKeyBitLen {
			return fmt.Errorf("%w: bitLen=%d > maxBitLen=%d", ErrInvalidRSAPublicKey, bitLen, MaxRSAKeyBitLen)
		}
		return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed, signature)
	case *ecdsa.PublicKey:
		if pubkeyAlgo != x509.ECDSA {
			return signaturePublicKeyAlgoMismatchError(pubkeyAlgo, pub)
		}
		if pub.Curve != elliptic.P256() {
			return ErrInvalidECDSAPublicKey
		}
		if !ecdsa.VerifyASN1(pub, hashed, signature) {
			return ErrECDSAVerificationFailure
		}
		return nil
	default:
		return ErrUnsupportedAlgorithm
	}
}

// Ref: https://github.com/golang/go/blob/go1.19.12/src/crypto/x509/x509.go#L812-L814
func signaturePublicKeyAlgoMismatchError(expectedPubKeyAlgo x509.PublicKeyAlgorithm, pubKey any) error {
	return fmt.Errorf("%w: expected an %s public key, but have public key of type %T", ErrPublicKeyAlgoMismatch, expectedPubKeyAlgo, pubKey)
}
