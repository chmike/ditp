package dir

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalid = errors.New("invalid dir")

// DIR is a downward path in a distributed information system (DIS) tree.
type DIR struct {
	d []uint64
}

// MaxDIRLen is the maximum number of uint64 identifiers in a DIR.
const MaxDIRLen = 7

// MaxBinaryLen is the maximum byte length of a binary encoded DIR.
const MaxBinaryLen = MaxDIRLen * maxBinaryIDLen // = 63

// maxBinaryIDLen is the maximum byte length of a binary encoded identifier.
const maxBinaryIDLen = 9

// MaxURILen is the maximum byte length of an ASCII encoded DIR.
const MaxURILen = (maxURIIDLen+1)*MaxDIRLen + 4 // = 88

// maxURIIDLen is the maximum byte length of an ASCII encoded ID.
const maxURIIDLen = 11 // = (64 + 2)/6

// Make returns a DIR with the given identifiers or an error when invalid.
// A DIR defines a path in a tree like graph from the root toward the leafs.
// A DIR may have at most MaxDIRLen identifiers. Only the first and last
// may be zero. When the last identifier is 0, the DIR is a node DIR. When
// there are at least two identifiers and the first is 0, the DIR is a
// relative DIR and the path doesn't start at the root of the tree.
func Make(ids ...uint64) (DIR, error) {
	if len(ids) > MaxDIRLen {
		return DIR{}, fmt.Errorf("%w: too many identifiers", ErrInvalid)
	}
	for i := 1; i < len(ids)-1; i++ {
		if ids[i] == 0 {
			return DIR{}, fmt.Errorf("%w: identifier %d is 0", ErrInvalid, i)
		}
	}
	return DIR{append(make([]uint64, 0, MaxDIRLen), ids...)}, nil
}

// MustMake calls Make and return the DIR or panics in case of error.
func MustMake(ids ...uint64) DIR {
	d, err := Make(ids...)
	if err != nil {
		panic(err)
	}
	return d
}

// String converts d to a human readable representation.
func (d DIR) String() string {
	var b strings.Builder
	b.WriteString("dir:")
	for i := 0; i < d.Len(); i++ {
		if i != 0 {
			b.WriteByte('.')
		}
		b.WriteString(strconv.FormatUint(d.d[i], 10))
	}
	return b.String()
}

// Len returns the number of identifiers in d.
func (d DIR) Len() int {
	return len(d.d)
}

// ID returns the ID at index i or 0 if out of range.
func (d DIR) ID(i int) uint64 {
	if i > 0 && i < d.Len() {
		return d.d[i]
	}
	return 0
}

// IDs copies the identifiers of d into ids. If ids is nil, or not
// bi enough, returns a new slice.
func (d DIR) IDs(ids []uint64) []uint64 {
	if len(ids) > 0 {
		ids = ids[:0]
	}
	return append(ids, d.d...)
}

// Copy copies the identifiers of d into d2.
func (d DIR) Copy(d2 DIR) DIR {
	d2.SetNil()
	return DIR{append(d2.d, d.d...)}
}

// InfoID returns the information ID or 0 when d is nil.
func (d DIR) InfoID() uint64 {
	if d.Len() == 0 {
		return 0
	}
	return d.d[d.Len()-1]
}

// Equal return true if d is equal to d2.
func (d DIR) Equal(d2 DIR) bool {
	if d.Len() != d2.Len() {
		return false
	}
	for i := 0; i < d.Len(); i++ {
		if d.d[i] != d2.d[i] {
			return false
		}
	}
	return true
}

// Contains returns true if d is a node DIR and d2 is a
// node or information contained in d.
func (d DIR) Contains(d2 DIR) bool {
	if d.IsInfo() || d.Len()-1 >= d2.Len() {
		return false
	}
	for i, id := range d.d[:len(d.d)-1] {
		if d2.d[i] != id {
			return false
		}
	}
	return d.Len() != d2.Len() || d2.InfoID() != 0
}

// IsNil returns true if d is a nil DIR. A nil DIR has no identifiers.
func (d DIR) IsNil() bool {
	return d.Len() == 0
}

// SetNil set d as a nil DIR.
func (d *DIR) SetNil() {
	if d.Len() != 0 {
		d.d = d.d[:0]
	}
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
	return d.Len() > 1 && d.d[0] == 0
}

// IsAbsolute returns true if d is an absolute DIR or nil.
func (d DIR) IsAbsolute() bool {
	return d.Len() <= 1 || d.d[0] != 0
}

// AppendBinary appends d binary encoded to b. The identifiers are encoded using
// a LEB-128 variant encoding where the 8 most significant bits are encoded in a
// single byte when not zero.
func (d DIR) AppendBinary(b []byte) []byte {
	if d.Len() == 0 {
		return b
	}
	if b == nil {
		b = make([]byte, 0, maxBinaryIDLen*d.Len())
	}
	for _, v := range d.d {
		for i := 0; v >= 0x80 && i < 8; i++ {
			b = append(b, byte(v)|0x80)
			v >>= 7
		}
		b = append(b, byte(v))
	}
	return b
}

// DecodeBinary decodes the binary DIR in b using d as storage.
// Pass DIR{} when no storage is available.
func DecodeBinary(d DIR, b []byte) (DIR, error) {
	d.SetNil()
	if len(b) == 0 {
		return d, nil
	}
	if d.d == nil {
		d.d = make([]uint64, 0, MaxDIRLen)
	}
	for j := 0; j < MaxDIRLen && len(b) > 0; j++ {
		var v uint64
		var s byte
		var i int
		n := min(len(b), 8)
		for i < n && b[i] >= 0x80 {
			v |= uint64(b[i]&0x7f) << s
			s += 7
			i++
		}
		if i == len(b) {
			d.SetNil()
			return d, fmt.Errorf("%w: truncated binary identifier", ErrInvalid)
		}
		d.d = append(d.d, v|(uint64(b[i])<<s))
		b = b[i+1:]
	}
	if len(b) > 0 {
		d.SetNil()
		return d, fmt.Errorf("%w: too many identifiers", ErrInvalid)
	}
	for i := 1; i < d.Len()-1; i++ {
		if d.d[i] == 0 {
			d.SetNil()
			return d, fmt.Errorf("%w: identifier %d is 0", ErrInvalid, i)
		}
	}
	return d, nil
}

// URI returns d encoded as a URI string. Requires that d is valid.
func (d DIR) URI() string {
	return string(d.AppendURI(make([]byte, 0, (maxURIIDLen+1)*d.Len()+4)))
}

// AppendURI appends d URI encoded to b.
func (d DIR) AppendURI(b []byte) []byte {
	const txtChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
	b = append(b, "dis:"...)
	if d.Len() > 0 {
		for _, id := range d.d {
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

// DecodeURI decodes the binary DIR in b using d as storage.
// Pass DIR{} when no storage is available.
func DecodeURI[T string | []byte](d DIR, b T) (DIR, error) {
	var txtTbl = []int8{
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
	if d.Len() != 0 {
		d.d = d.d[:0]
	}
	if len(b) < 5 || b[0] != 'd' || b[1] != 'i' || b[2] != 's' || b[3] != ':' || b[len(b)-1] != '/' {
		return d, fmt.Errorf("%w: URI must start with \"dis:\" and end with \"/\"", ErrInvalid)
	}
	t := b[4:]
	for d.Len() < MaxDIRLen && t[0] != '/' {
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
			d.SetNil()
			return d, fmt.Errorf("%w: identifier overflow", ErrInvalid)
		}
		d.d = append(d.d, v)
		t = t[i:]
		if d.Len() == MaxDIRLen || t[0] != '.' {
			break
		}
		t = t[1:]
	}
	if t[0] != '/' {
		d.SetNil()
		return d, fmt.Errorf("%w: invalid characters in URI", ErrInvalid)
	}
	if len(t) > 1 {
		d.SetNil()
		return d, fmt.Errorf("%w: character '/' is URI", ErrInvalid)
	}
	for i := 1; i < d.Len()-1; i++ {
		if d.d[i] == 0 {
			d.SetNil()
			return d, fmt.Errorf("%w: identifier %d is 0", ErrInvalid, i)
		}
	}
	return d, nil
}
