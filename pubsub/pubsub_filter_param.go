package pubsub

import (
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/bloom"
)

type FilterParam struct {
	lock    sync.RWMutex
	address map[ids.ShortID]struct{}
	filter  bloom.Filter
}

func (f *FilterParam) Filter() bloom.Filter {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.filter
}

func (f *FilterParam) SetFilter(filter bloom.Filter) bloom.Filter {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.filter = filter
	return f.filter
}

func (f *FilterParam) CheckAddress(addr2check []byte) bool {
	f.lock.RLock()
	defer f.lock.RUnlock()
	if f.filter != nil && f.filter.Check(addr2check) {
		return true
	}
	addr2checkID := ByteToID(addr2check)
	_, ok := f.address[addr2checkID]
	return ok
}

func (f *FilterParam) HasFilter() bool {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.filter != nil || len(f.address) > 0
}

func (f *FilterParam) UpdateAddressMulti(unsubscribe bool, bl ...[]byte) {
	for _, b := range bl {
		f.UpdateAddress(unsubscribe, ByteToID(b))
	}
}

func (f *FilterParam) Len() int {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return len(f.address)
}

func (f *FilterParam) UpdateAddress(unsubscribe bool, address ids.ShortID) {
	switch unsubscribe {
	case true:
		f.lock.Lock()
		delete(f.address, address)
		f.lock.Unlock()
	default:
		f.lock.Lock()
		f.address[address] = struct{}{}
		f.lock.Unlock()
	}
}

func NewFilterParam() *FilterParam {
	return &FilterParam{address: make(map[ids.ShortID]struct{})}
}
