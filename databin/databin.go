package databin

import "sync"

type DataBin struct {
	alldata map[string]*DataRingBuffer
	dbMux   sync.RWMutex
	size    int
}

func NewDataBin(length int) *DataBin {
	return &DataBin{
		alldata: make(map[string]*DataRingBuffer),
		size:    length,
	}
}

func (db *DataBin) GetRingBuffer(place string) *DataRingBuffer {
	db.dbMux.Lock()
	defer db.dbMux.Unlock()
	drb, ok := db.alldata[place]
	if !ok {
		drb = NewDataRingBuffer(db.size)
		db.alldata[place] = drb
	}
	return drb
}

func (db *DataBin) LookupRingBuffer(place string) *DataRingBuffer {
	db.dbMux.RLock()
	defer db.dbMux.RUnlock()
	return db.alldata[place]
}

func (db *DataBin) ListRingBuffer() []string {
	db.dbMux.RLock()
	defer db.dbMux.RUnlock()

	keys := make([]string, len(db.alldata))
	i := 0
	for k := range db.alldata {
		keys[i] = k
		i++
	}

	return keys
}
