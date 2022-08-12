package sds

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"testing"
	"time"
)

func init() {
	seed := time.Now().UnixNano()
	println(seed)
	rand.Seed(seed)
}

func randUint64() uint64 {
	return rand.Uint64()
}

func randInt64() int64 {
	return int64(randUint64())
}

func randString() string {
	b := make([]byte, rand.Int()%2048)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = (b[i] % 26) + 'a'
	}
	return string(b)
}

type testBytesBuffer struct {
	err error
	buf bytes.Buffer
}

func (bb *testBytesBuffer) Write(p []byte) (int, error) {
	if bb.err != nil {
		return 0, bb.err
	}
	return bb.buf.Write(p)
}

func (bb *testBytesBuffer) Read(p []byte) (int, error) {
	if bb.err != nil {
		return 0, bb.err
	}
	return bb.buf.Read(p)
}

type varint int64
type uvarint uint64
type bbyte byte

func randEl() interface{} {
	switch rand.Int() % 17 {
	case 0:
		return int8(randInt64())
	case 1:
		return int16(randInt64())
	case 2:
		return int32(randInt64())
	case 3:
		return int64(randInt64())
	case 4:
		return uint8(randUint64())
	case 5:
		return uint16(randUint64())
	case 6:
		return uint32(randUint64())
	case 7:
		return uint64(randUint64())
	case 8:
		return rand.Float32()
	case 9:
		return rand.Float64()
	case 10:
		return true
	case 11:
		return false
	case 12:
		return randString()
	case 13:
		return []byte(randString())
	case 14:
		return uvarint(randUint64())
	case 15:
		return varint(randInt64())
	case 16:
		return bbyte(randInt64())
	}
	panic("invalid")
}

func TestSnapBits(t *testing.T) {
	start := time.Now()
	for time.Since(start) < time.Second {
		N := 10_000 // number of random elements
		els := make([]interface{}, N)
		var bb testBytesBuffer
		w := NewWriter(&bb)
		for i := 0; i < len(els); i++ {
			els[i] = randEl()
			switch v := els[i].(type) {
			case int8:
				w.WriteInt8(v)
			case int16:
				w.WriteInt16(v)
			case int32:
				w.WriteInt32(v)
			case int64:
				w.WriteInt64(v)
			case uint8:
				w.WriteUint8(v)
			case uint16:
				w.WriteUint16(v)
			case uint32:
				w.WriteUint32(v)
			case uint64:
				w.WriteUint64(v)
			case float32:
				w.WriteFloat32(v)
			case float64:
				w.WriteFloat64(v)
			case bool:
				w.WriteBool(v)
			case string:
				w.WriteString(v)
			case []byte:
				w.WriteBytes(v)
			case uvarint:
				w.WriteUvarint(uint64(v))
			case varint:
				w.WriteVarint(int64(v))
			case bbyte:
				w.WriteByte(byte(v))
			default:
				panic("invalid")
			}
		}
		if err := w.Flush(); err != nil {
			t.Fatal(err)
		}
		if err := w.Error(); err != nil {
			t.Fatal(err)
		}
		r := NewReader(&bb)
		for i := 0; i < len(els); i++ {
			var v interface{}
			var err error
			switch els[i].(type) {
			case int8:
				v, err = r.ReadInt8()
			case int16:
				v, err = r.ReadInt16()
			case int32:
				v, err = r.ReadInt32()
			case int64:
				v, err = r.ReadInt64()
			case uint8:
				v, err = r.ReadUint8()
			case uint16:
				v, err = r.ReadUint16()
			case uint32:
				v, err = r.ReadUint32()
			case uint64:
				v, err = r.ReadUint64()
			case float32:
				v, err = r.ReadFloat32()
			case float64:
				v, err = r.ReadFloat64()
			case bool:
				v, err = r.ReadBool()
			case string:
				v, err = r.ReadString()
			case []byte:
				v, err = r.ReadBytes()
			case uvarint:
				v, err = r.ReadUvarint()
			case varint:
				v, err = r.ReadVarint()
			case bbyte:
				v, err = r.ReadByte()
			default:
				panic("invalid")
			}
			if err != nil {
				t.Fatalf("expected nil, got '%v'", err)
			}
			var eq bool
			switch el := els[i].(type) {
			case []byte:
				eq = string(v.([]byte)) == string(el)
			case uvarint:
				eq = uint64(v.(uint64)) == uint64(el)
			case varint:
				eq = int64(v.(int64)) == int64(el)
			case bbyte:
				eq = byte(v.(byte)) == byte(el)
			default:
				eq = v == els[i]
			}
			if !eq {
				t.Fatalf("expected %T'%v', got %T'%#v'", els[i], els[i], v, v)
			}
		}
		if err := r.Error(); err != nil {
			t.Fatal(err)
		}
		_, err := r.ReadByte()
		if err != io.EOF {
			t.Fatalf("expected '%v', got '%v'", io.EOF, err)
		}
		if err := r.Error(); err != io.EOF {
			t.Fatalf("expected '%v', got '%v'", io.EOF, err)
		}
	}
}

var errFake = errors.New("fake error")

func TestWriteErrors(t *testing.T) {
	var bb testBytesBuffer
	bb.err = errFake
	w := NewWriter(&bb)

	if err := w.WriteBool(true); err != nil {
		// This should work because the underlying buffered writer
		t.Fatalf("expected '%v', got '%v'", nil, err)
	}
	if err := w.Flush(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteInt8(0); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteInt16(0); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteInt32(0); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteInt64(0); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteUint8(0); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteUint16(0); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteUint32(0); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteUint64(0); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteFloat32(0); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteFloat64(0); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteBool(false); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteString(""); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteBytes(nil); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteUvarint(uint64(0)); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteVarint(int64(0)); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if err := w.WriteByte(byte(0)); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
}

func TestReadErrors(t *testing.T) {
	var bb testBytesBuffer
	bb.err = errFake
	r := NewReader(&bb)

	if _, err := r.ReadInt8(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadInt16(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadInt32(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadInt64(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadUint8(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadUint16(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadUint32(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadUint64(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadFloat32(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadFloat64(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadBool(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadString(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadBytes(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadUvarint(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadVarint(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}
	if _, err := r.ReadByte(); err != errFake {
		t.Fatalf("expected '%v', got '%v'", errFake, err)
	}

}
