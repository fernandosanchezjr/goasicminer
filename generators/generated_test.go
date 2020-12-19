package generators

import "testing"

func TestGenerated(t *testing.T) {
	var a = Generated{
		ExtraNonce2: 123,
		NTime:       456,
		Version0:    123,
		Version1:    133,
		Version2:    143,
		Version3:    153,
	}
	if a != a {
		t.Fail()
	}
	var testMap = make(map[Generated]int)
	testMap[a] = 1
	var b = a
	b.Version3 = 103
	if a == b {
		t.Fail()
	}
	testMap[b] = 2
	if value := testMap[a]; value != 1 {
		t.Fail()
	}
	if value := testMap[b]; value != 2 {
		t.Fail()
	}
	b.Version3 = 153
	if a != b {
		t.Fail()
	}
}
