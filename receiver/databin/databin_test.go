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

	testDb.GetRingBuffer("poyo")

	foundPoyo := false
	foundHoge := false
	keys := testDb.ListRingBuffer()
	for _, k := range keys {
		if k == "hoge" {
			foundHoge = true
		}
		if k == "poyo" {
			foundPoyo = true
		}
	}
	if !foundPoyo || !foundHoge {
		t.Errorf("keys invalid. %+v", keys)
	}

}
