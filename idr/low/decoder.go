package low

import (
	"math"
	"math/bits"
	"time"

	"github.com/chmike/ditp/dir"
)

// Decoder is a low level IDR decoder.
type Decoder []byte

// Byte returns the byte in front of the remaining bytes.
func Byte(d Decoder) (Decoder, byte) {
	return d[1:], d[0]
}

// Bool returns the bool in front of the remaining bytes.
func Bool(d Decoder) (Decoder, bool) {
	return d[1:], d[0] != 0
}

// Bytes returns the n bytes in front of the remaining bytes.
func Bytes(d Decoder, n int) (Decoder, []byte) {
	return d[n:], d[:n]
}

// VarUint64 returns the uint64 in front of the remaining bytes.
func VarUint64(d Decoder) (Decoder, uint64) {
	var v uint64
	var s byte
	for i := 0; i < 8; i++ {
		t := d[i]
		if t < 0x80 {
			return d[i+1:], v | uint64(t)<<s
		}
		v |= uint64(t&0x7F) << s
		s += 7
	}
	return d[9:], v | (uint64(d[8]) << s)
}

// VarUint returns the compact encoded uint in front of the remaining bytes.
func VarUint(d Decoder) (Decoder, uint) {
	d, v := VarUint64(d)
	return d, uint(v)
}

// Tag returns the TagT in front of the remaining bytes.
func Tag(d Decoder) (Decoder, TagT) {
	d, v := VarUint64(d)
	return d, TagT(v)
}

// Size return the uint64 in front of the remaining bytes.
func Size(d Decoder) (Decoder, uint64) {
	return VarUint64(d)
}

// VarInt64 returns the int64 in front of the remaining bytes.
func VarInt64(d Decoder) (Decoder, int64) {
	d, x := VarUint64(d)
	if x&1 != 0 {
		return d, int64(^(x >> 1))
	}
	return d, int64(x >> 1)
}

// VarInt returns the int64 in front of the remaining bytes.
func VarInt(d Decoder) (Decoder, int) {
	d, v := VarInt64(d)
	return d, int(v)
}

// VarFloat returns the float64 in front of the remaining bytes.
func VarFloat(d Decoder) (Decoder, float64) {
	d, v := VarUint64(d)
	return d, math.Float64frombits(bits.ReverseBytes64(v))
}

// VarComplex returns the complex in front of the remaining bytes.
func VarComplex(d Decoder) (Decoder, complex128) {
	d, v1 := VarFloat(d)
	d, v2 := VarFloat(d)
	return d, complex(v1, v2)
}

// Uint8 returns the uint8 in front of the remaining bytes.
func Uint8(d Decoder) (Decoder, uint8) {
	return Byte(d)
}

// Uint16 returns the uint16 in front of the remaining bytesr.
func Uint16(d Decoder) (Decoder, uint16) {
	_ = d[1]
	return d[2:], uint16(d[0]) | uint16(d[1])<<8
}

// Uint32 returns the uint32 in front of the remaining bytes.
func Uint32(d Decoder) (Decoder, uint32) {
	_ = d[3]
	return d[4:], uint32(d[0]) | uint32(d[1])<<8 | uint32(d[2])<<16 | uint32(d[3])<<24
}

// Uint64 returns the uint64 in front of the remaining bytes.
func Uint64(d Decoder) (Decoder, uint64) {
	_ = d[7]
	return d[8:], uint64(d[0]) | uint64(d[1])<<8 | uint64(d[2])<<16 | uint64(d[3])<<24 |
		uint64(d[4])<<32 | uint64(d[5])<<40 | uint64(d[6])<<48 | uint64(d[7])<<56
}

// Int8 returns the int8 in front of the remaining bytes.
func Int8(d Decoder) (Decoder, int8) {
	d, v := Uint8(d)
	return d, int8(v)
}

// Int16 returns the int16 in front of the remaining bytes.
func Int16(d Decoder) (Decoder, int16) {
	d, v := Uint16(d)
	return d, int16(v)
}

// Int32 returns the int32 in front of the remaining bytes.
func Int32(d Decoder) (Decoder, int32) {
	d, v := Uint32(d)
	return d, int32(v)
}

// Int64 returns the int64 in front of the remaining bytes.
func Int64(d Decoder) (Decoder, int64) {
	d, v := Uint64(d)
	return d, int64(v)
}

// Float32 returns the float32 in front of the remaining bytes.
func Float32(d Decoder) (Decoder, float32) {
	d, v := Uint32(d)
	return d, math.Float32frombits(v)
}

// Float64 returns the float64 in front of the remaining bytes.
func Float64(d Decoder) (Decoder, float64) {
	d, v := Uint64(d)
	return d, math.Float64frombits(v)
}

// Complex64 returns the complex64 in front of the remaining bytes.
func Complex64(d Decoder) (Decoder, complex64) {
	d, v1 := Float32(d)
	d, v2 := Float32(d)
	return d, complex(v1, v2)
}

// Complex128 returns the complex128 in front of the remaining bytes.
func Complex128(d Decoder) (Decoder, complex128) {
	d, v1 := Float64(d)
	d, v2 := Float64(d)
	return d, complex(v1, v2)
}

// Bytes returns the []byte in front of the remaining bytes without
// making a copy.
func Blob(d Decoder, max uint64) (Decoder, []byte) {
	d, n := VarUint64(d)
	if n > max {
		panic("IDR decoder: data too big")
	}
	return d[n:], d[:n]
}

// String returns a copy of the string in front of the remaining bytes.
func String(d Decoder, max uint64) (Decoder, string) {
	d, v := Blob(d, max)
	return d, string(v)
}

// DIR returns the DIR in front of the remaining bytes and store it in b.
func DIR(d Decoder, b dir.DIR) (Decoder, dir.DIR) {
	l := int(d[0]) + 1
	b, err := dir.DecodeBinary(b, d[1:l])
	if err != nil {
		panic(err)
	}
	return d[l:], b
}

// VarTime returns the next value as a time or time.Zero when in error.
func VarTime(d Decoder) (Decoder, time.Time) {
	l := int(d[0]) + 1
	t := Decoder(d[1:l])
	t, utcsec := VarInt64(t)
	t, nano := VarUint64(t)
	var tm time.Time
	if len(t) == 0 {
		tm = time.Unix(utcsec, int64(nano)).UTC()
	} else {
		var offset int
		t, offset = VarInt(t)
		tm = time.Unix(utcsec, int64(nano)).In(time.FixedZone("", offset))
	}
	if len(t) != 0 {
		panic("IDR decoder: Time: trailing data")
	}
	return d[l:], tm
}

// Time returns the next value as a time or time.Zero when in error.
func Time(d Decoder) (Decoder, time.Time) {
	d, utcsec := Int64(d)
	d, nano := Uint32(d)
	d, offset := Int32(d)
	if offset == 0 {
		return d, time.Unix(utcsec, int64(nano)).UTC()
	}
	return d, time.Unix(utcsec, int64(nano)).In(time.FixedZone("", int(offset)))
}

// skipping methods

// SkipByte skips a byte value.
func SkipByte(d Decoder) Decoder {
	return d[1:]
}

// SkipBytes skips n bytes.
func SkipBytes(d Decoder, n uint64) Decoder {
	return d[n:]
}

// SkipBool skips a bool value.
func SkipBool(d Decoder) Decoder {
	return SkipByte(d)
}

// SkipVarUint64 skips a compact encoded uint64 value.
func SkipVarUint64(d Decoder) Decoder {
	for i := 0; i < 8; i++ {
		t := d[i]
		if t < 0x80 {
			return SkipBytes(d, uint64(i+1))
		}
	}
	return SkipBytes(d, 9)
}

// SkipVarUint skips a compact encoded uint  value.
func SkipVarUint(d Decoder) Decoder {
	return SkipVarUint64(d)
}

// SkipSize skips a Size value.
func SkipSize(d Decoder) Decoder {
	return SkipVarUint(d)
}

// SkipVarInt64 skips a compact encoded int64 value.
func SkipVarInt64(d Decoder) Decoder {
	return SkipVarUint(d)
}

// SkipVarInt skips a compact encoded int value.
func SkipVarInt(d Decoder) Decoder {
	return SkipVarUint(d)
}

// SkipVarFloat skips a compact encoded float value.
func SkipVarFloat(d Decoder) Decoder {
	return SkipVarUint(d)
}

// SkipVarComplex skips a compact encoded complex value.
func SkipVarComplex(d Decoder) Decoder {
	return SkipVarUint(SkipVarUint(d))
}

// SkipUint8 skips a uint8 value.
func SkipUint8(d Decoder) Decoder {
	return SkipByte(d)
}

// SkipUint16 skips a uint16 value.
func SkipUint16(d Decoder) Decoder {
	return SkipBytes(d, 2)
}

// SkipUint32 skips a uint32 value.
func SkipUint32(d Decoder) Decoder {
	return SkipBytes(d, 4)
}

// SkipUint64 skips a uint64 value.
func SkipUint64(d Decoder) Decoder {
	return SkipBytes(d, 8)
}

// SkipInt8 skips a int8 value.
func SkipInt8(d Decoder) Decoder {
	return SkipUint8(d)
}

// SkipInt16 skips a int16 value.
func SkipInt16(d Decoder) Decoder {
	return SkipUint16(d)
}

// SkipInt32 skips a int32 value.
func SkipInt32(d Decoder) Decoder {
	return SkipUint32(d)

}

// SkipInt64 skips a int64 value.
func SkipInt64(d Decoder) Decoder {
	return SkipUint64(d)
}

// SkipFloat32 skips a float32 value.
func SkipFloat32(d Decoder) Decoder {
	return SkipUint32(d)
}

// SkipFloat64 skips a float64 value.
func SkipFloat64(d Decoder) Decoder {
	return SkipUint64(d)
}

// SkipComplex64 skips a complex64 value.
func SkipComplex64(d Decoder) Decoder {
	return SkipUint64(d)
}

// SkipComplex128 skips a complex128 value.
func SkipComplex128(d Decoder) Decoder {
	return SkipBytes(d, 16)
}

// SkipBlob skips a blob value.
func SkipBlob(d Decoder, max uint64) Decoder {
	d, n := VarUint64(d)
	if n > max {
		panic("IDR decoder: data too big")
	}
	return SkipBytes(d, n)
}

// SkipString skips a string value.
func SkipString(d Decoder, max uint64) Decoder {
	return SkipBlob(d, max)
}

// SkipDIR skips a DIR value.
func SkipDIR(d Decoder) Decoder {
	return SkipBlob(d, dir.MaxBinaryLen)
}

// SkipVarTime skips a time value.
func SkipVarTime(d Decoder) Decoder {
	return SkipBlob(d, 30)
}

// SkipTime skips a time value.
func SkipTime(d Decoder) Decoder {
	return SkipBytes(d, 16)
}
