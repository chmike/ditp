package low

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type SmallStruct struct {
	Name     string
	BirthDay time.Time
	Phone    string
	Siblings int
	Spouse   bool
	Money    float64
}

var a = SmallStruct{
	Name:     "benchmark",
	BirthDay: time.Now(),
	Phone:    "709-345678",
	Siblings: 3,
	Spouse:   true,
	Money:    10000,
}

func randString(l int) string {
	buf := make([]byte, l)
	for i := 0; i < (l+1)/2; i++ {
		buf[i] = byte(rand.Intn(256))
	}
	return fmt.Sprintf("%x", buf)[:l]
}

const MaxSmallStructNameSize = 16
const MaxSmallStructPhoneSize = 10

func generateSmallStruct() []*SmallStruct {
	a := make([]*SmallStruct, 0, 1000)
	for i := 0; i < 1000; i++ {
		a = append(a, &SmallStruct{
			Name:     randString(MaxSmallStructNameSize),
			BirthDay: time.Now(),
			Phone:    randString(MaxSmallStructPhoneSize),
			Siblings: rand.Intn(5),
			Spouse:   rand.Intn(2) == 1,
			Money:    rand.Float64(),
		})
	}
	return a
}

func encodeEx(o any) Encoder {
	a := o.(*SmallStruct)
	e := make([]byte, 0, 64)
	e = AppendString(e, a.Name)
	// e = AppendTime(e, a.BirthDay)
	e = AppendInt64(e, a.BirthDay.UnixMicro())
	e = AppendString(e, a.Phone)
	e = AppendVarInt(e, a.Siblings)
	e = AppendBool(e, a.Spouse)
	e = AppendFloat64(e, a.Money)
	return e
}

func decodeEx(d Decoder, a *SmallStruct) Decoder {
	d, a.Name = String(d, 255)
	//d, a.BirthDay = Time(d)
	d, tmp1 := Int64(d)
	a.BirthDay = time.UnixMicro(tmp1)
	d, a.Phone = String(d, 255)
	d, a.Siblings = VarInt(d)
	d, a.Spouse = Bool(d)
	d, a.Money = Float64(d)
	return d
}

var e Encoder
var d Decoder

func BenchmarkEncode(b *testing.B) {
	data := generateSmallStruct()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		e = encodeEx(data[rand.Intn(len(data))])
	}
}

var a2 SmallStruct

func BenchmarkDecode(b *testing.B) {
	// Pre-encode 1000 entries
	src := generateSmallStruct()
	encoded := make([][]byte, len(src))
	for i, v := range src {
		encoded[i] = encodeEx(v)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		data := encoded[rand.Intn(len(encoded))]
		d = Decoder(data)
		d = decodeEx(d, &a2)
	}
}

// func BenchmarkDecode(b *testing.B) {
// 	e = encodeEx(&a)
// 	data := []byte(e)
// 	b.ResetTimer()
// 	for n := 0; n < b.N; n++ {
// 		d = Decoder(data)
// 		d = decodeEx(d, &a2)
// 	}
// }
