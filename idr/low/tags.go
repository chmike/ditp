package low

import "fmt"

type TagT uint64

const (
	NoneTag TagT = iota
	BoolTag
	ByteTag
	Uint8Tag
	Uint16Tag
	Uint32Tag
	Uint64Tag
	Int8Tag
	Int16Tag
	Int32Tag
	Int64Tag
	Float32Tag
	Float64Tag
	Complex64Tag
	Complex128Tag
	TimeTag
	BytesTag
	SizeTag
	BlobTag
	StringTag
	DIRTag
	VarUintTag
	VarIntTag
	VarUint64Tag
	VarInt64Tag
	VarFloatTag
	VarComplexTag
	VarTimeTag
	MaxTag
	InvalidTag = ^TagT(0)
)

func (t TagT) String() string {
	var str = []string{
		"NoneTag",
		"BoolTag",
		"ByteTag",
		"Uint8Tag",
		"Uint16Tag",
		"Uint32Tag",
		"Uint64Tag",
		"Int8Tag",
		"Int16Tag",
		"Int32Tag",
		"Int64Tag",
		"Float32Tag",
		"Float64Tag",
		"Complex64Tag",
		"Complex128Tag",
		"TimeTag",
		"BytesTag",
		"SizeTag",
		"BlobTag",
		"StringTag",
		"DIRTag",
		"VarUintTag",
		"VarIntTag",
		"VarUint64Tag",
		"VarInt64Tag",
		"VarFloatTag",
		"VarComplexTag",
		"VarTimeTag",
		"InvalidTag",
	}
	if t < MaxTag {
		return str[t]
	}
	if t == InvalidTag {
		return str[len(str)-1]
	}
	return fmt.Sprintf("(%d)Tag", t)
}

// TagFromString returns the TagT given the string.
func TagFromString(s string) TagT {
	var m = map[string]TagT{
		"NoneTag":       0,
		"BoolTag":       1,
		"ByteTag":       2,
		"Uint8Tag":      3,
		"Uint16Tag":     4,
		"Uint32Tag":     5,
		"Uint64Tag":     6,
		"Int8Tag":       7,
		"Int16Tag":      8,
		"Int32Tag":      9,
		"Int64Tag":      10,
		"Float32Tag":    11,
		"Float64Tag":    12,
		"Complex64Tag":  13,
		"Complex128Tag": 14,
		"TimeTag":       15,
		"BytesTag":      16,
		"SizeTag":       17,
		"BlobTag":       18,
		"StringTag":     19,
		"DIRTag":        20,
		"VarUintTag":    21,
		"VarIntTag":     22,
		"VarUint64Tag":  23,
		"VarInt64Tag":   24,
		"VarFloatTag":   25,
		"VarComplexTag": 26,
		"VarTimeTag":    27,
		"InvalidTag":    ^TagT(0),
	}
	if t, OK := m[s]; OK {
		return t
	}
	var v TagT
	n, err := fmt.Sscanf(s, "(%d)Tag", &v)
	if n != 1 || err != nil {
		return InvalidTag
	}
	return v
}
