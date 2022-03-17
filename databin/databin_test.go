package databin

import "testing"

func Test_DataBin(t *testing.T) {

	testDb := NewDataBin(10)

	drb1 := testDb.GetRingBuffer("hoge")
	if drb1 == nil {
		t.Errorf("cannot create ring buffer")
	}

	drb2 := testDb.LookupRingBuffer("fuga")
	if drb2 != nil {
		t.Errorf("unexpected buffer returned")
	}

	drb3 := testDb.LookupRingBuffer("hoge")
	if drb1 != drb3 {
		t.Errorf("map collapsed")
	}
}
