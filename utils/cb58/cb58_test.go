// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cb58

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test encoding bytes to a string and decoding back to bytes
func TestEncodeDecode(t *testing.T) {
	type test struct {
		bytes []byte
		str   string
	}

	id := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	tests := []test{
		{
			nil,
			"45PJLL",
		},
		{
			[]byte{},
			"45PJLL",
		},
		{
			[]byte{0},
			"1c7hwa",
		},
		{
			[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 255},
			"1NVSVezva3bAtJesnUj",
		},
		{
			id[:],
			"SkB92YpWm4Q2ijQHH34cqbKkCZWszsiQgHVjtNeFF2HdvDQU",
		},
	}

	for _, test := range tests {
		// Encode the bytes
		strResult, err := Encode(test.bytes)
		if err != nil {
			t.Fatal(err)
		}
		// Make sure the string repr. is what we expected
		assert.Equal(t, test.str, strResult)
		// Decode the string
		bytesResult, err := Decode(strResult)
		if err != nil {
			t.Fatal(err)
		}
		// Make sure we got the same bytes back
		if !bytes.Equal(test.bytes, bytesResult) {
			t.Fatal("bytes not symmetric")
		}
	}
}
