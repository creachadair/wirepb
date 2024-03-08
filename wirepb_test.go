package wirepb_test

import (
	"bytes"
	"encoding/base64"
	"io"
	"testing"

	"github.com/creachadair/wirepb"
)

func TestScanner(t *testing.T) {
	input, err := base64.URLEncoding.DecodeString(`Ch0KDy0zlZA0zlZDICFZBoCFZBIEdGVzdCABKAEwAhACGAEgAA==`)
	if err != nil {
		t.Fatalf("Decode input: %v", err)
	}

	s := wirepb.NewScanner(bytes.NewReader(input))
	for s.Next() == nil {
		t.Logf("id=%d %s %v", s.ID(), s.Type(), s.Data())
		if s.Type() == wirepb.Len && s.ID() == 1 {
			s2 := wirepb.NewScanner(bytes.NewReader(s.Data()))
			for s2.Next() == nil {
				t.Logf("| id=%d %s %v", s2.ID(), s2.Type(), s2.Data())
			}
			if err := s2.Err(); err != io.EOF {
				t.Errorf("| Scanner err: got %v, want io.EOF", err)
			}
		}
	}
	if err := s.Err(); err != io.EOF {
		t.Errorf("Scanner err: got %v, want io.EOF", err)
	}
}
