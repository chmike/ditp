package low

import (
	"encoding/binary"
	"math"
	"math/bits"
	"time"

	"github.com/chmike/ditp/dir"
)

// Encoder is a low level IDR encoder.
type Encoder struct {
	b []byte
}

// With sets the encoder buffer to use. The encoder will append to b.
func (e *Encoder) With(b []byte) {
	e.b = b
}

// Reset clears the encoding buffer.
func (e *Encoder) Reset() {
	if len(e.b) != 0 {
		e.b = e.b[:0]
	}
}

// Len returns the byte length of the serialized data.
func (e Encoder) Len() int {
	return len(e.b)
}

// Bytes returns the serialized Bytes without copy.
func (e Encoder) Bytes() []byte {
	return e.b
}

// PutByte appends the byte v.
func (e *Encoder) PutByte(v byte) {
	e.b = append(e.b, v)
}

// PutBool appends the bool b.
func (e *Encoder) PutBool(v bool) {
	var t byte
	if v {
		t = 1
	}
	e.PutByte(t)
}

// PutBytes appends the byte v.
func (e *Encoder) PutBytes(v ...byte) {
	e.b = append(e.b, v...)
}

// PutVarUint appends the uint64 v using a compact encoding.
func (e *Encoder) PutVarUint(v uint64) {
	for i := 0; v >= 0x80 && i < 8; i++ {
		e.b = append(e.b, byte(v)|0x80)
		v >>= 7
	}
	e.b = append(e.b, byte(v))
}

// PutTag appends the Tag.
func (e *Encoder) PutTag(t Tag) {
	e.PutVarUint(uint64(t))
}

// PutSize appends v encoded as a size value.
func (e *Encoder) PutSize(v uint64) {
	e.PutVarUint(v)
}

// PutVarInt appends the int64 v using the VarUint encoding.
func (e *Encoder) PutVarInt(v int64) {
	x := uint64(v) << 1
	if v < 0 {
		x = ^x
	}
	e.PutVarUint(x)
}

// PutVarFloat appends the float64 using the QVarUint encoding.
func (e *Encoder) PutVarFloat(v float64) {
	e.PutVarUint(bits.ReverseBytes64(math.Float64bits(v)))
}

// PutVarComplex appends the complex128 using the QVarUint encoding.
func (e *Encoder) PutVarComplex(c complex128) {
	e.PutVarUint(bits.ReverseBytes64(math.Float64bits(real(c))))
	e.PutVarUint(bits.ReverseBytes64(math.Float64bits(imag(c))))
}

// PutUint8 appends the uint8 value v.
func (e *Encoder) PutUint8(v uint8) {
	e.b = append(e.b, v)
}

// PutUint16 appends the uint16 value v.
func (e *Encoder) PutUint16(v uint16) {
	e.b = binary.LittleEndian.AppendUint16(e.b, v)
}

// PutUint32 appends the uint32 value v.
func (e *Encoder) PutUint32(v uint32) {
	e.b = binary.LittleEndian.AppendUint32(e.b, v)
}

// PutUint64 appends the uint64 value v.
func (e *Encoder) PutUint64(v uint64) {
	e.b = binary.LittleEndian.AppendUint64(e.b, v)
}

// PutInt8 appends the int8 value v.
func (e *Encoder) PutInt8(v int8) {
	e.b = append(e.b, byte(v))
}

// PutInt16 appends the int16 value v.
func (e *Encoder) PutInt16(v int16) {
	e.b = binary.LittleEndian.AppendUint16(e.b, uint16(v))
}

// PutInt32 appends the int32 value v.
func (e *Encoder) PutInt32(v int32) {
	e.b = binary.LittleEndian.AppendUint32(e.b, uint32(v))
}

// PutInt64 appends the int64 value v.
func (e *Encoder) PutInt64(v int64) {
	e.b = binary.LittleEndian.AppendUint64(e.b, uint64(v))
}

// PutFloat32 appends the float32 value v.
func (e *Encoder) PutFloat32(v float32) {
	e.b = binary.LittleEndian.AppendUint32(e.b, math.Float32bits(v))
}

// PutFloat32 appends the float32 value v.
func (e *Encoder) PutFloat64(v float64) {
	e.b = binary.LittleEndian.AppendUint64(e.b, math.Float64bits(v))
}

// PutComplex64 appends the complex64 value v.
func (e *Encoder) PutComplex64(v complex64) {
	e.b = binary.LittleEndian.AppendUint32(e.b, math.Float32bits(real(v)))
	e.b = binary.LittleEndian.AppendUint32(e.b, math.Float32bits(imag(v)))
}

// PutComplex128 appends the complex128 value v.
func (e *Encoder) PutComplex128(v complex128) {
	e.b = binary.LittleEndian.AppendUint64(e.b, math.Float64bits(real(v)))
	e.b = binary.LittleEndian.AppendUint64(e.b, math.Float64bits(imag(v)))
}

// PutBlob appends the byte slice b prefixed with its size encoded as QVarUint.
func (e *Encoder) PutBlob(b []byte) {
	e.PutVarUint(uint64(len(b)))
	e.b = append(e.b, b...)
}

// PutString appends the string s prefixed with its size encoded as QVarUint.
func (e *Encoder) PutString(s string) {
	e.PutVarUint(uint64(len(s)))
	e.b = append(e.b, s...)
}

// PutDIR appends the DIR d truncated to at most 7 uint64 encoded as VarUint.
func (e *Encoder) PutDIR(d dir.DIR) {
	p := len(e.b)
	e.b = append(e.b, 0)
	e.b = d.AppendBinary(e.b)
	e.b[p] = byte(len(e.b) - p - 1)
}

// PutVarTime appends the time t.
func (e *Encoder) PutVarTime(t time.Time) {
	nano := t.Nanosecond()
	zName, offset := t.Zone()
	utcsec := t.Unix()
	p := len(e.b)
	e.b = append(e.b, 0)
	e.PutVarInt(utcsec)
	e.PutUint32(uint32(nano))
	e.PutVarInt(int64(offset))
	if offset != 0 {
		e.PutString(zName)
	}
	e.b[p] = byte(len(e.b) - p - 1)
}

// PutTime appends the time t.
func (e *Encoder) PutTime(t time.Time) {
	e.PutInt64(t.Unix())
	e.PutUint32(uint32(t.Nanosecond()))
	_, offset := t.Zone()
	e.PutInt32(int32(offset))
}
