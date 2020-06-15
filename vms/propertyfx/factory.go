package propertyfx

import (
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/snow"
)

// ID that this Fx uses when labeled
var (
	ID = ids.NewID([32]byte{'p', 'r', 'o', 'p', 'e', 'r', 't', 'y', 'f', 'x'})
)

// Factory ...
type Factory struct{}

// New ...
func (f *Factory) New(*snow.Context) (interface{}, error) { return &Fx{}, nil }
