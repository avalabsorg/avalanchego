// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"github.com/ava-labs/gecko/ids"
)

// Note that since an AVA network has exactly one Platform Chain,
// and the Platform Chain defines the genesis state of the network
// (who is staking, which chains exist, etc.), defining the genesis
// state of the Platform Chain is the same as defining the genesis
// state of the network.

// Config contains the genesis addresses used to construct a genesis
type Config struct {
	MintAddresses, FundedAddresses, StakerIDs                   []string
	ParsedMintAddresses, ParsedFundedAddresses, ParsedStakerIDs []ids.ShortID
	EVMBytes                                                    []byte
}

func (c *Config) init() error {
	c.ParsedMintAddresses = nil
	for _, addrStr := range c.MintAddresses {
		addr, err := ids.ShortFromString(addrStr)
		if err != nil {
			return err
		}
		c.ParsedMintAddresses = append(c.ParsedMintAddresses, addr)
	}
	c.ParsedFundedAddresses = nil
	for _, addrStr := range c.FundedAddresses {
		addr, err := ids.ShortFromString(addrStr)
		if err != nil {
			return err
		}
		c.ParsedFundedAddresses = append(c.ParsedFundedAddresses, addr)
	}
	c.ParsedStakerIDs = nil
	for _, addrStr := range c.StakerIDs {
		addr, err := ids.ShortFromString(addrStr)
		if err != nil {
			return err
		}
		c.ParsedStakerIDs = append(c.ParsedStakerIDs, addr)
	}
	return nil
}

// Hard coded genesis constants
var (
	DenaliConfig = Config{
		MintAddresses: []string{
			"95YUFjhDG892VePMzpwKF9JzewGKvGRi3",
		},
		FundedAddresses: []string{
			"9uKvvA7E35QCwLvAaohXTCfFejbf3Rv17",
			"JLrYNMYXANGj43BfWXBxMMAEenUBp1Sbn",
			"7TUTzwrU6nbZtWHjTHEpdneUvjKBxb3EM",
			"77mPUXBdQKwQpPoX6rckCZGLGGdkuG1G6",
			"4gGWdFZ4Gax1B466YKXyKRRpWLb42Afdt",
			"CKTkzAPsRxCreyiDTnjGxLmjMarxF28fi",
			"4ABm9gFHVtsNdcKSd1xsacFkGneSgzpaa",
			"DpL8PTsrjtLzv5J8LL3D2A6YcnCTqrNH9",
			"ZdhZv6oZrmXLyFDy6ovXAu6VxmbTsT2h",
			"6cesTteH62Y5mLoDBUASaBvCXuL2AthL",
		},
		StakerIDs: []string{
			"NX4zVkuiRJZYe6Nzzav7GXN3TakUet3Co",
			"CMsa8cMw4eib1Hb8GG4xiUKAq5eE1BwUX",
			"DsMP6jLhi1MkDVc3qx9xx9AAZWx8e87Jd",
			"N86eodVZja3GEyZJTo3DFUPGpxEEvjGHs",
			"EkKeGSLUbHrrtuayBtbwgWDRUiAziC3ao",
		},
		EVMBytes: []byte{
			0x7b, 0x22, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
			0x22, 0x3a, 0x7b, 0x22, 0x63, 0x68, 0x61, 0x69,
			0x6e, 0x49, 0x64, 0x22, 0x3a, 0x34, 0x33, 0x31,
			0x31, 0x30, 0x2c, 0x22, 0x68, 0x6f, 0x6d, 0x65,
			0x73, 0x74, 0x65, 0x61, 0x64, 0x42, 0x6c, 0x6f,
			0x63, 0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x64,
			0x61, 0x6f, 0x46, 0x6f, 0x72, 0x6b, 0x42, 0x6c,
			0x6f, 0x63, 0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22,
			0x64, 0x61, 0x6f, 0x46, 0x6f, 0x72, 0x6b, 0x53,
			0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x22, 0x3a,
			0x74, 0x72, 0x75, 0x65, 0x2c, 0x22, 0x65, 0x69,
			0x70, 0x31, 0x35, 0x30, 0x42, 0x6c, 0x6f, 0x63,
			0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x65, 0x69,
			0x70, 0x31, 0x35, 0x30, 0x48, 0x61, 0x73, 0x68,
			0x22, 0x3a, 0x22, 0x30, 0x78, 0x32, 0x30, 0x38,
			0x36, 0x37, 0x39, 0x39, 0x61, 0x65, 0x65, 0x62,
			0x65, 0x61, 0x65, 0x31, 0x33, 0x35, 0x63, 0x32,
			0x34, 0x36, 0x63, 0x36, 0x35, 0x30, 0x32, 0x31,
			0x63, 0x38, 0x32, 0x62, 0x34, 0x65, 0x31, 0x35,
			0x61, 0x32, 0x63, 0x34, 0x35, 0x31, 0x33, 0x34,
			0x30, 0x39, 0x39, 0x33, 0x61, 0x61, 0x63, 0x66,
			0x64, 0x32, 0x37, 0x35, 0x31, 0x38, 0x38, 0x36,
			0x35, 0x31, 0x34, 0x66, 0x30, 0x22, 0x2c, 0x22,
			0x65, 0x69, 0x70, 0x31, 0x35, 0x35, 0x42, 0x6c,
			0x6f, 0x63, 0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22,
			0x65, 0x69, 0x70, 0x31, 0x35, 0x38, 0x42, 0x6c,
			0x6f, 0x63, 0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22,
			0x62, 0x79, 0x7a, 0x61, 0x6e, 0x74, 0x69, 0x75,
			0x6d, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x3a,
			0x30, 0x2c, 0x22, 0x63, 0x6f, 0x6e, 0x73, 0x74,
			0x61, 0x6e, 0x74, 0x69, 0x6e, 0x6f, 0x70, 0x6c,
			0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x3a,
			0x30, 0x2c, 0x22, 0x70, 0x65, 0x74, 0x65, 0x72,
			0x73, 0x62, 0x75, 0x72, 0x67, 0x42, 0x6c, 0x6f,
			0x63, 0x6b, 0x22, 0x3a, 0x30, 0x7d, 0x2c, 0x22,
			0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x22, 0x3a, 0x22,
			0x30, 0x78, 0x30, 0x22, 0x2c, 0x22, 0x74, 0x69,
			0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22,
			0x3a, 0x22, 0x30, 0x78, 0x30, 0x22, 0x2c, 0x22,
			0x65, 0x78, 0x74, 0x72, 0x61, 0x44, 0x61, 0x74,
			0x61, 0x22, 0x3a, 0x22, 0x30, 0x78, 0x30, 0x30,
			0x22, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x4c, 0x69,
			0x6d, 0x69, 0x74, 0x22, 0x3a, 0x22, 0x30, 0x78,
			0x35, 0x66, 0x35, 0x65, 0x31, 0x30, 0x30, 0x22,
			0x2c, 0x22, 0x64, 0x69, 0x66, 0x66, 0x69, 0x63,
			0x75, 0x6c, 0x74, 0x79, 0x22, 0x3a, 0x22, 0x30,
			0x78, 0x30, 0x22, 0x2c, 0x22, 0x6d, 0x69, 0x78,
			0x48, 0x61, 0x73, 0x68, 0x22, 0x3a, 0x22, 0x30,
			0x78, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x22, 0x2c, 0x22, 0x63, 0x6f, 0x69, 0x6e,
			0x62, 0x61, 0x73, 0x65, 0x22, 0x3a, 0x22, 0x30,
			0x78, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x22, 0x2c, 0x22, 0x61, 0x6c, 0x6c, 0x6f,
			0x63, 0x22, 0x3a, 0x7b, 0x22, 0x35, 0x37, 0x32,
			0x66, 0x34, 0x64, 0x38, 0x30, 0x66, 0x31, 0x30,
			0x66, 0x36, 0x36, 0x33, 0x62, 0x35, 0x30, 0x34,
			0x39, 0x66, 0x37, 0x38, 0x39, 0x35, 0x34, 0x36,
			0x66, 0x32, 0x35, 0x66, 0x37, 0x30, 0x62, 0x62,
			0x36, 0x32, 0x61, 0x37, 0x66, 0x22, 0x3a, 0x7b,
			0x22, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65,
			0x22, 0x3a, 0x22, 0x30, 0x78, 0x33, 0x33, 0x62,
			0x32, 0x65, 0x33, 0x63, 0x39, 0x66, 0x64, 0x30,
			0x38, 0x30, 0x34, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x22, 0x7d, 0x7d, 0x2c,
			0x22, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x22,
			0x3a, 0x22, 0x30, 0x78, 0x30, 0x22, 0x2c, 0x22,
			0x67, 0x61, 0x73, 0x55, 0x73, 0x65, 0x64, 0x22,
			0x3a, 0x22, 0x30, 0x78, 0x30, 0x22, 0x2c, 0x22,
			0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x48, 0x61,
			0x73, 0x68, 0x22, 0x3a, 0x22, 0x30, 0x78, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x22,
			0x7d,
		},
	}
	CascadeConfig = Config{
		MintAddresses: []string{
			"95YUFjhDG892VePMzpwKF9JzewGKvGRi3",
		},
		FundedAddresses: []string{
			"9uKvvA7E35QCwLvAaohXTCfFejbf3Rv17",
			"JLrYNMYXANGj43BfWXBxMMAEenUBp1Sbn",
			"7TUTzwrU6nbZtWHjTHEpdneUvjKBxb3EM",
			"77mPUXBdQKwQpPoX6rckCZGLGGdkuG1G6",
			"4gGWdFZ4Gax1B466YKXyKRRpWLb42Afdt",
			"CKTkzAPsRxCreyiDTnjGxLmjMarxF28fi",
			"4ABm9gFHVtsNdcKSd1xsacFkGneSgzpaa",
			"DpL8PTsrjtLzv5J8LL3D2A6YcnCTqrNH9",
			"ZdhZv6oZrmXLyFDy6ovXAu6VxmbTsT2h",
			"6cesTteH62Y5mLoDBUASaBvCXuL2AthL",
		},
		StakerIDs: []string{
			"NX4zVkuiRJZYe6Nzzav7GXN3TakUet3Co",
			"CMsa8cMw4eib1Hb8GG4xiUKAq5eE1BwUX",
			"DsMP6jLhi1MkDVc3qx9xx9AAZWx8e87Jd",
			"N86eodVZja3GEyZJTo3DFUPGpxEEvjGHs",
			"EkKeGSLUbHrrtuayBtbwgWDRUiAziC3ao",
		},
		EVMBytes: []byte{
			0x7b, 0x22, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
			0x22, 0x3a, 0x7b, 0x22, 0x63, 0x68, 0x61, 0x69,
			0x6e, 0x49, 0x64, 0x22, 0x3a, 0x34, 0x33, 0x31,
			0x31, 0x30, 0x2c, 0x22, 0x68, 0x6f, 0x6d, 0x65,
			0x73, 0x74, 0x65, 0x61, 0x64, 0x42, 0x6c, 0x6f,
			0x63, 0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x64,
			0x61, 0x6f, 0x46, 0x6f, 0x72, 0x6b, 0x42, 0x6c,
			0x6f, 0x63, 0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22,
			0x64, 0x61, 0x6f, 0x46, 0x6f, 0x72, 0x6b, 0x53,
			0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x22, 0x3a,
			0x74, 0x72, 0x75, 0x65, 0x2c, 0x22, 0x65, 0x69,
			0x70, 0x31, 0x35, 0x30, 0x42, 0x6c, 0x6f, 0x63,
			0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x65, 0x69,
			0x70, 0x31, 0x35, 0x30, 0x48, 0x61, 0x73, 0x68,
			0x22, 0x3a, 0x22, 0x30, 0x78, 0x32, 0x30, 0x38,
			0x36, 0x37, 0x39, 0x39, 0x61, 0x65, 0x65, 0x62,
			0x65, 0x61, 0x65, 0x31, 0x33, 0x35, 0x63, 0x32,
			0x34, 0x36, 0x63, 0x36, 0x35, 0x30, 0x32, 0x31,
			0x63, 0x38, 0x32, 0x62, 0x34, 0x65, 0x31, 0x35,
			0x61, 0x32, 0x63, 0x34, 0x35, 0x31, 0x33, 0x34,
			0x30, 0x39, 0x39, 0x33, 0x61, 0x61, 0x63, 0x66,
			0x64, 0x32, 0x37, 0x35, 0x31, 0x38, 0x38, 0x36,
			0x35, 0x31, 0x34, 0x66, 0x30, 0x22, 0x2c, 0x22,
			0x65, 0x69, 0x70, 0x31, 0x35, 0x35, 0x42, 0x6c,
			0x6f, 0x63, 0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22,
			0x65, 0x69, 0x70, 0x31, 0x35, 0x38, 0x42, 0x6c,
			0x6f, 0x63, 0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22,
			0x62, 0x79, 0x7a, 0x61, 0x6e, 0x74, 0x69, 0x75,
			0x6d, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x3a,
			0x30, 0x2c, 0x22, 0x63, 0x6f, 0x6e, 0x73, 0x74,
			0x61, 0x6e, 0x74, 0x69, 0x6e, 0x6f, 0x70, 0x6c,
			0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x3a,
			0x30, 0x2c, 0x22, 0x70, 0x65, 0x74, 0x65, 0x72,
			0x73, 0x62, 0x75, 0x72, 0x67, 0x42, 0x6c, 0x6f,
			0x63, 0x6b, 0x22, 0x3a, 0x30, 0x7d, 0x2c, 0x22,
			0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x22, 0x3a, 0x22,
			0x30, 0x78, 0x30, 0x22, 0x2c, 0x22, 0x74, 0x69,
			0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22,
			0x3a, 0x22, 0x30, 0x78, 0x30, 0x22, 0x2c, 0x22,
			0x65, 0x78, 0x74, 0x72, 0x61, 0x44, 0x61, 0x74,
			0x61, 0x22, 0x3a, 0x22, 0x30, 0x78, 0x30, 0x30,
			0x22, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x4c, 0x69,
			0x6d, 0x69, 0x74, 0x22, 0x3a, 0x22, 0x30, 0x78,
			0x35, 0x66, 0x35, 0x65, 0x31, 0x30, 0x30, 0x22,
			0x2c, 0x22, 0x64, 0x69, 0x66, 0x66, 0x69, 0x63,
			0x75, 0x6c, 0x74, 0x79, 0x22, 0x3a, 0x22, 0x30,
			0x78, 0x30, 0x22, 0x2c, 0x22, 0x6d, 0x69, 0x78,
			0x48, 0x61, 0x73, 0x68, 0x22, 0x3a, 0x22, 0x30,
			0x78, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x22, 0x2c, 0x22, 0x63, 0x6f, 0x69, 0x6e,
			0x62, 0x61, 0x73, 0x65, 0x22, 0x3a, 0x22, 0x30,
			0x78, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x22, 0x2c, 0x22, 0x61, 0x6c, 0x6c, 0x6f,
			0x63, 0x22, 0x3a, 0x7b, 0x22, 0x35, 0x37, 0x32,
			0x66, 0x34, 0x64, 0x38, 0x30, 0x66, 0x31, 0x30,
			0x66, 0x36, 0x36, 0x33, 0x62, 0x35, 0x30, 0x34,
			0x39, 0x66, 0x37, 0x38, 0x39, 0x35, 0x34, 0x36,
			0x66, 0x32, 0x35, 0x66, 0x37, 0x30, 0x62, 0x62,
			0x36, 0x32, 0x61, 0x37, 0x66, 0x22, 0x3a, 0x7b,
			0x22, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65,
			0x22, 0x3a, 0x22, 0x30, 0x78, 0x33, 0x33, 0x62,
			0x32, 0x65, 0x33, 0x63, 0x39, 0x66, 0x64, 0x30,
			0x38, 0x30, 0x34, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x22, 0x7d, 0x7d, 0x2c,
			0x22, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x22,
			0x3a, 0x22, 0x30, 0x78, 0x30, 0x22, 0x2c, 0x22,
			0x67, 0x61, 0x73, 0x55, 0x73, 0x65, 0x64, 0x22,
			0x3a, 0x22, 0x30, 0x78, 0x30, 0x22, 0x2c, 0x22,
			0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x48, 0x61,
			0x73, 0x68, 0x22, 0x3a, 0x22, 0x30, 0x78, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x22,
			0x7d,
		},
	}
	DefaultConfig = Config{
		MintAddresses: []string{},
		FundedAddresses: []string{
			// Private key: ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN
			"6Y3kysjF9jnHnYkdS9yGAuoHyae2eNmeV",
		},
		StakerIDs: []string{
			"7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg",
			"MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ",
			"NFBbbJ4qCmNaCzeW7sxErhvWqvEQMnYcN",
			"GWPcbFJZFfZreETSoWjPimr846mXEKCtu",
			"P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5",
		},
		EVMBytes: []byte{
			0x7b, 0x22, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
			0x22, 0x3a, 0x7b, 0x22, 0x63, 0x68, 0x61, 0x69,
			0x6e, 0x49, 0x64, 0x22, 0x3a, 0x34, 0x33, 0x31,
			0x31, 0x30, 0x2c, 0x22, 0x68, 0x6f, 0x6d, 0x65,
			0x73, 0x74, 0x65, 0x61, 0x64, 0x42, 0x6c, 0x6f,
			0x63, 0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x64,
			0x61, 0x6f, 0x46, 0x6f, 0x72, 0x6b, 0x42, 0x6c,
			0x6f, 0x63, 0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22,
			0x64, 0x61, 0x6f, 0x46, 0x6f, 0x72, 0x6b, 0x53,
			0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x22, 0x3a,
			0x74, 0x72, 0x75, 0x65, 0x2c, 0x22, 0x65, 0x69,
			0x70, 0x31, 0x35, 0x30, 0x42, 0x6c, 0x6f, 0x63,
			0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x65, 0x69,
			0x70, 0x31, 0x35, 0x30, 0x48, 0x61, 0x73, 0x68,
			0x22, 0x3a, 0x22, 0x30, 0x78, 0x32, 0x30, 0x38,
			0x36, 0x37, 0x39, 0x39, 0x61, 0x65, 0x65, 0x62,
			0x65, 0x61, 0x65, 0x31, 0x33, 0x35, 0x63, 0x32,
			0x34, 0x36, 0x63, 0x36, 0x35, 0x30, 0x32, 0x31,
			0x63, 0x38, 0x32, 0x62, 0x34, 0x65, 0x31, 0x35,
			0x61, 0x32, 0x63, 0x34, 0x35, 0x31, 0x33, 0x34,
			0x30, 0x39, 0x39, 0x33, 0x61, 0x61, 0x63, 0x66,
			0x64, 0x32, 0x37, 0x35, 0x31, 0x38, 0x38, 0x36,
			0x35, 0x31, 0x34, 0x66, 0x30, 0x22, 0x2c, 0x22,
			0x65, 0x69, 0x70, 0x31, 0x35, 0x35, 0x42, 0x6c,
			0x6f, 0x63, 0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22,
			0x65, 0x69, 0x70, 0x31, 0x35, 0x38, 0x42, 0x6c,
			0x6f, 0x63, 0x6b, 0x22, 0x3a, 0x30, 0x2c, 0x22,
			0x62, 0x79, 0x7a, 0x61, 0x6e, 0x74, 0x69, 0x75,
			0x6d, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x3a,
			0x30, 0x2c, 0x22, 0x63, 0x6f, 0x6e, 0x73, 0x74,
			0x61, 0x6e, 0x74, 0x69, 0x6e, 0x6f, 0x70, 0x6c,
			0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x3a,
			0x30, 0x2c, 0x22, 0x70, 0x65, 0x74, 0x65, 0x72,
			0x73, 0x62, 0x75, 0x72, 0x67, 0x42, 0x6c, 0x6f,
			0x63, 0x6b, 0x22, 0x3a, 0x30, 0x7d, 0x2c, 0x22,
			0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x22, 0x3a, 0x22,
			0x30, 0x78, 0x30, 0x22, 0x2c, 0x22, 0x74, 0x69,
			0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22,
			0x3a, 0x22, 0x30, 0x78, 0x30, 0x22, 0x2c, 0x22,
			0x65, 0x78, 0x74, 0x72, 0x61, 0x44, 0x61, 0x74,
			0x61, 0x22, 0x3a, 0x22, 0x30, 0x78, 0x30, 0x30,
			0x22, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x4c, 0x69,
			0x6d, 0x69, 0x74, 0x22, 0x3a, 0x22, 0x30, 0x78,
			0x35, 0x66, 0x35, 0x65, 0x31, 0x30, 0x30, 0x22,
			0x2c, 0x22, 0x64, 0x69, 0x66, 0x66, 0x69, 0x63,
			0x75, 0x6c, 0x74, 0x79, 0x22, 0x3a, 0x22, 0x30,
			0x78, 0x30, 0x22, 0x2c, 0x22, 0x6d, 0x69, 0x78,
			0x48, 0x61, 0x73, 0x68, 0x22, 0x3a, 0x22, 0x30,
			0x78, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x22, 0x2c, 0x22, 0x63, 0x6f, 0x69, 0x6e,
			0x62, 0x61, 0x73, 0x65, 0x22, 0x3a, 0x22, 0x30,
			0x78, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x22, 0x2c, 0x22, 0x61, 0x6c, 0x6c, 0x6f,
			0x63, 0x22, 0x3a, 0x7b, 0x22, 0x37, 0x35, 0x31,
			0x61, 0x30, 0x62, 0x39, 0x36, 0x65, 0x31, 0x30,
			0x34, 0x32, 0x62, 0x65, 0x65, 0x37, 0x38, 0x39,
			0x34, 0x35, 0x32, 0x65, 0x63, 0x62, 0x32, 0x30,
			0x32, 0x35, 0x33, 0x66, 0x62, 0x61, 0x34, 0x30,
			0x64, 0x62, 0x65, 0x38, 0x35, 0x22, 0x3a, 0x7b,
			0x22, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65,
			0x22, 0x3a, 0x22, 0x30, 0x78, 0x33, 0x33, 0x62,
			0x32, 0x65, 0x33, 0x63, 0x39, 0x66, 0x64, 0x30,
			0x38, 0x30, 0x34, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x22, 0x7d, 0x7d, 0x2c,
			0x22, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x22,
			0x3a, 0x22, 0x30, 0x78, 0x30, 0x22, 0x2c, 0x22,
			0x67, 0x61, 0x73, 0x55, 0x73, 0x65, 0x64, 0x22,
			0x3a, 0x22, 0x30, 0x78, 0x30, 0x22, 0x2c, 0x22,
			0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x48, 0x61,
			0x73, 0x68, 0x22, 0x3a, 0x22, 0x30, 0x78, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x22,
			0x7d,
		},
	}
)

// GetConfig ...
func GetConfig(networkID uint32) *Config {
	switch networkID {
	case DenaliID:
		return &DenaliConfig
	case CascadeID:
		return &CascadeConfig
	default:
		return &DefaultConfig
	}
}
