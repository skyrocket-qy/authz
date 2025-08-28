package util

import (
	"bytes"
	"encoding/gob"

	"github.com/skyrocket-qy/erx"
)

func EncodeGob(in any) ([]byte, error) {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(in); err != nil {
		return nil, erx.W(err)
	}

	return buf.Bytes(), nil
}

func DecodeGob(src []byte, target any) error {
	dec := gob.NewDecoder(bytes.NewReader(src))
	if err := dec.Decode(target); err != nil {
		return erx.W(err)
	}

	return nil
}
