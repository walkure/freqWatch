package databin

import (
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
	drb.data = make([]*FreqDatum, length)
	drb.head = -1 * length
	drb.tail = -1

	return drb
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

func (f *DataRingBuffer) PeekAll(peeker func(d *FreqDatum) bool) bool {
	if f.tail < 0 {
		return false
	}

	cbArgs := make([]*FreqDatum, len(f.data))

	func() {
		f.mu.RLock()
		defer f.mu.RUnlock()
		index := 0

		head := f.head
		if head < 0 {
			head = 0
		}

		for i := f.tail; ; i-- {
			if i < 0 {
				i = len(f.data) - 1
			}
			cbArgs[index] = f.data[i]
			index++

			if i == head {
				break
			}
		}
	}()

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
