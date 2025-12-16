package guc

import (
	"hash/crc32"
	"math"
	"sync"
)

type (
	SegKey interface {
		~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 |
			~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
	}

	HashFunc[K SegKey] func(key K) uint32

	Call func()

	ReadAction func() bool

	WriteActon func()
)

const (
	defaultStrip = 8
)

type SegmentLock[K SegKey] struct {
	locks    []*sync.Mutex
	lockMask uint32
	hf       HashFunc[K]
}

func NewSegmentLock[K SegKey](strip int, hf HashFunc[K]) *SegmentLock[K] {
	strip = correctStrip(strip)

	locks := make([]*sync.Mutex, strip)
	for i := 0; i < strip; i++ {
		locks[i] = new(sync.Mutex)
	}

	if hf == nil {
		hf = defaultHf[K]
	}

	return &SegmentLock[K]{
		locks:    locks,
		lockMask: uint32(strip - 1),
		hf:       hf,
	}
}

func (sl *SegmentLock[K]) SafeCall(key K, call Call) {
	lock := sl.getLock(key)

	lock.Lock()
	defer lock.Unlock()

	call()
}

func (sl *SegmentLock[K]) getLock(key K) *sync.Mutex {
	idx := sl.hf(key) & sl.lockMask
	return sl.locks[idx]
}

type SegmentRwLock[K SegKey] struct {
	locks    []*sync.RWMutex
	lockMask uint32
	hf       HashFunc[K]
}

func NewSegmentRwLock[K SegKey](strip int, hf HashFunc[K]) *SegmentRwLock[K] {
	strip = correctStrip(strip)

	locks := make([]*sync.RWMutex, strip)
	for i := 0; i < strip; i++ {
		locks[i] = new(sync.RWMutex)
	}

	if hf == nil {
		hf = defaultHf[K]
	}

	return &SegmentRwLock[K]{
		locks:    locks,
		lockMask: uint32(strip - 1),
		hf:       hf,
	}
}

func (slr *SegmentRwLock[K]) ReadWriteWithDoubleCheck(key K, ra ReadAction, wa WriteActon) {
	lock := slr.getLock(key)

	lock.RLock()
	doWrite := ra()
	lock.RUnlock()

	if !doWrite {
		return
	}

	lock.Lock()
	defer lock.Unlock()

	// double check
	doWrite = ra()
	if !doWrite {
		return
	}

	wa()
}

type LockAction func(lock *sync.RWMutex) (any, error)

func (slr *SegmentRwLock[K]) WithLock(key K, la LockAction) (any, error) {
	lock := slr.getLock(key)

	return la(lock)
}

func (slr *SegmentRwLock[K]) PureRead(key K, call Call) {
	lock := slr.getLock(key)

	lock.RLock()
	defer lock.RUnlock()

	call()
}

func (slr *SegmentRwLock[K]) SafeCall(key K, call Call) {
	lock := slr.getLock(key)

	lock.Lock()
	defer lock.Unlock()

	call()
}

func (slr *SegmentRwLock[K]) getLock(key K) *sync.RWMutex {
	idx := slr.hf(key) & slr.lockMask
	return slr.locks[idx]
}

func correctStrip(strip int) int {
	if strip == 0 {
		return defaultStrip
	}

	return 1 << uint(math.Log2(float64(strip-1))+1)
}

func defaultHf[K SegKey](key K) uint32 {
	switch k := any(key).(type) {
	case string:
		return crc32.ChecksumIEEE([]byte(k))
	case int:
		return uint32(k) * 0x9e3779b9
	case int8:
		return uint32(k) * 0x9e3779b9
	case int16:
		return uint32(k) * 0x9e3779b9
	case int32:
		return uint32(k) * 0x9e3779b9
	case int64:
		return uint32(k) * 0x9e3779b9
	case uint:
		return uint32(k) * 0x9e3779b9
	case uint8:
		return uint32(k) * 0x9e3779b9
	case uint16:
		return uint32(k) * 0x9e3779b9
	case uint32:
		return k * 0x9e3779b9
	case uint64:
		return uint32(k) * 0x9e3779b9
	default:
		panic("unsupported key type")
	}
}
