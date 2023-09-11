package databin

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func Test_DataRingBufferSeq(t *testing.T) {
	drb := NewDataRingBuffer(4)

	drb.PeekFromNewer(func(d *FreqDatum) bool {
		t.Errorf("unexpected callback(nil buffer)")
		return false
	})

	tests := []string{
		"0",
		"10",
		"210",
		"3210",
		"4321",
		"5432",
		"6543",
		"7654",
		"8765",
		"9876",
		"a987",
		"ba98",
		"cba9",
		"dcba",
		"edcb",
		"fedc",
	}
	for idx, wants := range tests {
		t.Run(wants, func(t *testing.T) {
			sb := strings.Builder{}
			drb.PushBack(&FreqDatum{Epoch: int64(idx)})
			drb.PeekFromNewer(func(d *FreqDatum) bool {
				sb.WriteString(fmt.Sprintf("%x", d.Epoch))
				return false
			})
			if sb.String() != wants {
				t.Errorf("expected[%s] got[%s]\n", wants, sb.String())
			}

		})
	}

	drb.Init(-1)

	drb.PeekFromNewer(func(d *FreqDatum) bool {
		t.Errorf("unexpected callback(nil buffer)")
		return false
	})
}

func Test_DataRingBufferSeqRev(t *testing.T) {
	drb := NewDataRingBuffer(4)

	tests := []string{
		"0",
		"01",
		"012",
		"0123",
		"1234",
		"2345",
		"3456",
		"4567",
		"5678",
		"6789",
		"789a",
		"89ab",
		"9abc",
		"abcd",
		"bcde",
		"cdef",
	}
	for idx, wants := range tests {
		t.Run(wants, func(t *testing.T) {
			sb := strings.Builder{}
			drb.PushBack(&FreqDatum{Epoch: int64(idx)})
			drb.PeekFromOlder(func(d *FreqDatum) bool {
				sb.WriteString(fmt.Sprintf("%x", d.Epoch))
				return false
			})
			if sb.String() != wants {
				t.Errorf("expected[%s] got[%s]\n", wants, sb.String())
			}

		})
	}
}

func Test_DataRingBufferRandom(t *testing.T) {
	drb := NewDataRingBuffer(40)
	var wg sync.WaitGroup
	for i := 0; i < 50000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			drb.PushBack(&FreqDatum{Epoch: int64(i)})
		}(i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			drb.PeekFromNewer(func(d *FreqDatum) bool {
				return false
			})
		}()
	}
	wg.Wait()
}

func Test_DataRingBufferD(t *testing.T) {
	const buflen = 601
	drb := NewDataRingBuffer(buflen)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 80000; i++ {
			drb.PushBack(&FreqDatum{Epoch: int64(i)})
		}
	}()

	for n := 0; n < 500; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			prev := int64(-1)
			item := 0
			drb.PeekFromNewer(func(d *FreqDatum) bool {
				if prev > 0 && prev <= d.Epoch {
					t.Errorf("data collapsed prev:%d now:%d\n", prev, d.Epoch)
				}
				prev = d.Epoch
				item++
				return false
			})
			/*
				if item != buflen {
					t.Errorf("loop %d error!!!!\n", item)
				}
			*/
		}()
	}

	wg.Wait()
}
