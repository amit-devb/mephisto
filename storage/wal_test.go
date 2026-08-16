package storage

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestWALWriteAndRecover(t *testing.T) {
	// 1. Create a temporary directory for our test log file
	tmpDir := t.TempDir()
	walPath := filepath.Join(tmpDir, "wal.log")

	// 2. Open the WAL
	wal, err := OpenWAL(walPath)
	if err != nil {
		t.Fatalf("failed to open WAL: %v", err)
	}

	// 3. Write some records
	testRecords := []Record{
		{Type: OpPut, Key: []byte("key1"), Value: []byte("value1")},
		{Type: OpPut, Key: []byte("key2"), Value: []byte("value2")},
		{Type: OpDelete, Key: []byte("key1"), Value: nil},
	}

	for _, rec := range testRecords {
		err := wal.Write(rec.Type, rec.Key, rec.Value)
		if err != nil {
			t.Fatalf("failed to write record: %v", err)
		}
	}

	// 4. Close the WAL file
	err = wal.Close()
	if err != nil {
		t.Fatalf("failed to close WAL: %v", err)
	}

	// 5. Recover the records from the file path
	recovered, err := RecoverWAL(walPath)
	if err != nil {
		t.Fatalf("failed to recover WAL: %v", err)
	}

	// 6. Verify that recovered records match what we wrote
	if len(recovered) != len(testRecords) {
		t.Fatalf("expected %d records, got %d", len(testRecords), len(recovered))
	}

	for i, rec := range testRecords {
		got := recovered[i]
		if got.Type != rec.Type {
			t.Errorf("record %d: expected type %d, got %d", i, rec.Type, got.Type)
		}
		if !bytes.Equal(got.Key, rec.Key) {
			t.Errorf("record %d: expected key %s, got %s", i, rec.Key, got.Key)
		}
		if !bytes.Equal(got.Value, rec.Value) {
			t.Errorf("record %d: expected value %s, got %s", i, rec.Value, got.Value)
		}
	}
}

func TestRecoverNonExistentWAL(t *testing.T) {
	// verifies that recovering a path does not exist returns (nil, nil)
	recovered, err := RecoverWAL("non-existence-file-12345.log")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if recovered != nil {
		t.Errorf("expected nil records, got: %v", recovered)
	}
}
