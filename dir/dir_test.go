package dir

import (
	"bytes"
	"reflect"
	"testing"
)

func m(ids ...uint64) DIR {
	return MustMake(ids...)
}

func doesPanic(f func()) (res bool) {
	defer func() {
		if r := recover(); r != nil {
			res = true
		}
	}()
	f()
	return
}

func TestDIR(t *testing.T) {
	tests := []struct {
		d                                                []uint64
		n, isNil, isInfo, isNode, isAbsolute, isRelative bool
		e                                                string
	}{
		// 0
		{d: nil, n: true, isNil: true, isAbsolute: true, isInfo: true},
		{d: []uint64{}, n: true, isNil: true, isAbsolute: true, isInfo: true},
		{d: []uint64{1}, isAbsolute: true, isInfo: true},
		{d: []uint64{1, 2}, isAbsolute: true, isInfo: true},
		{d: []uint64{1, 2, 0, 1}, e: "invalid dir: identifier 2 is 0"},
		// 5
		{d: []uint64{0}, isAbsolute: true, isNode: true},
		{d: []uint64{1, 0}, isAbsolute: true, isNode: true},
		{d: []uint64{0, 0}, isRelative: true, isNode: true},
		{d: []uint64{0, 1}, isRelative: true, isInfo: true},
		{d: []uint64{1, 2, 3, 4, 5, 6, 7, 8}, e: "invalid dir: too many identifiers"},
		// 10
		{d: []uint64{0, 1, 0}, isRelative: true, isNode: true},
	}
	for i, test := range tests {
		d, err := Make(test.d...)
		var errStr string
		if err != nil {
			errStr = err.Error()
		}
		if test.e != errStr {
			t.Errorf("%d expect error %q, got %q", i, test.e, errStr)
			continue
		}
		if test.e != "" {
			continue
		}
		if d.IsNil() != test.isNil {
			t.Errorf("%d expect isNil %v, got %v", i, test.isNil, d.IsNil())
		}
		if d.IsInfo() != test.isInfo {
			t.Errorf("%d expect isInfo %v, got %v", i, test.isInfo, d.IsInfo())
		}
		if d.IsNode() != test.isNode {
			t.Errorf("%d expect isNode %v, got %v", i, test.isNode, d.IsNode())
		}
		if d.IsAbsolute() != test.isAbsolute {
			t.Errorf("%d expect isAbsolute %v, got %v", i, test.isAbsolute, d.IsAbsolute())
		}
		if d.IsRelative() != test.isRelative {
			t.Errorf("%d expect isRelative %v, got %v", i, test.isRelative, d.IsRelative())
		}
	}

	d := m(1, 2)
	if exp := "dir:1.2"; d.String() != exp {
		t.Errorf("expect %q, got %q", exp, d.String())
	}

	d = DIR{}
	if d.InfoID() != 0 {
		t.Errorf("expect 0, got %d", d.InfoID())
	}

	d = m(1, 234)
	if d.InfoID() != 234 {
		t.Errorf("expect 234, got %d", d.InfoID())
	}

	if d.Len() != 2 {
		t.Errorf("expect len 2, got %d", d.Len())
	}

	d2 := m(d.IDs(nil)...)
	if !d.Equal(d2) {
		t.Errorf("expect equal")
	}
	d2 = m(d.IDs([]uint64{1})...)
	if !d.Equal(d2) {
		t.Errorf("expect equal")
	}

	d2.d[0] = 2
	if d.Equal(d2) {
		t.Errorf("expect not equal")
	}
	d2.d = append(d2.d, 123)
	if d.Equal(d2) {
		t.Errorf("expect not equal")
	}

	d = d2.Copy(d)
	if !d.Equal(d2) {
		t.Errorf("expect equal")
	}

	if d.ID(2) != 123 {
		t.Errorf("expect 123, got %d", d.ID(2))
	}
	if d.ID(3) != 0 {
		t.Errorf("expect 0, got %d", d.ID(2))
	}

	if !doesPanic(func() {
		MustMake(1, 2, 0, 3)
	}) {
		t.Error("expect panics")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		a, b DIR
		r    bool
	}{
		// 0
		{a: m(), b: m(), r: false},
		{a: m(0), b: m(1), r: true},
		{a: m(1, 0), b: m(1, 2), r: true},
		{a: m(1, 0), b: m(1, 0), r: false},
		{a: m(1, 0), b: m(1, 2, 0), r: true},
		// 5
		{a: m(2, 0), b: m(1, 2, 0), r: false},
	}
	for i, test := range tests {
		r := test.a.Contains(test.b)
		if r != test.r {
			t.Errorf("%d expect %v, got %v for %q contains %q", i, test.r, r, test.a, test.b)
		}
	}
}

func TestAppendBinary(t *testing.T) {
	var tests = []struct {
		i DIR
		o []byte
	}{
		// 0
		{i: DIR{}, o: nil},
		{i: m(1), o: []byte{1}},
		{i: m(0), o: []byte{0}},
		{i: m(0, 0), o: []byte{0, 0}},
		{i: m(0, 1, 0), o: []byte{0, 1, 0}},
		// 5
		{i: m(0x00), o: []byte{0}},
		{i: m(0x7F), o: []byte{0x7F}},
		{i: m(0x80), o: []byte{0x80, 0x01}},
		{i: m(0x3FFF), o: []byte{0xFF, 0x7F}},
		{i: m(0x4000), o: []byte{0x80, 0x80, 0x01}},
		// 10
		{i: m(0x1FFFFF), o: []byte{0xff, 0xff, 0x7f}},
		{i: m(0x200000), o: []byte{0x80, 0x80, 0x80, 0x01}},
		{i: m(0x0FFFFFFF), o: []byte{0xFF, 0xFF, 0xFF, 0x7F}},
		{i: m(0x10000000), o: []byte{0x80, 0x80, 0x80, 0x80, 0x01}},
		{i: m(0x7FFFFFFFF), o: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x7F}},
		// 15
		{i: m(0x0800000000), o: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x01}},
		{i: m(0x3FFFFFFFFFF), o: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F}},
		{i: m(0x040000000000), o: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}},
		{i: m(0x1FFFFFFFFFFFF), o: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F}},
		{i: m(0x02000000000000), o: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}},
		// 20
		{i: m(0x0FFFFFFFFFFFFFF), o: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F}},
		{i: m(0x0100000000000000), o: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}},
		{i: m(0xFFFFFFFFFFFFFFFF), o: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
	}
	for i, test := range tests {
		b := test.i.AppendBinary(nil)
		if !bytes.Equal(test.o, b) {
			t.Errorf("%d expect %#v, got %#v", i, test.o, b)
		}
	}
}

func TestDecodeBinary(t *testing.T) {
	var tests = []struct {
		o DIR
		i []byte
		e string
	}{
		// 0
		{o: DIR{}, i: nil},
		{o: m(1), i: []byte{1}},
		{o: m(0), i: []byte{0}},
		{o: m(0, 0), i: []byte{0, 0}},
		{o: m(0, 1, 0), i: []byte{0, 1, 0}},
		// 5
		{o: m(0x00), i: []byte{0}},
		{o: m(0x7F), i: []byte{0x7F}},
		{o: m(0x80), i: []byte{0x80, 0x01}},
		{o: m(0x3FFF), i: []byte{0xFF, 0x7F}},
		{o: m(0x4000), i: []byte{0x80, 0x80, 0x01}},
		// 10
		{o: m(0x1FFFFF), i: []byte{0xff, 0xff, 0x7f}},
		{o: m(0x200000), i: []byte{0x80, 0x80, 0x80, 0x01}},
		{o: m(0x0FFFFFFF), i: []byte{0xFF, 0xFF, 0xFF, 0x7F}},
		{o: m(0x10000000), i: []byte{0x80, 0x80, 0x80, 0x80, 0x01}},
		{o: m(0x7FFFFFFFF), i: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x7F}},
		// 15
		{o: m(0x0800000000), i: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x01}},
		{o: m(0x3FFFFFFFFFF), i: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F}},
		{o: m(0x040000000000), i: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}},
		{o: m(0x1FFFFFFFFFFFF), i: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F}},
		{o: m(0x02000000000000), i: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}},
		// 20
		{o: m(0x0FFFFFFFFFFFFFF), i: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F}},
		{o: m(0x0100000000000000), i: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}},
		{o: m(0xFFFFFFFFFFFFFFFF), i: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{i: []byte{0x80, 0x80}, e: "invalid dir: truncated binary identifier"},
		{i: []byte{1, 2, 3, 4, 5, 6, 7, 8}, e: "invalid dir: too many identifiers"},
		// 25
		{i: []byte{1, 2, 3, 0, 5, 6, 7}, e: "invalid dir: identifier 3 is 0"},
	}
	for i, test := range tests {
		d, err := DecodeBinary(DIR{}, test.i)
		var errStr string
		if err != nil {
			errStr = err.Error()
		}
		if test.e != errStr {
			t.Errorf("%d expect error %q, got %q", i, test.e, errStr)
			continue
		}
		if test.e != "" {
			continue
		}
		if !d.Equal(test.o) {
			t.Errorf("%d expect %v, got %v", i, test.o, d)
		}
	}
}

func TestAppendURI(t *testing.T) {
	var tests = []struct {
		i DIR
		o string
	}{
		// 0
		{i: DIR{}, o: "dis:/"},
		{i: m(1), o: "dis:1/"},
		{i: m(0), o: "dis:0/"},
		{i: m(0, 0), o: "dis:0.0/"},
		{i: m(0, 1, 0), o: "dis:0.1.0/"},
		// 5
		{i: m(0x3F), o: "dis:_/"},
		{i: m(0x7F), o: "dis:_1/"},
		{i: m(0x80), o: "dis:02/"},
		{i: m(0x3FFF), o: "dis:__3/"},
		{i: m(0x4000), o: "dis:004/"},
		// 10
		{i: m(0x1FFFFF), o: "dis:___7/"},
		{i: m(0x200000), o: "dis:0008/"},
		{i: m(0x0FFFFFFF), o: "dis:____F/"},
		{i: m(0x10000000), o: "dis:0000G/"},
		{i: m(0x7FFFFFFFF), o: "dis:_____V/"},
		// 15
		{i: m(0x0800000000), o: "dis:00000W/"},
		{i: m(0x3FFFFFFFFFF), o: "dis:_______/"},
		{i: m(0x040000000000), o: "dis:00000001/"},
		{i: m(0x1FFFFFFFFFFFF), o: "dis:________1/"},
		{i: m(0x02000000000000), o: "dis:000000002/"},
		// 20
		{i: m(0x0FFFFFFFFFFFFFF), o: "dis:_________3/"},
		{i: m(0x0100000000000000), o: "dis:0000000004/"},
		{i: m(0xFFFFFFFFFFFFFFFF), o: "dis:__________F/"},
		{i: m(0, 1, 2, 0), o: "dis:0.1.2.0/"},
		{i: m(1, 0), o: "dis:1.0/"},
	}
	for i, test := range tests {
		b := test.i.AppendURI(nil)
		if test.o != string(b) {
			t.Errorf("%d expect %q for %q, got %q", i, test.o, test.i, string(b))
		}
	}

	s := m(1, 2, 3, 4, 5, 6, 7).URI()
	if exp := "dis:1.2.3.4.5.6.7/"; s != exp {
		t.Errorf("expect %q, got %q", exp, s)
	}

	s = m(1, 2345, 0).URI()
	if exp := "dis:1.fa.0/"; s != exp {
		t.Errorf("expect %q, got %q", exp, s)
	}
}

func TestDecodeURI(t *testing.T) {
	var tests = []struct {
		o DIR
		i string
		e string
	}{
		// 0
		{o: DIR{}, i: "dis:/"},
		{o: m(1), i: "dis:1/"},
		{o: m(0), i: "dis:0/"},
		{o: m(0, 0), i: "dis:0.0/"},
		{o: m(0, 1, 0), i: "dis:0.1.0/"},
		// 5
		{o: m(0x3F), i: "dis:_/"},
		{o: m(0x7F), i: "dis:_1/"},
		{o: m(0x80), i: "dis:02/"},
		{o: m(0x3FFF), i: "dis:__3/"},
		{o: m(0x4000), i: "dis:004/"},
		// 10
		{o: m(0x1FFFFF), i: "dis:___7/"},
		{o: m(0x200000), i: "dis:0008/"},
		{o: m(0x0FFFFFFF), i: "dis:____F/"},
		{o: m(0x10000000), i: "dis:0000G/"},
		{o: m(0x7FFFFFFFF), i: "dis:_____V/"},
		// 15
		{o: m(0x0800000000), i: "dis:00000W/"},
		{o: m(0x3FFFFFFFFFF), i: "dis:_______/"},
		{o: m(0x040000000000), i: "dis:00000001/"},
		{o: m(0x1FFFFFFFFFFFF), i: "dis:________1/"},
		{o: m(0x02000000000000), i: "dis:000000002/"},
		// 20
		{o: m(0x0FFFFFFFFFFFFFF), i: "dis:_________3/"},
		{o: m(0x0100000000000000), i: "dis:0000000004/"},
		{o: m(0xFFFF_FFFF_FFFF_FFFF), i: "dis:__________F/"},
		{o: m(0, 1, 2, 0), i: "dis:0.1.2.0/"},
		{i: "", e: "invalid dir: URI must start with \"dis:\" and end with \"/\""},
		// 25
		{i: "dis:", e: "invalid dir: URI must start with \"dis:\" and end with \"/\""},
		{i: "dis:.../", e: "invalid dir: identifier 1 is 0"},
		{i: "dis:.. /", e: "invalid dir: invalid characters in URI"},
		{i: "dis:__________G/", e: "invalid dir: identifier overflow"},
		{i: "dis:1.2.3.4.5.6.7/x", e: "invalid dir: URI must start with \"dis:\" and end with \"/\""},
		// 30
		{i: "dis:1.2.3.4.5.6.7./", e: "invalid dir: invalid characters in URI"},
		{i: "dis:1.2.3.4.5.6.7.1/", e: "invalid dir: invalid characters in URI"},
		{i: "dis:1.2.3.0.5.6.7/", e: "invalid dir: identifier 3 is 0"},
		{i: "dis:1.2.3..5.6.7/", e: "invalid dir: identifier 3 is 0"},
		{o: m(1, 0), i: "dis:1.0/"},
		// 35
		{o: m(0, 1), i: "dis:0.1/"},
		{i: "dis:0.1//", e: "invalid dir: character '/' is URI"},
	}
	for i, test := range tests {
		d, err := DecodeURI(DIR{}, test.i)
		var errStr string
		if err != nil {
			errStr = err.Error()
		}
		if test.e != errStr {
			t.Errorf("%d expect error %q, got %q", i, test.e, errStr)
			continue
		}
		if test.e != "" {
			continue
		}
		if !reflect.DeepEqual(test.o, d) {
			t.Errorf("%d expect %v, got %v", i, test.o, d)
		}
	}

	s := m(1, 2, 3, 4, 5, 6, 7).URI()
	if exp := "dis:1.2.3.4.5.6.7/"; s != exp {
		t.Errorf("expect %q, got %q", exp, s)
	}

	DecodeURI(m(1), "dis:1.2/")
}
