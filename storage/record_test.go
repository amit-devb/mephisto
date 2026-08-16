package storage

import (
	"bytes"
	"testing"
)

func TestRecordRoundTrip(t *testing.T) {
	// Let's define some test cases
	tests := []struct {
		name   string
		record Record
	}{
		{
			name: "Normal Put Operation",
			record: Record{
				Type:  OpPut,
				Key:   []byte("username"),
				Value: []byte("alice123"),
			},
		},
		{
			name: "Delete Operation (Empty Value)",
			record: Record{
				Type:  OpDelete,
				Key:   []byte("username"),
				Value: nil, // Or []byte{}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Encode the record
			encoded := tc.record.Encode()

			// 2. Wrap the bytes in a bytes.Buffer so it implements io.Reader
			reader := bytes.NewBuffer(encoded)

			// 3. Decode it back
			decoded, err := DecodeRecord(reader)
			if err != nil {
				t.Fatalf("failed to decode record: %v", err)
			}

			// $. Verify that decoded match tc.record
			if decoded.Type != tc.record.Type {
				t.Errorf("expected type %d, got %d", tc.record.Type, decoded.Type)
			}

			if !bytes.Equal(decoded.Key, tc.record.Key) {
				t.Errorf("expected key %s, got %s", tc.record.Key, decoded.Key)
			}

			if !bytes.Equal(decoded.Value, tc.record.Value) {
				t.Errorf("expected value %s, got %s", tc.record.Value, decoded.Value)
			}
		})
	}
}

func TestRecordCorruption(t *testing.T) {
	record := Record{
		Type:  OpPut,
		Key:   []byte("important-key"),
		Value: []byte("secure-data"),
	}

	encoded := record.Encode()

	// Corrupt a byte in the payload (let's say the last byte)
	encoded[len(encoded)-1] = encoded[len(encoded)-1] ^ 0xFF

	reader := bytes.NewReader(encoded)
	_, err := DecodeRecord(reader)

	if err != ErrChecksumMismatch {
		t.Fatalf("expected ErrChecksumMismatch, got: %v", err)
	}
}

func TestRecordTruncation(t *testing.T) {
	record := Record{
		Type:  OpPut,
		Key:   []byte("another-key"),
		Value: []byte("some-data"),
	}

	encoded := record.Encode()

	// Let's truncate the record (chop off the last 3 bytes)
	truncated := encoded[:len(encoded)-3]

	reader := bytes.NewReader(truncated)
	_, err := DecodeRecord(reader)

	if err != ErrTruncatedRecord {
		t.Fatalf("expected ErrTruncatedRecord, got: %v", err)
	}
}
