package databin

import "sync"

var alldata map[string]*DataRingBuffer

func init() {
	alldata = make(map[string]*DataRingBuffer)
}

var dbMux sync.RWMutex

// 60(secs) * 10(mins)
const bufferLength = 60 * 10

func GetRingBuffer(place string) *DataRingBuffer {
	dbMux.Lock()
	defer dbMux.Unlock()
	drb, ok := alldata[place]
	if !ok {
		drb = NewDataRingBuffer(bufferLength)
		alldata[place] = drb
	}
	return drb
}

func LookupRingBuffer(place string) *DataRingBuffer {
	dbMux.RLock()
	defer dbMux.RUnlock()
	return alldata[place]
}
