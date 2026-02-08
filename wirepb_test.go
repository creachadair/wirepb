package wirepb_test

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"math"
	"testing"

	"github.com/creachadair/wirepb"
)

func TestScanner(t *testing.T) {
	input, err := base64.URLEncoding.DecodeString(`Ch0KDy0zlZA0zlZDICFZBoCFZBIEdGVzdCABKAEwAhACGAEgAA==`)
	if err != nil {
		t.Fatalf("Decode input: %v", err)
	}

	s := wirepb.NewScanner(bytes.NewReader(input))
	for s.Next() {
		t.Logf("id=%d %s %v", s.ID(), s.Type(), s.Data())
		if s.Type() == wirepb.Len && s.ID() == 1 {
			s2 := wirepb.NewScanner(bytes.NewReader(s.Data()))
			for s2.Next() {
				t.Logf("| id=%d %s %v", s2.ID(), s2.Type(), s2.Data())
			}
			if err := s2.Err(); err != nil {
				t.Errorf("| Scanner unexpected error: %v", err)
			}
		}
	}
	if err := s.Err(); err != nil {
		t.Errorf("Scanner unexpected error: %v", err)
	}
}

func TestBuilder(t *testing.T) {
	var b wirepb.Builder

	if b.Len() != 0 {
		t.Errorf("Empty b.Len() = %d, want 0", b.Len())
	}

	// Write a bunch of data into the message, and verify it round-trips.
	b.Bool(1, true)
	b.Int32(2, -25)
	b.String(3, "hello")
	b.Message(4, func(sb *wirepb.Builder) {
		sb.Uint32(5, 100)
		sb.Float64(6, 3.14159)
	})
	b.Fixed32(7, 12345678)
	b.String(3, "world") // repeated, out of order

	t.Logf("Encoded message (%d bytes): %v", b.Len(), b.Bytes())
	s := wirepb.NewScanner(bytes.NewReader(b.Bytes()))
	var pos int
	advance := func() {
		if !s.Next() {
			if err := s.Err(); err != nil {
				t.Fatalf("Advance scanner: %v", err)
			}
		}
		pos++
	}
	check := func(wantID int, wantType wirepb.Type, wantData []byte) {
		advance()
		if s.ID() != wantID {
			t.Errorf("Token %d: got id=%d, want %d", pos, s.ID(), wantID)
		}
		if s.Type() != wantType {
			t.Errorf("Token %d: got type=%v, want %v", pos, s.Type(), wantType)
		}
		if wantData != nil && !bytes.Equal(s.Data(), wantData) {
			t.Errorf("Token %d: got data %v, want %v", pos, s.Data(), wantData)
		}
	}

	check(1, wirepb.Varint, []byte{1})
	check(2, wirepb.Varint, []byte{25*2 + 1}) // zig-zag
	check(3, wirepb.Len, []byte("hello"))
	check(4, wirepb.Len, nil) // contents checked below
	save, sp := s, pos

	s, pos = wirepb.NewScanner(bytes.NewReader(s.Data())), 0
	check(5, wirepb.Varint, []byte{100})
	check(6, wirepb.I64, binary.LittleEndian.AppendUint64(nil, math.Float64bits(3.14159)))
	advance()
	if s.Err() != nil {
		t.Errorf("Unexpected error at end of message (pos=%d): %v", pos, s.Err())
	}
	s, pos = save, sp

	check(7, wirepb.I32, []byte{0x4e, 0x61, 0xbc, 0}) // 12345678 in little-endian hex
	check(3, wirepb.Len, []byte("world"))
	advance()
	if s.Err() != nil {
		t.Errorf("Unexpected error at end of message (pos=%d): %v", pos, s.Err())
	}
}
