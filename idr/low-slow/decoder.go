package low

import (
	"math"
	"math/bits"
	"time"

	"github.com/chmike/ditp/dir"
)

// Decoder is a low level IDR decoder.
type Decoder struct {
	b []byte
}

// Decode set the decoding buffer to b.
func (d *Decoder) Decode(b []byte) {
	d.b = b
}

// Decode returns a decoder set to decode b.
func Decode(b []byte) Decoder {
	return Decoder{b}
}

// Len returns the number of bytes left to decode.
func (d Decoder) Len() int {
	return len(d.b)
}

// Peek returns the remaining bytes to decode without making a copy.
func (d Decoder) Peek() []byte {
	return d.b
}

// Bytes returns the n bytes in front of the remaining bytes.
func (d *Decoder) Bytes(n int) []byte {
	v := d.b[:n]
	d.b = d.b[n:]
	return v
}

// Byte returns the byte in front of the remaining bytes.
func (d *Decoder) Byte() byte {
	v := d.b[0]
	d.b = d.b[1:]
	return v
}

// Bool returns the bool in front of the remaining bytes.
func (d *Decoder) Bool() bool {
	return d.Byte() == 1
}

// VarUint returns the uint64 in front of the remaining bytes.
func (d *Decoder) VarUint() uint64 {
	var v uint64
	var s byte
	for i := 0; i < 8; i++ {
		t := d.b[i]
		if t < 0x80 {
			d.SkipBytes(uint64(i + 1))
			return v | uint64(t)<<s
		}
		v |= uint64(t&0x7F) << s
		s += 7
	}
	v |= uint64(d.b[8]) << s
	d.SkipBytes(9)
	return v
}

// Tag returns the Tag in front of the remaining bytes.
func (d *Decoder) Tag() Tag {
	return Tag(d.VarUint())
}

// Size return the uint64 in front of the remaining bytes.
func (d *Decoder) Size() uint64 {
	return d.VarUint()
}

// VarInt returns the int64 in front of the remaining bytes.
func (d *Decoder) VarInt() int64 {
	x := d.VarUint()
	if x&1 != 0 {
		return int64(^(x >> 1))
	}
	return int64(x >> 1)
}

// VarFloat returns the float64 in front of the remaining bytes.
func (d *Decoder) VarFloat() float64 {
	return math.Float64frombits(bits.ReverseBytes64(d.VarUint()))
}

// VarComplex returns the complex in front of the remaining bytes.
func (d *Decoder) VarComplex() complex128 {
	return complex(d.VarFloat(), d.VarFloat())
}

// Uint8 returns the uint8 in front of the remaining bytes.
func (d *Decoder) Uint8() uint8 {
	return d.Byte()
}

// Uint16 returns the uint16 in front of the remaining bytesr.
func (d *Decoder) Uint16() uint16 {
	_ = d.b[1]
	v := uint16(d.b[0]) | uint16(d.b[1])<<8
	d.b = d.b[2:]
	return v
}

// Uint32 returns the uint32 in front of the remaining bytes.
func (d *Decoder) Uint32() uint32 {
	_ = d.b[3]
	v := uint32(d.b[0]) | uint32(d.b[1])<<8 | uint32(d.b[2])<<16 | uint32(d.b[3])<<24
	d.b = d.b[4:]
	return v
}

// Uint64 returns the uint64 in front of the remaining bytes.
func (d *Decoder) Uint64() uint64 {
	_ = d.b[7]
	v := uint64(d.b[0]) | uint64(d.b[1])<<8 | uint64(d.b[2])<<16 | uint64(d.b[3])<<24 |
		uint64(d.b[4])<<32 | uint64(d.b[5])<<40 | uint64(d.b[6])<<48 | uint64(d.b[7])<<56
	d.b = d.b[8:]
	return v
}

// Int8 returns the int8 in front of the remaining bytes.
func (d *Decoder) Int8() int8 {
	return int8(d.Uint8())
}

// Int16 returns the int16 in front of the remaining bytes.
func (d *Decoder) Int16() int16 {
	return int16(d.Uint16())
}

// Int32 returns the int32 in front of the remaining bytes.
func (d *Decoder) Int32() int32 {
	return int32(d.Uint32())
}

// Int64 returns the int64 in front of the remaining bytes.
func (d *Decoder) Int64() int64 {
	return int64(d.Uint64())
}

// Float32 returns the float32 in front of the remaining bytes.
func (d *Decoder) Float32() float32 {
	return math.Float32frombits(d.Uint32())
}

// Float64 returns the float64 in front of the remaining bytes.
func (d *Decoder) Float64() float64 {
	return math.Float64frombits(d.Uint64())
}

// Complex64 returns the complex64 in front of the remaining bytes.
func (d *Decoder) Complex64() complex64 {
	return complex(d.Float32(), d.Float32())
}

// Complex128 returns the complex128 in front of the remaining bytes.
func (d *Decoder) Complex128() complex128 {
	return complex(d.Float64(), d.Float64())
}

// Bytes returns the []byte in front of the remaining bytes without
// making a copy.
func (d *Decoder) Blob(max uint64) []byte {
	n := d.VarUint()
	if n > max {
		panic("IDR decoder: data too big")
	}
	v := d.b[:n]
	d.SkipBytes(n)
	return v
}

// String returns a copy of the string in front of the remaining bytes.
func (d *Decoder) String(max uint64) string {
	return string(d.Blob(max))
}

// DIR returns the DIR in front of the remaining bytes and store it in b.
func (d *Decoder) DIR(b dir.DIR) dir.DIR {
	l := int(d.b[0]) + 1
	b, err := dir.DecodeBinary(b, d.b[1:l])
	if err != nil {
		panic(err)
	}
	d.SkipBytes(uint64(l))
	return b
}

// VarTime returns the next value as a time or time.Zero when in error.
func (d *Decoder) VarTime() time.Time {
	l := int(d.b[0]) + 1
	t := Decode(d.b[1:l])
	utcsec := t.VarInt()
	nano := t.Uint32()
	offset := int(t.VarInt())
	var tm time.Time
	if offset == 0 {
		tm = time.Unix(utcsec, int64(nano)).UTC()
	} else {
		tm = time.Unix(utcsec, int64(nano)).In(time.FixedZone(t.String(64), offset))
	}
	if t.Len() != 0 {
		panic("IDR decoder: Time: trailing data")
	}
	d.SkipBytes(uint64(l))
	return tm
}

// Time returns the next value as a time or time.Zero when in error.
func (d *Decoder) Time() time.Time {
	utcsec := d.Int64()
	nano := d.Uint32()
	offset := d.Int32()
	tm := time.Unix(utcsec, int64(nano)).UTC()
	if offset != 0 {
		tm = tm.In(time.FixedZone("", int(offset)))
	}
	return tm
}

// skipping methods

// SkipByte skips a byte value.
func (d *Decoder) SkipByte() {
	d.b = d.b[1:]
}

// SkipBytes skips n bytes.
func (d *Decoder) SkipBytes(n uint64) {
	d.b = d.b[n:]
}

// SkipBool skips a bool value.
func (d *Decoder) SkipBool() {
	d.SkipByte()
}

// SkipVarUint skips a VarUint value.
func (d *Decoder) SkipVarUint() {
	for i := 0; i < 8; i++ {
		t := d.b[i]
		if t < 0x80 {
			d.SkipBytes(uint64(i + 1))
			return
		}
	}
	d.SkipBytes(9)
}

// SkipSize skips a Size value.
func (d *Decoder) SkipSize() {
	d.SkipVarUint()
}

// SkipVarInt skips a VarInt value.
func (d *Decoder) SkipVarInt() {
	d.SkipVarUint()
}

// SkipFloat skips a float value.
func (d *Decoder) SkipFloat() {
	d.SkipVarUint()
}

// SkipComplex skips a complex value.
func (d *Decoder) SkipComplex() {
	d.SkipVarUint()
	d.SkipVarUint()
}

// SkipUint8 skips a uint8 value.
func (d *Decoder) SkipUint8() {
	d.SkipByte()
}

// SkipUint16 skips a uint16 value.
func (d *Decoder) SkipUint16() {
	d.SkipBytes(2)
}

// SkipUint32 skips a uint32 value.
func (d *Decoder) SkipUint32() {
	d.SkipBytes(4)
}

// SkipUint64 skips a uint64 value.
func (d *Decoder) SkipUint64() {
	d.SkipBytes(8)
}

// SkipInt8 skips a int8 value.
func (d *Decoder) SkipInt8() {
	d.SkipUint8()
}

// SkipInt16 skips a int16 value.
func (d *Decoder) SkipInt16() {
	d.SkipUint16()
}

// SkipInt32 skips a int32 value.
func (d *Decoder) SkipInt32() {
	d.SkipUint32()

}

// SkipInt64 skips a int64 value.
func (d *Decoder) SkipInt64() {
	d.SkipUint64()
}

// SkipFloat32 skips a float32 value.
func (d *Decoder) SkipFloat32() {
	d.SkipUint32()
}

// SkipFloat64 skips a float64 value.
func (d *Decoder) SkipFloat64() {
	d.SkipUint64()
}

// SkipComplex64 skips a complex64 value.
func (d *Decoder) SkipComplex64() {
	d.SkipUint64()
}

// SkipComplex128 skips a complex128 value.
func (d *Decoder) SkipComplex128() {
	d.SkipBytes(16)
}

// SkipBlob skips a blob value.
func (d *Decoder) SkipBlob(max uint64) {
	n := d.VarUint()
	if n > max {
		panic("IDR decoder: data too big")
	}
	d.SkipBytes(n)
}

// SkipString skips a string value.
func (d *Decoder) SkipString(max uint64) {
	d.SkipBlob(max)
}

// SkipDIR skips a DIR value.
func (d *Decoder) SkipDIR() {
	d.SkipBlob(dir.MaxBinaryLen)
}

// SkipVarTime skips a time value.
func (d *Decoder) SkipVarTime() {
	d.SkipBlob(24)
}

// SkipTime skips a time value.
func (d *Decoder) SkipTime() {
	d.SkipBytes(16)
}
