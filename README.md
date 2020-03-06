# snapbits

[![GoDoc](https://godoc.org/github.com/tidwall/snapbits?status.svg)](https://godoc.org/github.com/tidwall/snapbits)

This package provides an fast and simple way for reading and writing snappy
compressed bit streams.

Supports reading and writing most basic types including: 
`int8`, `int16`, `int32`, `int64`, `uint8`, `uint16`, `uint32`, `uint64`,
`byte`, `bool`, `float32`, `float64`, `[]byte`, `string`.
Also `uvarint` and `varint`. 

*Now isn't that nice.*

## Usage

### Installing

To start using Snapbits, install Go and run `go get`:

```sh
$ go get -u github.com/tidwall/snapbits
```

### Basic operations

```go
// create a writer
var bb bytes.Buffer
w := snapbits.NewWriter(&bb) 

// write some stuff
err = w.WriteString("Hello Jello")
err = w.WriteBytes(someBinary)
err = w.WriteUvarint(8589869056)
err = w.WriteVarint(-119290019)
err = w.WriteUint16(-119290019)

// close the reader when done
w.Close()

// create a reader
r := snapbits.NewReader(&bb)

// read some stuff
s, err = w.ReadString()
b, err = w.ReadBytes()
x, err = w.ReadUvarint()
x, err = w.ReadVarint()
x, err = w.ReadUint16()
```

## License

`snapbits` source code is available under the MIT License.