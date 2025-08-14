package pkg

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

func TestGob[T any](in *T) {
	// Marshal to bytes
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	start := time.Now()
	if err := enc.Encode(in); err != nil {
		log.Fatalf("gob marshal failed: %v", err)
	}
	marshalTime := time.Since(start)

	marshaledData := buf.Bytes()
	size := len(marshaledData)
	fmt.Printf("Marshal Time: %s\n", marshalTime)
	fmt.Printf("Size: %d bytes\n", size)

	// Unmarshal from bytes
	dec := gob.NewDecoder(bytes.NewReader(marshaledData))
	var unmarshaledUsers T
	start = time.Now()
	if err := dec.Decode(&unmarshaledUsers); err != nil {
		log.Fatalf("gob unmarshal failed: %v", err)
	}
	unmarshalTime := time.Since(start)

	fmt.Printf("Unmarshal Time: %s\n\n", unmarshalTime)
}

func EncodeGob(in any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(in); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DecodeGob(src []byte, target any) error {
	dec := gob.NewDecoder(bytes.NewReader(src))
	if err := dec.Decode(target); err != nil {
		return err
	}

	return nil
}
