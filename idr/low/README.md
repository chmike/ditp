# Low level IDR encoder and decoder

The Information Data Representation (IDR) is a value encoding convention.

The low level encoding package supports basic type values
encoding/decoding methods. The decoding methods panics in case
invalid or truncated encoding is met.

A type tag enum is also provided but intended to be used with
values of type `any`. Composed types require a special type
encoding which is not covered by this package as it only
supports basic type encoding.

## Tags

A Type tag is a uint64 value identifying the type of a value. It is
encoded as a `VarUint` which is a varying length encoding. The
value `InvalidTag` identifies an invalid tag value. The special tag
`NoneTag` is not a type identifier. It may be used as an end of
value sequence marker.

The tags `NoneTag` to `MaxTag` form the set of predefined types marker.

## Encoder

An encoder encodes various types of values in IDR into a buffer.
The default `Encoder` value is a valid encoder initialized with a nil
buffer. The method `With` allows to set a buffer to use for
encoding values. When the buffer is `nil`, the buffer is instantiated
and grown as needed by the `append` function.

The method `Reset` clears the encoding buffer. The method `Bytes`
returns the slice of the encoded values without copy. The `Len` method
returns the current length of the encoded data.

The methods `PutXXX` encode the corresponding type. The Methods `PutVarXXX`
does the same but use a slightly more compact encoding which is also a bit
slower to encode and decode. It uses a `LEB-128` variant where the 8 most
significant bits are encoded in the last byte so that the encoding length
can't exceed 9 bytes.

The method `PutTime` is a fast time encoding with nano second precision and
time offset. The method `PutVarTime` uses a slightly more compact but slower
encoding of the time. The time zone abbreviation is not included since Go
doesn't provide a mean the check its validity.

## Decoder

A decoder decodes various types of IDR encoded values from a given byte
slice. The function `Decode` returns a `Decoder` initialized to decode
the given byte slice. The method `Decode` sets the Decoder to decode
the given slice. The method `Len` returns the number of bytes left to
decode. The method `Peek` returns the slice of bytes left to decode
without making a copy.

The decoder can decode all type of values that can be encoded. It also
has `SkipXXX` methods to skip the corresponding value. It is needed in
case a user needs to ignore the subsequent value.

The decoder does minimal value validity checking as it is context dependent.
The `Blob` and `String` decoding or skipping methods require a maximum size
value to check that the size is not bogus. It doesn't check that the string
contains valid UTF-8 data. The DIR value decoder does panic if the encoding
is invalid.
