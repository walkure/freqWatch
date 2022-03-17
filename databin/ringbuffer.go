package databin

import (
	"sort"
	"sync"
)

type DataRingBuffer struct {
	data []*FreqDatum
	head int
	tail int
	mu   sync.RWMutex
}

func NewDataRingBuffer(length int) *DataRingBuffer {
	if length <= 0 {
		panic("invalid length! OTL")
	}
	drb := &DataRingBuffer{}
	drb.Init(length)

	return drb
}

func (f *DataRingBuffer) Init(length int) {
	if length < 0 {
		if f.data == nil {
			panic("RingBuffer broken")
		}
		length = len(f.data)
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	f.data = make([]*FreqDatum, length)
	f.head = -1 * length
	f.tail = -1
}

func (f *DataRingBuffer) PushBack(d *FreqDatum) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.head++
	if f.head >= len(f.data) {
		f.head = 0
	}
	f.tail++
	if f.tail >= len(f.data) {
		f.tail = 0
	}

	f.data[f.tail] = d
}

func (f *DataRingBuffer) PeekAll(startFromOlder bool, peeker func(d *FreqDatum) bool) bool {
	if f.tail < 0 {
		return false
	}

	cbArgs := make([]*FreqDatum, 0, len(f.data))

	func() {
		f.mu.RLock()
		defer f.mu.RUnlock()

		head := f.head
		if head < 0 {
			head = 0
		}

		for i := f.tail; ; i-- {
			if i < 0 {
				i = len(f.data) - 1
			}
			cbArgs = append(cbArgs, f.data[i])

			if i == head {
				break
			}
		}
	}()

	if startFromOlder {
		sort.Slice(cbArgs, func(i, j int) bool {
			return cbArgs[i].Epoch < cbArgs[j].Epoch
		})
	}

	for _, cbArg := range cbArgs {
		if cbArg == nil {
			break
		}
		if peeker(cbArg) {
			return true
		}
	}
	return false
}

func (f *DataRingBuffer) Length() int {
	return len(f.data)
}
