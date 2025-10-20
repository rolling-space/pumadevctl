package internal

import "testing"

func TestFindNextAvailablePortBlock_Basic(t *testing.T) {
	entries := []Entry{
		{Domain: "a.test", Mapping: "127.0.0.1:36000"}, // reserves 36000-36009
		{Domain: "b.test", Mapping: "36010"},           // reserves 36010-36019
		{Domain: "c.test", IsSymlink: true},            // ignored
	}
	base, err := FindNextAvailablePortBlock(entries, 36000, 36050, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base != 36020 {
		t.Fatalf("expected 36020, got %d", base)
	}
}

func TestFindNextAvailablePortBlock_RangeTooSmall(t *testing.T) {
	entries := []Entry{}
	if _, err := FindNextAvailablePortBlock(entries, 36000, 36008, 10); err == nil {
		t.Fatalf("expected error for too-small range")
	}
}

func TestFindNextAvailablePortBlock_Overlaps(t *testing.T) {
	// Fill 36000-36029 using three misaligned entries
	entries := []Entry{
		{Domain: "a.test", Mapping: "36005"}, // reserves 36005-36014 covers first two blocks partially
		{Domain: "b.test", Mapping: "36020"}, // reserves 36020-36029
	}
	base, err := FindNextAvailablePortBlock(entries, 36000, 36049, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// First aligned candidate blocks: 36000,36010,36020,36030
	// 36000 overlaps a.test (36005-36014), 36010 also overlaps, 36020 overlaps b.test, so 36030 wins
	if base != 36030 {
		t.Fatalf("expected 36030, got %d", base)
	}
}
