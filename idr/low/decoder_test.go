package low

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/chmike/ditp/dir"
)

func TestDecoder(t *testing.T) {
	tests := []struct {
		t TagT
		o any
		i []byte
	}{
		// 0
		{t: BoolTag, o: true, i: []byte{1}},
		{t: BoolTag, o: false, i: []byte{0}},
		{t: ByteTag, o: byte(0xFE), i: []byte{0xFE}},
		{t: BytesTag, o: []byte{1, 2}, i: []byte{1, 2}},
		{t: BytesTag, o: []byte("IDR0"), i: []byte{'I', 'D', 'R', '0'}},
		// 5
		{t: VarUint64Tag, o: uint64(0x7f), i: []byte{0x7F}},
		{t: VarUint64Tag, o: uint64(0x3fff), i: []byte{0xff, 0x7f}},
		{t: VarUint64Tag, o: uint64(0x1f_ffff), i: []byte{0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, o: uint64(0xfff_ffff), i: []byte{0xff, 0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, o: uint64(0x7_ffff_ffff), i: []byte{0xff, 0xff, 0xff, 0xff, 0x7f}},
		// 10
		{t: VarUint64Tag, o: uint64(0x3ff_ffff_ffff), i: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, o: uint64(0x1_ffff_ffff_ffff), i: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, o: uint64(0xff_ffff_ffff_ffff), i: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, o: uint64(0xffff_ffff_ffff_ffff), i: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{t: VarInt64Tag, o: int64(-63), i: []byte{0x7d}},
		// 15
		{t: VarFloatTag, o: float64(2.), i: []byte{0x40}},
		{t: VarFloatTag, o: float64(-2.), i: []byte{0xC0, 0x01}},
		{t: VarFloatTag, o: float64(.5), i: []byte{0xBF, 0xC0, 0x03}},
		{t: VarComplexTag, o: 1 + 5i, i: []byte{0xbf, 0xe0, 0x3, 0xc0, 0x28}},
		{t: Uint8Tag, o: uint8(0x80), i: []byte{0x80}},
		// 20
		{t: Uint16Tag, o: uint16(0x8180), i: []byte{0x80, 0x81}},
		{t: Uint32Tag, o: uint32(0x83828180), i: []byte{0x80, 0x81, 0x82, 0x83}},
		{t: Uint64Tag, o: uint64(0x8786858483828180), i: []byte{0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87}},
		{t: Int8Tag, o: int8(-1), i: []byte{0xFF}},
		{t: Int16Tag, o: int16(-256), i: []byte{0x00, 0xFF}},
		// 25
		{t: Int32Tag, o: int32(-55555), i: []byte{0xFD, 0x26, 0xFF, 0xFF}},
		{t: Int64Tag, o: int64(-876543210), i: []byte{0x16, 0x03, 0xC1, 0xCB, 0xFF, 0xff, 0xFF, 0xFF}},
		{t: Float32Tag, o: float32(2.), i: []byte{0x00, 0x00, 0x00, 0x40}},
		{t: Float64Tag, o: float64(-2.), i: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC0}},
		{t: Complex64Tag, o: complex64(1. + 5i), i: []byte{0x0, 0x0, 0x80, 0x3f, 0x0, 0x0, 0xa0, 0x40}},
		// 30
		{t: Complex128Tag, o: 1. + 5i, i: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x3f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14, 0x40}},
		{t: SizeTag, o: uint64(0xFF), i: []byte{0xFF, 0x01}},
		{t: BlobTag, o: []byte{1, 2, 3, 4}, i: []byte{4, 1, 2, 3, 4}},
		{t: StringTag, o: "hello", i: []byte{5, 'h', 'e', 'l', 'l', 'o'}},
		{t: DIRTag, o: dir.MustMake(1, 2, 3, 4, 5, 6, 7), i: []byte{0x7, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7}},
		// 35
		{t: VarTimeTag, o: "2023-10-06T10:00:00+00:00", i: []byte{0x6, 0xc0, 0xea, 0xfe, 0xd1, 0xc, 0x0}},
		{t: VarTimeTag, o: "2023-10-06T10:00:00.5+01:00", i: []byte{0xc, 0xa0, 0xb2, 0xfe, 0xd1, 0xc, 0x80, 0xca, 0xb5, 0xee, 0x1, 0xa0, 0x38}},
		{t: VarTimeTag, o: "2023-10-06T10:00:00-07:00", i: []byte{0x9, 0xa0, 0xf4, 0x81, 0xd2, 0xc, 0x0, 0xdf, 0x89, 0x3}},
		{t: VarTimeTag, o: "2023-10-06T10:00:00-05:00", i: []byte{0x9, 0xe0, 0x83, 0x81, 0xd2, 0xc, 0x0, 0x9f, 0x99, 0x2}},
		{t: VarTimeTag, o: "2023-10-06T10:00:00+02:00", i: []byte{0x8, 0x80, 0xfa, 0xfd, 0xd1, 0xc, 0x0, 0xc0, 0x70}},
		//40
		{t: TimeTag, o: "2023-10-06T10:00:00+00:00", i: []byte{0xa0, 0xda, 0x1f, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		{t: TimeTag, o: "2023-10-06T10:00:00.5+01:00", i: []byte{0x90, 0xcc, 0x1f, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x65, 0xcd, 0x1d, 0x10, 0xe, 0x0, 0x0}},
		{t: TimeTag, o: "2023-10-06T10:00:00-07:00", i: []byte{0x10, 0x3d, 0x20, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x90, 0x9d, 0xff, 0xff}},
		{t: TimeTag, o: "2023-10-06T10:00:00-05:00", i: []byte{0xf0, 0x20, 0x20, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xb0, 0xb9, 0xff, 0xff}},
		{t: TimeTag, o: "2023-10-06T10:00:00+02:00", i: []byte{0x80, 0xbe, 0x1f, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x1c, 0x0, 0x0}},
		//45
		{t: DIRTag, o: dir.DIR{}, i: []byte{0x00}},
		{t: NoneTag, o: TagT(0), i: []byte{0}},
		{t: NoneTag, o: TagT(12345), i: []byte{0xB9, 0x60}},
		{t: VarUintTag, o: uint(123), i: []byte{0x7B}},
	}
	for i, test := range tests {
		d := Decoder(test.i)
		var v any
		switch test.t {
		case NoneTag:
			d, v = Tag(d)
		case BoolTag:
			d, v = Bool(d)
		case ByteTag:
			d, v = Byte(d)
		case BytesTag:
			d, v = Bytes(d, len(test.i))
		case VarUintTag:
			d, v = VarUint(d)
		case VarIntTag:
			d, v = VarInt(d)
		case VarUint64Tag:
			d, v = VarUint64(d)
		case VarInt64Tag:
			d, v = VarInt64(d)
		case SizeTag:
			d, v = Size(d)
		case VarFloatTag:
			d, v = VarFloat(d)
		case VarComplexTag:
			d, v = VarComplex(d)
		case Uint8Tag:
			d, v = Uint8(d)
		case Uint16Tag:
			d, v = Uint16(d)
		case Uint32Tag:
			d, v = Uint32(d)
		case Uint64Tag:
			d, v = Uint64(d)
		case Int8Tag:
			d, v = Int8(d)
		case Int16Tag:
			d, v = Int16(d)
		case Int32Tag:
			d, v = Int32(d)
		case Int64Tag:
			d, v = Int64(d)
		case Float32Tag:
			d, v = Float32(d)
		case Float64Tag:
			d, v = Float64(d)
		case Complex64Tag:
			d, v = Complex64(d)
		case Complex128Tag:
			d, v = Complex128(d)
		case BlobTag:
			d, v = Blob(d, 10)
		case StringTag:
			d, v = String(d, 10)
		case DIRTag:
			d, v = DIR(d, dir.DIR{})
		case VarTimeTag:
			d, v = VarTime(d)
			v = v.(time.Time).Format("2006-01-02T15:04:05.999999999-07:00")
		case TimeTag:
			d, v = Time(d)
			v = v.(time.Time).Format("2006-01-02T15:04:05.999999999-07:00")
		default:
			t.Errorf("%3d unsupported type tag %v", i, test.t)
			continue
		}
		if len(d) != 0 {
			t.Errorf("%3d expected len %d, got %d", i, 0, len(d))
		}
		if !reflect.DeepEqual(v, test.o) {
			t.Errorf("%3d expected value %#v, got %#v", i, test.o, v)
		}
	}

	tmi := time.Now()
	e := Encoder{}
	e = PutVarTime(e, tmi)
	_, tmo := VarTime(Decoder(e))
	if !tmi.Equal(tmo) {
		t.Errorf("expect %q equal %q", tmi, tmo)
	}

	tmi = time.Now()
	e = Encoder{}
	e = PutTime(e, tmi)
	_, tmo = Time(Decoder(e))
	if !tmi.Equal(tmo) {
		t.Errorf("expect %q equal %q", tmi, tmo)
	}
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

func TestPanics(t *testing.T) {
	d := Decoder([]byte{8, 1, 2, 3, 4, 5, 6, 7, 8})
	if !doesPanic(func() { DIR(d, dir.DIR{}) }) {
		t.Error("expect DIR panics")
	}

	d = Decoder([]byte{0xb, 0xc0, 0xea, 0xfe, 0xd1, 0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1})
	if !doesPanic(func() { VarTime(d) }) {
		t.Error("expect VarTime panics")
	}

	d = Decoder([]byte{4, 1, 2, 3, 4})
	if !doesPanic(func() { Blob(d, 3) }) {
		t.Error("expect Blob panics")
	}

	d = Decoder([]byte{4, 1, 2, 3, 4})
	if !doesPanic(func() { SkipBlob(d, 3) }) {
		t.Error("expect SkipBlob panics")
	}
	d = Decoder([]byte{0x11, 0xa0, 0xf4, 0x81, 0xd2, 0xc, 0x0, 0x0, 0x0, 0x0, 0xdf, 0x89, 0x3, 0x3, 0x4d, 0x44, 0x54, 0x00})
	if !doesPanic(func() { VarTime(d) }) {
		t.Error("expect VarTime panics")
	}
}

func TestSkip(t *testing.T) {
	tests := []struct {
		t TagT
		i []byte
	}{
		// 0
		{t: BoolTag, i: []byte{1}},
		{t: BoolTag, i: []byte{0}},
		{t: ByteTag, i: []byte{0xFE}},
		{t: BytesTag, i: []byte{1, 2}},
		{t: BytesTag, i: []byte{'I', 'D', 'R', '0'}},
		// 5
		{t: VarUint64Tag, i: []byte{0x7F}},
		{t: VarUint64Tag, i: []byte{0xff, 0x7f}},
		{t: VarUint64Tag, i: []byte{0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, i: []byte{0xff, 0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, i: []byte{0xff, 0xff, 0xff, 0xff, 0x7f}},
		// 10
		{t: VarUint64Tag, i: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, i: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, i: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, i: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{t: VarInt64Tag, i: []byte{0x7d}},
		// 15
		{t: VarFloatTag, i: []byte{0x40}},
		{t: VarFloatTag, i: []byte{0xC0, 0x01}},
		{t: VarFloatTag, i: []byte{0xBF, 0xC0, 0x03}},
		{t: VarComplexTag, i: []byte{0xbf, 0xe0, 0x3, 0xc0, 0x28}},
		{t: Uint8Tag, i: []byte{0x80}},
		// 20
		{t: Uint16Tag, i: []byte{0x80, 0x81}},
		{t: Uint32Tag, i: []byte{0x80, 0x81, 0x82, 0x83}},
		{t: Uint64Tag, i: []byte{0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87}},
		{t: Int8Tag, i: []byte{0xFF}},
		{t: Int16Tag, i: []byte{0x00, 0xFF}},
		// 25
		{t: Int32Tag, i: []byte{0xFD, 0x26, 0xFF, 0xFF}},
		{t: Int64Tag, i: []byte{0x16, 0x03, 0xC1, 0xCB, 0xFF, 0xff, 0xFF, 0xFF}},
		{t: Float32Tag, i: []byte{0x00, 0x00, 0x00, 0x40}},
		{t: Float64Tag, i: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC0}},
		{t: Complex64Tag, i: []byte{0x0, 0x0, 0x80, 0x3f, 0x0, 0x0, 0xa0, 0x40}},
		// 30
		{t: Complex128Tag, i: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x3f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14, 0x40}},
		{t: SizeTag, i: []byte{0xFF, 0x01}},
		{t: BlobTag, i: []byte{4, 1, 2, 3, 4}},
		{t: StringTag, i: []byte{5, 'h', 'e', 'l', 'l', 'o'}},
		{t: DIRTag, i: []byte{0x7, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7}},
		// 35
		{t: VarTimeTag, i: []byte{0x6, 0xc0, 0xea, 0xfe, 0xd1, 0xc, 0x0}},
		{t: VarTimeTag, i: []byte{0xc, 0xa0, 0xb2, 0xfe, 0xd1, 0xc, 0x80, 0x94, 0xeb, 0xdc, 0x3, 0xa0, 0x38}},
		{t: VarTimeTag, i: []byte{0x9, 0xa0, 0xf4, 0x81, 0xd2, 0xc, 0x0, 0xdf, 0x89, 0x3}},
		{t: VarTimeTag, i: []byte{0x9, 0xe0, 0x83, 0x81, 0xd2, 0xc, 0x0, 0x9f, 0x99, 0x2}},
		{t: VarTimeTag, i: []byte{0x8, 0x80, 0xfa, 0xfd, 0xd1, 0xc, 0x0, 0xc0, 0x70}},
		//40
		{t: TimeTag, i: []byte{0xa0, 0xda, 0x1f, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		{t: TimeTag, i: []byte{0x90, 0xcc, 0x1f, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x65, 0xcd, 0x1d, 0x10, 0xe, 0x0, 0x0}},
		{t: TimeTag, i: []byte{0x10, 0x3d, 0x20, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x90, 0x9d, 0xff, 0xff}},
		{t: TimeTag, i: []byte{0xf0, 0x20, 0x20, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xb0, 0xb9, 0xff, 0xff}},
		{t: TimeTag, i: []byte{0x80, 0xbe, 0x1f, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x1c, 0x0, 0x0}},
		//45
		{t: DIRTag, i: []byte{0x00}},
		{t: VarIntTag, i: []byte{0x7B}},
	}
	for i, test := range tests {
		d := Decoder(test.i)
		if !bytes.Equal(test.i, d) {
			t.Errorf("%3d expected value %#v, got %#v", i, test.i, d)
		}
		switch test.t {
		case BoolTag:
			d = SkipBool(d)
		case ByteTag:
			d = SkipByte(d)
		case BytesTag:
			d = SkipBytes(d, uint64(len(test.i)))
		case VarUintTag:
			d = SkipVarUint(d)
		case VarIntTag:
			d = SkipVarInt(d)
		case VarUint64Tag:
			d = SkipVarUint64(d)
		case VarInt64Tag:
			d = SkipVarInt64(d)
		case SizeTag:
			d = SkipSize(d)
		case VarFloatTag:
			d = SkipVarFloat(d)
		case VarComplexTag:
			d = SkipVarComplex(d)
		case Uint8Tag:
			d = SkipUint8(d)
		case Uint16Tag:
			d = SkipUint16(d)
		case Uint32Tag:
			d = SkipUint32(d)
		case Uint64Tag:
			d = SkipUint64(d)
		case Int8Tag:
			d = SkipInt8(d)
		case Int16Tag:
			d = SkipInt16(d)
		case Int32Tag:
			d = SkipInt32(d)
		case Int64Tag:
			d = SkipInt64(d)
		case Float32Tag:
			d = SkipFloat32(d)
		case Float64Tag:
			d = SkipFloat64(d)
		case Complex64Tag:
			d = SkipComplex64(d)
		case Complex128Tag:
			d = SkipComplex128(d)
		case BlobTag:
			d = SkipBlob(d, 10)
		case StringTag:
			d = SkipString(d, 10)
		case DIRTag:
			d = SkipDIR(d)
		case VarTimeTag:
			d = SkipVarTime(d)
		case TimeTag:
			d = SkipTime(d)
		default:
			t.Errorf("%3d unsupported type tag %v", i, test.t)
			continue
		}
		if len(d) != 0 {
			t.Errorf("%3d expected len %d, got %d", i, 0, len(d))
		}
	}
}
