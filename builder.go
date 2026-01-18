package wirepb

import (
	"encoding/binary"
	"math"
)

// A Builder is a constructor for a wire-format message. A zero value is ready
// for use and represents an empty message. The methods of a builder add fields
// to the message.
type Builder struct {
	buf []byte
}

// Bytes returns the current raw content of the builder.
//
// The builder retains ownership of the result, and the caller must not retain
// or modify the reported slice unless no further methods of the builder will
// be called.
func (b *Builder) Bytes() []byte { return b.buf }

// Len reports the length of the message so far constructed.
func (b *Builder) Len() int { return len(b.buf) }

// Bool appends a Boolean value to the message.
func (b *Builder) Bool(id int, ok bool) {
	if ok {
		b.uvarint(id, 1)
	} else {
		b.uvarint(id, 0)
	}
}

// Int32 appends a signed value to the message using zig-zag encoding.
func (b *Builder) Int32(id int, z int32) {
	if z < 0 {
		b.uvarint(id, 2*uint64(-z)+1)
	} else {
		b.uvarint(id, 2*uint64(z))
	}
}

// Uint32 appends an unsigned value to the message.
func (b *Builder) Uint32(id int, z uint32) { b.uvarint(id, uint64(z)) }

// Int64 appends a signed value to the message using zig-zag encoding.
func (b *Builder) Int64(id int, z int64) {
	if z < 0 {
		b.uvarint(id, 2*uint64(-z)+1)
	} else {
		b.uvarint(id, 2*uint64(z))
	}
}

// Uint64 appends an unsigned value to the message.
func (b *Builder) Uint64(id int, z uint64) { b.uvarint(id, z) }

// Fixed32 appends a 32-bit fixed-point value to the message.  The encoded byte
// order is little-endian.
func (b *Builder) Fixed32(id int, z uint32) {
	b.buf = packTag(id, I32, b.buf)
	b.buf = binary.LittleEndian.AppendUint32(b.buf, z)
}

// Fixed64 appends a 64-bit fixed-point value to the message.
// The encoded byte order is little endian.
func (b *Builder) Fixed64(id int, z uint64) {
	b.buf = packTag(id, I64, b.buf)
	b.buf = binary.LittleEndian.AppendUint64(b.buf, z)
}

// Float32 appends a 32-bit floating-point value to the message.
func (b *Builder) Float32(id int, f float32) {
	b.buf = packTag(id, I32, b.buf)
	b.buf = binary.LittleEndian.AppendUint32(b.buf, math.Float32bits(f))
}

// Float64 appends a 64-bit floating-point value to the message.
func (b *Builder) Float64(id int, f float64) {
	b.buf = packTag(id, I64, b.buf)
	b.buf = binary.LittleEndian.AppendUint64(b.buf, math.Float64bits(f))
}

// String appends data as an unstructured string value to the message.
func (b *Builder) String(id int, data string) {
	b.buf = packLen(id, data, b.buf)
}

// Message calls f with a new empty [Builder], and when f returns the contents
// of that builder are appended to b as a length-prefixed value.
func (b *Builder) Message(id int, f func(*Builder)) {
	var sub Builder
	f(&sub)
	b.buf = packLen(id, sub.buf, b.buf)
}

func (b *Builder) uvarint(id int, z uint64) {
	b.buf = packTag(id, Varint, b.buf)
	b.buf = binary.AppendUvarint(b.buf, z)
}

func packTag(id int, wtype Type, buf []byte) []byte {
	return binary.AppendUvarint(buf, (uint64(id)<<3)|uint64(wtype.Index()))
}

func packLen[S ~string | ~[]byte](id int, data S, buf []byte) []byte {
	buf = packTag(id, Len, buf)
	buf = binary.AppendUvarint(buf, uint64(len(data)))
	return append(buf, data...)
}
