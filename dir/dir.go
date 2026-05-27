package dir

import (
	"errors"
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

var ErrInvalid = errors.New("invalid dir")

// DIR is a downward path in a distributed information system (DIS) tree.
// d[0] is the number of identifiers. d[1..7] are the identifiers.
type DIR [8]uint64

// MaxIDs is the maximum number of uint64 identifiers in a DIR.
const MaxIDs = 7

// MaxBinaryLen is the maximum byte length of a binary encoded DIR.
const MaxBinaryLen = MaxIDs * maxBinaryIDLen // = 63

// maxBinaryIDLen is the maximum byte length of a binary encoded identifier.
const maxBinaryIDLen = 9

// MaxURILen is the maximum byte length of an ASCII encoded DIR.
const MaxURILen = (maxURIIDLen+1)*MaxIDs + 4 // = 88

// maxURIIDLen is the maximum byte length of an ASCII encoded ID.
const maxURIIDLen = 11 // = (64 + 2)/6

// Make returns a DIR with the given identifiers or an error when invalid.
// A DIR defines a path in a tree like graph from the root toward the leafs.
// A DIR may have at most MaxDIRLen identifiers. Only the first and last
// may be zero. When the last identifier is 0, the DIR is a node DIR. When
// there are at least two identifiers and the first is 0, the DIR is a
// relative DIR and the path doesn't start at the root of the tree.
func Make(ids ...uint64) (DIR, error) {
	if len(ids) > MaxIDs {
		return DIR{}, fmt.Errorf("%w: too many identifiers", ErrInvalid)
	}
	for i := 1; i < len(ids)-1; i++ {
		if ids[i] == 0 {
			return DIR{}, fmt.Errorf("%w: identifier %d is 0", ErrInvalid, i)
		}
	}
	var d DIR
	d[0] = uint64(copy(d[1:], ids))
	return d, nil
}

// MustMake calls Make and returns the DIR or panics in case of error.
func MustMake(ids ...uint64) DIR {
	d, err := Make(ids...)
	if err != nil {
		panic(err)
	}
	return d
}

// Check returns an error if d is an invalid DIR.
func (d DIR) Check() error {
	if d[0] > MaxIDs {
		return fmt.Errorf("%w: too many identifiers", ErrInvalid)
	}
	for i := uint64(2); i < d[0]; i++ {
		if d[i] == 0 {
			return fmt.Errorf("%w: identifier %d is 0", ErrInvalid, i-1)
		}
	}
	return nil
}

// String converts d to a human readable representation.
func (d DIR) String() string {
	var b strings.Builder
	b.WriteString("dir:")
	for i := 0; i < d.Len(); i++ {
		if i != 0 {
			b.WriteByte('.')
		}
		b.WriteString(strconv.FormatUint(d[i+1], 10))
	}
	return b.String()
}

// Len returns the number of identifiers in d.
func (d DIR) Len() int {
	return int(d[0])
}

// ID returns the identifier at index i, or 0 if out of range.
func (d DIR) ID(i int) uint64 {
	if i >= 0 && i < d.Len() {
		return d[i+1]
	}
	return 0
}

// IDs copies the identifiers of d into ids and returns the result.
func (d DIR) IDs(ids []uint64) []uint64 {
	return append(ids[:0], d[1:d.Len()+1]...)
}

// Copy returns a copy of d. Since DIR is a value type, this is equivalent
// to a plain assignment, but provided for API compatibility.
func (d DIR) Copy() DIR {
	return d
}

// InfoID returns the information ID or 0 when d is nil.
func (d DIR) InfoID() uint64 {
	return d[d.Len()]
}

// Equal returns true if d is equal to d2.
func (d DIR) Equal(d2 DIR) bool {
	return d == d2
}

// Prefixes returns true if d is a node DIR and d2 is prefixed with d.
func (d DIR) Prefixes(d2 DIR) bool {
	if d.IsInfo() || d.Len()-1 >= d2.Len() {
		return false
	}
	for i := 0; i < d.Len()-1; i++ {
		if d[i+1] != d2[i+1] {
			return false
		}
	}
	return d.Len() != d2.Len() || d2.InfoID() != 0
}

// IsNil returns true if d is a nil DIR. A nil DIR has no identifiers.
func (d DIR) IsNil() bool {
	return d[0] == 0
}

// SetNil returns a nil DIR.
func (d *DIR) SetNil() {
	*d = DIR{}
}

// IsNode returns true if d is a node DIR.
func (d DIR) IsNode() bool {
	return d.Len() > 0 && d.InfoID() == 0
}

// IsInfo returns true if d is an absolute path to an information or nil.
func (d DIR) IsInfo() bool {
	return d.Len() == 0 || d.InfoID() != 0
}

// IsRelative returns true if d is a relative DIR.
func (d DIR) IsRelative() bool {
	return d.Len() > 1 && d[1] == 0
}

// IsAbsolute returns true if d is an absolute DIR or nil.
func (d DIR) IsAbsolute() bool {
	return d.Len() <= 1 || d[1] != 0
}

// AppendBinary appends d binary encoded to b. The first byte encodes
// the number of bytes that follow. They contain the identifiers encoded
// using a LEB-128 variant encoding where the 8 most significant bits are
// encoded in a single byte when not zero.
func (d DIR) AppendBinary(b []byte) []byte {
	n := d.Len()
	if n == 0 {
		return b
	}
	if b == nil {
		b = make([]byte, 0, MaxBinaryLen)
	}
	for i := 1; i <= n; i++ {
		v := d[i]
		for j := 0; v >= 0x80 && j < 8; j++ {
			b = append(b, byte(v)|0x80)
			v >>= 7
		}
		b = append(b, byte(v))
	}
	return b
}

// BinarySize returns the byte size of the binary encoded DIR.
func (d DIR) BinarySize() int {
	var l int
	for i := uint64(1); i <= d[0]; i++ {
		v := d[i]
		if v >= 0x80 {
			l += (bits.Len64((v<<1)>>8) + 6) / 7
		}
		l++
	}
	return l
}

// EncodeBinary encodes the DIR d in binary at offset n in b and returns the
// offset of the first byte after the binary encoded DIR. Panics if b is
// not big enough.
func (d DIR) EncodeBinary(b []byte, n int) int {
	if l := d.Len(); l > 0 {
		for i := 1; i <= l; i++ {
			v := d[i]
			for j := 0; v >= 0x80 && j < 8; j++ {
				b[n] = byte(v) | 0x80
				n++
				v >>= 7
			}
			b[n] = byte(v)
			n++
		}
	}
	return n
}

// DecodeBinary decodes the binary encoded DIR in b.
func DecodeBinary(b []byte) (DIR, error) {
	var d DIR
	if len(b) == 0 {
		return DIR{}, nil
	}
	var n uint64
	var i int
	for n < MaxIDs && i < len(b) {
		var v uint64
		var s byte
		limit := min(len(b), i+8)
		for i < limit && b[i] >= 0x80 {
			v |= uint64(b[i]&0x7f) << s
			s += 7
			i++
		}
		if i == len(b) {
			return DIR{}, fmt.Errorf("%w: truncated binary identifier", ErrInvalid)
		}
		n++
		d[n] = v | (uint64(b[i]) << s)
		i++
	}
	if i < len(b) {
		return DIR{}, fmt.Errorf("%w: too many identifiers", ErrInvalid)
	}
	d[0] = uint64(n)
	for i := 2; i < int(n); i++ {
		if d[i] == 0 {
			return DIR{}, fmt.Errorf("%w: identifier %d is 0", ErrInvalid, i-1)
		}
	}
	return d, nil
}

// URI returns d encoded as a URI string.
func (d DIR) URI() string {
	return string(d.AppendURI(make([]byte, 0, (maxURIIDLen+1)*d.Len()+4)))
}

// AppendURI appends d URI encoded to b.
func (d DIR) AppendURI(b []byte) []byte {
	const txtChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
	b = append(b, "dis:"...)
	n := d.Len()
	if n > 0 {
		for i := 1; i <= n; i++ {
			id := d[i]
			if id == 0 {
				b = append(b, '0')
			} else {
				for id != 0 {
					b = append(b, txtChars[id&0x3F])
					id >>= 6
				}
			}
			b = append(b, '.')
		}
		b = b[:len(b)-1]
	}
	return append(b, '/')
}

// DecodeURI decodes a URI encoded DIR from b.
func DecodeURI[T string | []byte](b T) (DIR, error) {
	var txtTbl = [256]int8{
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 62, -1, -1,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, -1, -1, -1, -1, -1, -1,
		-1, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
		25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, -1, -1, -1, -1, 63,
		-1, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50,
		51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	}
	var d DIR
	if len(b) < 5 || b[0] != 'd' || b[1] != 'i' || b[2] != 's' || b[3] != ':' || b[len(b)-1] != '/' {
		return d, fmt.Errorf("%w: URI must start with \"dis:\" and end with \"/\"", ErrInvalid)
	}
	t := b[4:]
	n := 0
	for n < MaxIDs && t[0] != '/' {
		var v uint64
		var s byte
		var i int
		for i < maxURIIDLen {
			c := txtTbl[t[i]]
			if c < 0 {
				break
			}
			v |= uint64(c) << s
			s += 6
			i++
		}
		if i == maxURIIDLen && txtTbl[t[i-1]] > 0xF {
			return DIR{}, fmt.Errorf("%w: identifier overflow", ErrInvalid)
		}
		n++
		d[n] = v
		t = t[i:]
		if n == MaxIDs || t[0] != '.' {
			break
		}
		t = t[1:]
	}
	if t[0] != '/' {
		return DIR{}, fmt.Errorf("%w: invalid characters in URI", ErrInvalid)
	}
	if len(t) > 1 {
		return DIR{}, fmt.Errorf("%w: character '/' in URI", ErrInvalid)
	}
	d[0] = uint64(n)
	for i := 2; i < n; i++ {
		if d[i] == 0 {
			return DIR{}, fmt.Errorf("%w: identifier %d is 0", ErrInvalid, i-1)
		}
	}
	return d, nil
}
