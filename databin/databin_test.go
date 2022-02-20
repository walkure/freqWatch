package databin

import "testing"

func Test_DataBin(t *testing.T) {

	drb1 := GetRingBuffer("hoge")
	if drb1 == nil {
		t.Errorf("cannot create ring buffer")
	}

	drb2 := LookupRingBuffer("fuga")
	if drb2 != nil {
		t.Errorf("unexpected buffer returned")
	}

	drb3 := LookupRingBuffer("hoge")
	if drb1 != drb3 {
		t.Errorf("map collapsed")
	}
}
