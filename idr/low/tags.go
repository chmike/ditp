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
	VarFloatTag
	VarComplexTag
	VarTimeTag
	MaxTag
	InvalidTag = ^TagT(0)
)

func (t TagT) String() string {
	const str = "NoneTag" + // 7 7
		"BoolTag" + // 7 14
		"ByteTag" + // 7 21
		"Uint8Tag" + // 8 29
		"Uint16Tag" + // 9 38
		"Uint32Tag" + // 9 47
		"Uint64Tag" + // 9 56
		"Int8Tag" + // 7 63
		"Int16Tag" + // 8 71
		"Int32Tag" + // 8 79
		"Int64Tag" + // 8 87
		"Float32Tag" + // 10 97
		"Float64Tag" + // 10 107
		"Complex64Tag" + // 12 119
		"Complex128Tag" + // 13 132
		"TimeTag" + // 7 139
		"BytesTag" + // 8 147
		"SizeTag" + // 7 154
		"BlobTag" + // 7 161
		"StringTag" + // 9 170
		"DIRTag" + // 6 176
		"VarUintTag" + // 10 186
		"VarIntTag" + // 9 195
		"VarFloatTag" + // 11 206
		"VarComplexTag" + // 13 219
		"VarTimeTag" + // 10 229
		"InvalidTag" // 10 239
	var idx = []byte{0, 7, 14, 21, 29, 38, 47, 56, 63, 71, 79, 87, 97, 107, 119, 132, 139, 147, 154, 161, 170, 176, 186, 195, 206, 219, 229, 239}
	if t < MaxTag {
		return str[idx[t]:idx[t+1]]
	}
	if t == InvalidTag {
		return str[idx[MaxTag]:idx[MaxTag+1]]
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
		"VarFloatTag":   23,
		"VarComplexTag": 24,
		"VarTimeTag":    25,
		"MaxTag":        26,
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
