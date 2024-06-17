package low

import "testing"

func TestTags(t *testing.T) {
	for tag := NoneTag; tag < MaxTag+5; tag++ {
		s := tag.String()
		o := TagFromString(s)
		if o != tag {
			t.Errorf("%3d mismatch %d with %q", tag, o, s)
		}
	}

	s := TagT(100).String()
	e := "(100)Tag"
	if s != e {
		t.Errorf("expect %q, got %q", e, s)
	}
	o := TagFromString(s)
	if o != 100 {
		t.Errorf("expect 100, got %d", o)
	}
	o = TagFromString("(123)Toto")
	if o != InvalidTag {
		t.Errorf("expect InvalidTag, got %d", o)
	}
	s = InvalidTag.String()
	if s != "InvalidTag" {
		t.Errorf("expect %q, got %q", "InvalidTag", s)
	}
	o = TagFromString("InvalidTag")
	if o != InvalidTag {
		t.Errorf("expect InvalidTag, got %d", o)
	}
}
