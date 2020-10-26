package invitecode

import "testing"

func TestEncode(t *testing.T) {
	for i := 10000000000; i < 10000000100; i++ {
		t.Log(Encode(uint64(i)))
	}
}

func TestDecode(t *testing.T) {
	for i := 1000000000; i < 1000000100; i++ {
		result := Encode(uint64(i))
		t.Log(i, result)
		num := Decode(result)
		if num != uint64(i) {
			t.Fatal("num != i")
		}
	}
}
