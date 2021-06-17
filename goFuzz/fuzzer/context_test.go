package fuzzer

import "testing"

func TestFuzzContextEnqueueQueryEntry(t *testing.T) {
	c := NewFuzzContext()

	entry1 := &FuzzQueryEntry{}
	entry2 := &FuzzQueryEntry{}
	c.EnqueueQueryEntry(entry1)
	c.EnqueueQueryEntry(entry2)

	if c.fuzzingQueue.Len() != 2 {
		t.Fail()
	}

	if c.fuzzingQueue.Front().Value != entry1 {
		t.Fail()
	}

	if c.fuzzingQueue.Back().Value != entry2 {
		t.Fail()
	}
}

func TestFuzzContextDequeueQueryEntry(t *testing.T) {
	c := NewFuzzContext()

	entry1 := &FuzzQueryEntry{}
	entry2 := &FuzzQueryEntry{}
	c.EnqueueQueryEntry(entry1)
	c.EnqueueQueryEntry(entry2)
	c.DequeueQueryEntry()
	if c.fuzzingQueue.Len() != 1 {
		t.Fail()
	}

	if c.fuzzingQueue.Front().Value != entry2 {
		t.Fail()
	}

	if c.fuzzingQueue.Back().Value != entry2 {
		t.Fail()
	}

	re, _ := c.DequeueQueryEntry()
	if re != entry2 {
		t.Fail()
	}
}
