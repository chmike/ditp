package low

import (
	"bytes"
	"testing"
	"time"

	"github.com/chmike/ditp/dir"
)

// tme return the given time string as time.Time.
func tme(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		panic(err)
	}
	return t
}

func TestEncoder(t *testing.T) {
	tests := []struct {
		t TagT
		i any
		o []byte
	}{
		// 0
		{t: BoolTag, i: true, o: []byte{1}},
		{t: BoolTag, i: false, o: []byte{0}},
		{t: ByteTag, i: byte(0xFE), o: []byte{0xFE}},
		{t: BytesTag, i: []byte{1, 2}, o: []byte{1, 2}},
		{t: BytesTag, i: []byte("IDR0"), o: []byte{'I', 'D', 'R', '0'}},
		// 5
		{t: VarUint64Tag, i: uint64(0x7f), o: []byte{0x7F}},
		{t: VarUint64Tag, i: uint64(0x3fff), o: []byte{0xff, 0x7f}},
		{t: VarUint64Tag, i: uint64(0x1f_ffff), o: []byte{0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, i: uint64(0xfff_ffff), o: []byte{0xff, 0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, i: uint64(0x7_ffff_ffff), o: []byte{0xff, 0xff, 0xff, 0xff, 0x7f}},
		// 10
		{t: VarUint64Tag, i: uint64(0x3ff_ffff_ffff), o: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, i: uint64(0x1_ffff_ffff_ffff), o: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, i: uint64(0xff_ffff_ffff_ffff), o: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{t: VarUint64Tag, i: uint64(0xffff_ffff_ffff_ffff), o: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{t: VarInt64Tag, i: int64(-63), o: []byte{0x7d}},
		// 15
		{t: VarFloatTag, i: float64(2.), o: []byte{0x40}},
		{t: VarFloatTag, i: float64(-2.), o: []byte{0xC0, 0x01}},
		{t: VarFloatTag, i: float64(.5), o: []byte{0xBF, 0xC0, 0x03}},
		{t: VarComplexTag, i: 1 + 5i, o: []byte{0xbf, 0xe0, 0x3, 0xc0, 0x28}},
		{t: Uint8Tag, i: uint8(0x80), o: []byte{0x80}},
		// 20
		{t: Uint16Tag, i: uint16(0x8180), o: []byte{0x80, 0x81}},
		{t: Uint32Tag, i: uint32(0x83828180), o: []byte{0x80, 0x81, 0x82, 0x83}},
		{t: Uint64Tag, i: uint64(0x8786858483828180), o: []byte{0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87}},
		{t: Int8Tag, i: int8(-1), o: []byte{0xFF}},
		{t: Int16Tag, i: int16(-256), o: []byte{0x00, 0xFF}},
		// 25
		{t: Int32Tag, i: int32(-55555), o: []byte{0xFD, 0x26, 0xFF, 0xFF}},
		{t: Int64Tag, i: int64(-876543210), o: []byte{0x16, 0x03, 0xC1, 0xCB, 0xFF, 0xff, 0xFF, 0xFF}},
		{t: Float32Tag, i: float32(2.), o: []byte{0x00, 0x00, 0x00, 0x40}},
		{t: Float64Tag, i: float64(-2.), o: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC0}},
		{t: Complex64Tag, i: complex64(1. + 5i), o: []byte{0x0, 0x0, 0x80, 0x3f, 0x0, 0x0, 0xa0, 0x40}},
		// 30
		{t: Complex128Tag, i: 1. + 5i, o: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x3f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14, 0x40}},
		{t: SizeTag, i: uint64(0xFF), o: []byte{0xFF, 0x01}},
		{t: BlobTag, i: []byte{1, 2, 3, 4}, o: []byte{4, 1, 2, 3, 4}},
		{t: StringTag, i: "hello", o: []byte{5, 'h', 'e', 'l', 'l', 'o'}},
		{t: DIRTag, i: dir.MustMake(1, 2, 3, 4, 5, 6, 7), o: []byte{0x7, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7}},
		// 35
		{t: VarTimeTag, i: tme("2023-10-06T10:00:00Z"), o: []byte{0x6, 0xc0, 0xea, 0xfe, 0xd1, 0xc, 0x0}},
		{t: VarTimeTag, i: tme("2023-10-06T10:00:00.5+01:00"), o: []byte{0xc, 0xa0, 0xb2, 0xfe, 0xd1, 0xc, 0x80, 0xca, 0xb5, 0xee, 0x1, 0xa0, 0x38}},
		{t: VarTimeTag, i: tme("2023-10-06T10:00:00-07:00"), o: []byte{0x9, 0xa0, 0xf4, 0x81, 0xd2, 0xc, 0x0, 0xdf, 0x89, 0x3}},
		{t: VarTimeTag, i: tme("2023-10-06T10:00:00-05:00"), o: []byte{0x9, 0xe0, 0x83, 0x81, 0xd2, 0xc, 0x0, 0x9f, 0x99, 0x2}},
		{t: VarTimeTag, i: tme("2023-10-06T10:00:00+02:00"), o: []byte{0x8, 0x80, 0xfa, 0xfd, 0xd1, 0xc, 0x0, 0xc0, 0x70}},
		//40
		{t: TimeTag, i: tme("2023-10-06T10:00:00Z"), o: []byte{0xa0, 0xda, 0x1f, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		{t: TimeTag, i: tme("2023-10-06T10:00:00.5+01:00"), o: []byte{0x90, 0xcc, 0x1f, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x65, 0xcd, 0x1d, 0x10, 0xe, 0x0, 0x0}},
		{t: TimeTag, i: tme("2023-10-06T10:00:00-07:00"), o: []byte{0x10, 0x3d, 0x20, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x90, 0x9d, 0xff, 0xff}},
		{t: TimeTag, i: tme("2023-10-06T10:00:00-05:00"), o: []byte{0xf0, 0x20, 0x20, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xb0, 0xb9, 0xff, 0xff}},
		{t: TimeTag, i: tme("2023-10-06T10:00:00+02:00"), o: []byte{0x80, 0xbe, 0x1f, 0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x1c, 0x0, 0x0}},
		//45
		{t: DIRTag, i: dir.DIR{}, o: []byte{0x00}},
		{t: NoneTag, i: TagT(0), o: []byte{0}},
		{t: NoneTag, i: TagT(12345), o: []byte{0xB9, 0x60}},
		{t: VarUintTag, i: uint(123), o: []byte{0x7B}},
	}
	e := Encoder(make([]byte, 0, 64))
	for i, test := range tests {
		e = Reset(e)
		switch test.t {
		case NoneTag:
			e = AppendTag(e, test.i.(TagT))
		case BoolTag:
			e = AppendBool(e, test.i.(bool))
		case ByteTag:
			e = AppendByte(e, test.i.(byte))
		case BytesTag:
			e = AppendBytes(e, test.i.([]byte)...)
		case VarUintTag:
			e = AppendVarUint(e, test.i.(uint))
		case VarIntTag:
			e = AppendVarInt(e, test.i.(int))
		case VarUint64Tag:
			e = AppendVarUint64(e, test.i.(uint64))
		case VarInt64Tag:
			e = AppendVarInt64(e, test.i.(int64))
		case SizeTag:
			e = AppendSize(e, test.i.(uint64))
		case VarFloatTag:
			e = AppendVarFloat(e, test.i.(float64))
		case VarComplexTag:
			e = AppendVarComplex(e, test.i.(complex128))
		case Uint8Tag:
			e = AppendUint8(e, test.i.(uint8))
		case Uint16Tag:
			e = AppendUint16(e, test.i.(uint16))
		case Uint32Tag:
			e = AppendUint32(e, test.i.(uint32))
		case Uint64Tag:
			e = AppendUint64(e, test.i.(uint64))
		case Int8Tag:
			e = AppendInt8(e, test.i.(int8))
		case Int16Tag:
			e = AppendInt16(e, test.i.(int16))
		case Int32Tag:
			e = AppendInt32(e, test.i.(int32))
		case Int64Tag:
			e = AppendInt64(e, test.i.(int64))
		case Float32Tag:
			e = AppendFloat32(e, test.i.(float32))
		case Float64Tag:
			e = AppendFloat64(e, test.i.(float64))
		case Complex64Tag:
			e = AppendComplex64(e, test.i.(complex64))
		case Complex128Tag:
			e = AppendComplex128(e, test.i.(complex128))
		case BlobTag:
			e = AppendBlob(e, test.i.([]byte))
		case StringTag:
			e = AppendString(e, test.i.(string))
		case DIRTag:
			e = AppendDIR(e, test.i.(dir.DIR))
		case VarTimeTag:
			e = AppendVarTime(e, test.i.(time.Time))
		case TimeTag:
			e = AppendTime(e, test.i.(time.Time))
		default:
			t.Errorf("%3d unsupported type %T", i, test.i)
			continue
		}
		if len(e) != len(test.o) {
			t.Errorf("%3d expected len %d, got %d", i, len(test.o), len(e))
		}
		if !bytes.Equal(e, test.o) {
			t.Errorf("%3d expected encoding %#v, got %#v", i, test.o, e)
		}
	}
}
