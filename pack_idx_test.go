package gitguts_test

import (
	"testing"

	"github.com/rubyist/gitguts"
)

func TestPackIndex(t *testing.T) {
	idx, err := gitguts.OpenPackIndex("fixtures/pack.idx")
	if err != nil {
		t.Fatal(err)
	}

	objects := []struct {
		oid    string
		offset int
	}{
		{oid: "aa763b87e9737787f9341fc4ced04dffc16c6490", offset: 12},
		{oid: "ba4fcdffc2882b2eaad6d56f2bc208e085a31f12", offset: 130},
		{oid: "065a7ba193a6fbc6c184eb5fcf5e0876c5569f5f", offset: 156},
	}

	for _, obj := range objects {
		oid, _ := gitguts.ToOID(obj.oid)

		offset, err := idx.OffsetOf(oid)
		if err != nil {
			t.Fatal(err)
		}

		if offset != obj.offset {
			t.Errorf("expected an offset of %d, got %d", obj.offset, offset)
		}
	}
}

func BenchmarkPackIndex(b *testing.B) {
	idx, err := gitguts.OpenPackIndex("fixtures/pack.idx")
	if err != nil {
		b.Fatal(err)
	}

	oid, _ := gitguts.ToOID("ba4fcdffc2882b2eaad6d56f2bc208e085a31f12")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.OffsetOf(oid)
	}
}
