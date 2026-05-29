package low

import (
	"encoding/binary"
	"math"
	"math/bits"
	"time"

	"github.com/chmike/ditp/dir"
)

// Encoder is a low level IDR encoder.
type Encoder []byte

// Reset clears the encoding buffer.
func Reset(e Encoder) Encoder {
	if len(e) == 0 {
		return e
	}
	return e[:0]
}

// AppendByte appends the byte v.
func AppendByte(e Encoder, v byte) Encoder {
	return append(e, v)
}

// AppendBool appends the bool b.
func AppendBool(e Encoder, v bool) Encoder {
	if v {
		return append(e, 1)
	}
	return append(e, 0)
}

// AppendBytes appends the byte v.
func AppendBytes(e Encoder, v ...byte) Encoder {
	return append(e, v...)
}

// AppendVarUint64 appends the uint64 v using a compact encoding.
func AppendVarUint64(e Encoder, v uint64) Encoder {
	for i := 0; v >= 0x80 && i < 8; i++ {
		e = append(e, byte(v)|0x80)
		v >>= 7
	}
	return append(e, byte(v))
}

// AppendVarUint appends the uint v using a compact encoding.
func AppendVarUint(e Encoder, v uint) Encoder {
	return AppendVarUint64(e, uint64(v))
}

// AppendTag appends the Tag.
func AppendTag(e Encoder, t TagT) Encoder {
	return AppendVarUint64(e, uint64(t))
}

// AppendSize appends v encoded as a size value.
func AppendSize(e Encoder, v uint64) Encoder {
	return AppendVarUint64(e, v)
}

// AppendVarInt64 appends the int64 v using the VarUint encoding.
func AppendVarInt64(e Encoder, v int64) Encoder {
	x := uint64(v) << 1
	if v < 0 {
		x = ^x
	}
	return AppendVarUint64(e, x)
}

// AppendVarInt appends the int v using the VarUint encoding.
func AppendVarInt(e Encoder, v int) Encoder {
	return AppendVarInt64(e, int64(v))
}

// AppendVarFloat appends the float64 using the QVarUint encoding.
func AppendVarFloat(e Encoder, v float64) Encoder {
	return AppendVarUint64(e, bits.ReverseBytes64(math.Float64bits(v)))
}

// AppendVarComplex appends the complex128 using the QVarUint encoding.
func AppendVarComplex(e Encoder, c complex128) Encoder {
	e = AppendVarUint64(e, bits.ReverseBytes64(math.Float64bits(real(c))))
	return AppendVarUint64(e, bits.ReverseBytes64(math.Float64bits(imag(c))))
}

// AppendUint8 appends the uint8 value v.
func AppendUint8(e Encoder, v uint8) Encoder {
	return append(e, v)
}

// AppendUint16 appends the uint16 value v.
func AppendUint16(e Encoder, v uint16) Encoder {
	return binary.LittleEndian.AppendUint16(e, v)
}

// AppendUint32 appends the uint32 value v.
func AppendUint32(e Encoder, v uint32) Encoder {
	return binary.LittleEndian.AppendUint32(e, v)
}

// AppendUint64 appends the uint64 value v.
func AppendUint64(e Encoder, v uint64) Encoder {
	return binary.LittleEndian.AppendUint64(e, v)
}

// AppendInt8 appends the int8 value v.
func AppendInt8(e Encoder, v int8) Encoder {
	return append(e, byte(v))
}

// AppendInt16 appends the int16 value v.
func AppendInt16(e Encoder, v int16) Encoder {
	return binary.LittleEndian.AppendUint16(e, uint16(v))
}

// AppendInt32 appends the int32 value v.
func AppendInt32(e Encoder, v int32) Encoder {
	return binary.LittleEndian.AppendUint32(e, uint32(v))
}

// AppendInt64 appends the int64 value v.
func AppendInt64(e Encoder, v int64) Encoder {
	return binary.LittleEndian.AppendUint64(e, uint64(v))
}

// AppendFloat32 appends the float32 value v.
func AppendFloat32(e Encoder, v float32) Encoder {
	return binary.LittleEndian.AppendUint32(e, math.Float32bits(v))
}

// AppendFloat32 appends the float32 value v.
func AppendFloat64(e Encoder, v float64) Encoder {
	return binary.LittleEndian.AppendUint64(e, math.Float64bits(v))
}

// AppendComplex64 appends the complex64 value v.
func AppendComplex64(e Encoder, v complex64) Encoder {
	e = binary.LittleEndian.AppendUint32(e, math.Float32bits(real(v)))
	return binary.LittleEndian.AppendUint32(e, math.Float32bits(imag(v)))
}

// AppendComplex128 appends the complex128 value v.
func AppendComplex128(e Encoder, v complex128) Encoder {
	e = binary.LittleEndian.AppendUint64(e, math.Float64bits(real(v)))
	return binary.LittleEndian.AppendUint64(e, math.Float64bits(imag(v)))
}

// AppendBlob appends the byte slice b prefixed with its size encoded as QVarUint.
func AppendBlob(e Encoder, b []byte) Encoder {
	return append(AppendVarUint64(e, uint64(len(b))), b...)
}

// AppendString appends the string s prefixed with its size encoded as QVarUint.
func AppendString(e Encoder, s string) Encoder {
	return append(AppendVarUint64(e, uint64(len(s))), s...)
}

// AppendDIR appends the DIR d truncated to at most 7 uint64 encoded as VarUint.
func AppendDIR(e Encoder, d dir.DIR) Encoder {
	p := len(e)
	e = append(e, 0)
	e = d.AppendBinary(e)
	e[p] = byte(len(e) - p - 1)
	return e
}

// AppendVarTime appends the time t in the most compact form.
func AppendVarTime(e Encoder, t time.Time) Encoder {
	utcsec := t.Unix()
	nano := uint64(t.Nanosecond())
	p := len(e)
	e = append(e, 0)
	e = AppendVarInt64(e, utcsec)
	e = AppendVarUint64(e, nano)
	if t.Location() != time.UTC {
		_, offset := t.Zone()
		e = AppendVarInt(e, offset)
	}
	e[p] = byte(len(e) - p - 1)
	return e
}

// AppendTime appends the time t.
func AppendTime(e Encoder, t time.Time) Encoder {
	e = AppendInt64(e, t.Unix())
	e = AppendUint32(e, uint32(t.Nanosecond()))
	if t.Location() == time.UTC {
		return AppendInt32(e, 0)
	}
	_, offset := t.Zone()
	return AppendInt32(e, int32(offset))
}

// size methods

// SizeByte returns the size of a byte value.
func SizeByte() int {
	return 1
}

// SizeBool returns the size of a bool value.
func SizeBool() int {
	return 1
}

// SizeVarUint64 returns the size of a compact encoded uint64 value.
func SizeVarUint64(v uint64) int {
	if v < 0x80 {
		return 1
	}
	return (bits.Len64((v<<1)>>8) + 13) / 7
}

// SizeVarUint returns the size of a compact encoded uint value.
func SizeVarUint(v uint) int {
	return SizeVarUint64(uint64(v))
}

// SizeSize returns the size of a Size value.
func SizeSize(v int) int {
	return SizeVarUint64(uint64(v))
}

// SizeVarInt64 returns the size of a compact encoded int64 value.
func SizeVarInt64(v int64) int {
	return SizeVarUint64(uint64(v << 1))
}

// SizeVarInt returns the size of a compact encoded int value.
func SizeVarInt(v int) int {
	return SizeVarUint(uint(v))
}

// SizeVarFloat returns the size of a compact encoded float value.
func SizeVarFloat(f float64) int {
	return SizeVarUint64(bits.ReverseBytes64(math.Float64bits(f)))
}

// SizeVarComplex returns the size of a compact encoded complex value.
func SizeVarComplex(v complex128) int {
	return SizeVarFloat(real(v)) + SizeVarFloat(imag(v))
}

// SizeUint8 returns the size of a uint8 value.
func SizeUint8() int {
	return 1
}

// SizeUint16 returns the size of a uint16 value.
func SizeUint16() int {
	return 2
}

// SizeUint32 returns the size of a uint32 value.
func SizeUint32() int {
	return 4
}

// SizeUint64 returns the size of a uint64 value.
func SizeUint64() int {
	return 8
}

// SizeInt8 returns the size of a int8 value.
func SizeInt8() int {
	return 1
}

// SizeInt16 returns the size of a int16 value.
func SizeInt16() int {
	return 2
}

// SizeInt32 returns the size of a int32 value.
func SizeInt32() int {
	return 4

}

// SizeInt64 returns the size of a int64 value.
func SizeInt64() int {
	return 8
}

// SizeFloat32 returns the size of a float32 value.
func SizeFloat32() int {
	return 4
}

// SizeFloat64 returns the size of a float64 value.
func SizeFloat64() int {
	return 8
}

// SizeComplex64 returns the size of a complex64 value.
func SizeComplex64() int {
	return 8
}

// SizeComplex128 returns the size of a complex128 value.
func SizeComplex128() int {
	return 16
}

// SizeBlob returns the size of a blob value.
func SizeBlob(b []byte) int {
	v := len(b)
	return SizeSize(v) + v
}

// SizeString returns the size of a string value.
func SizeString(s string) int {
	v := len(s)
	return SizeSize(v) + v
}

// SizeDIR returns the size of a DIR value.
func SizeDIR(d dir.DIR) int {
	return d.BinarySize() + 1
}

// SizeVarTime returns the size of a time value.
func SizeVarTime(t time.Time) int {
	l := 1 + SizeVarInt(t.Nanosecond()) + SizeVarUint64(uint64(t.Nanosecond()))
	if t.Location() != time.UTC {
		_, offset := t.Zone()
		l += SizeVarInt(offset)
	}
	return l
}

// SizeTime returns the size of a time value.
func SizeTime() int {
	return 16
}
