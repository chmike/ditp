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

// PutByte appends the byte v.
func PutByte(e Encoder, v byte) Encoder {
	return append(e, v)
}

// PutBool appends the bool b.
func PutBool(e Encoder, v bool) Encoder {
	if v {
		return append(e, 1)
	}
	return append(e, 0)
}

// PutBytes appends the byte v.
func PutBytes(e Encoder, v ...byte) Encoder {
	return append(e, v...)
}

// PutVarUint appends the uint64 v using a compact encoding.
func PutVarUint(e Encoder, v uint64) Encoder {
	for i := 0; v >= 0x80 && i < 8; i++ {
		e = append(e, byte(v)|0x80)
		v >>= 7
	}
	return append(e, byte(v))
}

// PutTag appends the Tag.
func PutTag(e Encoder, t TagT) Encoder {
	return PutVarUint(e, uint64(t))
}

// PutSize appends v encoded as a size value.
func PutSize(e Encoder, v uint64) Encoder {
	return PutVarUint(e, v)
}

// PutVarInt appends the int64 v using the VarUint encoding.
func PutVarInt(e Encoder, v int64) Encoder {
	x := uint64(v) << 1
	if v < 0 {
		x = ^x
	}
	return PutVarUint(e, x)
}

// PutVarFloat appends the float64 using the QVarUint encoding.
func PutVarFloat(e Encoder, v float64) Encoder {
	return PutVarUint(e, bits.ReverseBytes64(math.Float64bits(v)))
}

// PutVarComplex appends the complex128 using the QVarUint encoding.
func PutVarComplex(e Encoder, c complex128) Encoder {
	e = PutVarUint(e, bits.ReverseBytes64(math.Float64bits(real(c))))
	return PutVarUint(e, bits.ReverseBytes64(math.Float64bits(imag(c))))
}

// PutUint8 appends the uint8 value v.
func PutUint8(e Encoder, v uint8) Encoder {
	return append(e, v)
}

// PutUint16 appends the uint16 value v.
func PutUint16(e Encoder, v uint16) Encoder {
	return binary.LittleEndian.AppendUint16(e, v)
}

// PutUint32 appends the uint32 value v.
func PutUint32(e Encoder, v uint32) Encoder {
	return binary.LittleEndian.AppendUint32(e, v)
}

// PutUint64 appends the uint64 value v.
func PutUint64(e Encoder, v uint64) Encoder {
	return binary.LittleEndian.AppendUint64(e, v)
}

// PutInt8 appends the int8 value v.
func PutInt8(e Encoder, v int8) Encoder {
	return append(e, byte(v))
}

// PutInt16 appends the int16 value v.
func PutInt16(e Encoder, v int16) Encoder {
	return binary.LittleEndian.AppendUint16(e, uint16(v))
}

// PutInt32 appends the int32 value v.
func PutInt32(e Encoder, v int32) Encoder {
	return binary.LittleEndian.AppendUint32(e, uint32(v))
}

// PutInt64 appends the int64 value v.
func PutInt64(e Encoder, v int64) Encoder {
	return binary.LittleEndian.AppendUint64(e, uint64(v))
}

// PutFloat32 appends the float32 value v.
func PutFloat32(e Encoder, v float32) Encoder {
	return binary.LittleEndian.AppendUint32(e, math.Float32bits(v))
}

// PutFloat32 appends the float32 value v.
func PutFloat64(e Encoder, v float64) Encoder {
	return binary.LittleEndian.AppendUint64(e, math.Float64bits(v))
}

// PutComplex64 appends the complex64 value v.
func PutComplex64(e Encoder, v complex64) Encoder {
	e = binary.LittleEndian.AppendUint32(e, math.Float32bits(real(v)))
	return binary.LittleEndian.AppendUint32(e, math.Float32bits(imag(v)))
}

// PutComplex128 appends the complex128 value v.
func PutComplex128(e Encoder, v complex128) Encoder {
	e = binary.LittleEndian.AppendUint64(e, math.Float64bits(real(v)))
	return binary.LittleEndian.AppendUint64(e, math.Float64bits(imag(v)))
}

// PutBlob appends the byte slice b prefixed with its size encoded as QVarUint.
func PutBlob(e Encoder, b []byte) Encoder {
	return append(PutVarUint(e, uint64(len(b))), b...)
}

// PutString appends the string s prefixed with its size encoded as QVarUint.
func PutString(e Encoder, s string) Encoder {
	return append(PutVarUint(e, uint64(len(s))), s...)
}

// PutDIR appends the DIR d truncated to at most 7 uint64 encoded as VarUint.
func PutDIR(e Encoder, d dir.DIR) Encoder {
	p := len(e)
	e = append(e, 0)
	e = d.AppendBinary(e)
	e[p] = byte(len(e) - p - 1)
	return e
}

// PutVarTime appends the time t.
func PutVarTime(e Encoder, t time.Time) Encoder {
	utcsec := t.Unix()
	nano := t.Nanosecond()
	zName, offset := t.Zone()
	p := len(e)
	e = append(e, 0)
	e = PutVarInt(e, utcsec)
	e = PutUint32(e, uint32(nano))
	if zName != "UTC" {
		e = PutVarInt(e, int64(offset))
		e = PutString(e, zName)
	}
	e[p] = byte(len(e) - p - 1)
	return e
}

// PutTime appends the time t.
func PutTime(e Encoder, t time.Time) Encoder {
	e = PutInt64(e, t.Unix())
	e = PutUint32(e, uint32(t.Nanosecond()))
	_, offset := t.Zone()
	return PutInt32(e, int32(offset))
}
