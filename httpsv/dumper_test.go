package httpsv

import (
	"testing"

	"github.com/walkure/freq_recv/databin"
)

func Test_DumperSimple(t *testing.T) {
	dumper := NewDumperHandler()

	got := dumper.generateJSON("testPlace")
	if got != "" {
		t.Errorf("init:unxpected response(expects \"\"):%s\n", got)
	}

	// set test data
	dbr := databin.GetRingBuffer("testPlace")
	dbr.PushBack(&databin.FreqDatum{Epoch: 1, Freq: 0.1})
	dbr.PushBack(&databin.FreqDatum{Epoch: 2, Freq: 0.2})
	dbr.PushBack(&databin.FreqDatum{Epoch: 3, Freq: 0.3})

	got = dumper.generateJSON("testPlace")
	expected := `[{"t":3,"f":0.300000},{"t":2,"f":0.200000},{"t":1,"f":0.100000}]`
	if got != expected {
		t.Errorf("first-generate:unexpected response.expected:%s got:%s\n", expected, got)
	}

	dbr.PushBack(&databin.FreqDatum{Epoch: 4, Freq: 0.4})
	got = dumper.generateJSON("testPlace")
	if got != expected {
		t.Errorf("first-generate-cached:unexpected response.expected:%s got:%s\n", expected, got)
	}

	dumper.InvalidateJsonCache("testPlace")
	got = dumper.generateJSON("testPlace")
	expected = `[{"t":4,"f":0.400000},{"t":3,"f":0.300000},{"t":2,"f":0.200000},{"t":1,"f":0.100000}]`
	if got != expected {
		t.Errorf("second-generate:unexpected response.expected:%s got:%s\n", expected, got)
	}

}
