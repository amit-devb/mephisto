package storage

import (
	"io"
	"os"
	"sync"
)

// WAL represents a Write-ahead log for disk storage.
type WAL struct {
	mu   sync.Mutex
	file *os.File
}

/*
1. Write appends a new record to the end of the WAL.
2. It is thread-safe (uses a mutex) and ensures durability on disk.
*/
func (w *WAL) Write(op OpType, key []byte, value []byte) error {
	// 1. Acquire a lock so only one thread can write at a time.
	w.mu.Lock()

	// defer tells Go: "Run w.mu.Unlock() automatically when this function finished"
	// This is a safety net so we never forget to unlock, even if an error occurs early!
	defer w.mu.Unlock()

	// 2. Create the Record
	rec := Record{
		Type:  op,
		Key:   key,
		Value: value,
	}

	// 3. Encode the Record to bytes
	encoded := rec.Encode()

	// 4. Write the Record to bytes
	_, err := w.file.Write(encoded)
	if err != nil {
		return err
	}

	// %. Force the OS to flush its memory buffere to physical storage (fsync)
	return w.file.Sync()
}

// Close closes the underlying log file
func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}

// ---------------------------- Functions --------------------------------
/*
1. OpenWAL opens a WAL file at the given path.
2. If the files does not exist, it creates it.
*/
func OpenWAL(path string) (*WAL, error) {
	/*
		1. Open the file with options:
			- os.O_CREATE: create it if it does not exist.
			- os.O_APPEND: append any new writes to the end of the file
			- os.O_WRONLY: write-only mode.
		2. Permission 0644: gives read and write permission to the owner and only read permission to group and others.
	*/
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{
		file: file,
	}, nil
}

/*
1. RecoverWAL reads the WAL file from start to finisg and returns all valid records.
2. If the WAL file does not exists, it returns (nill, nil)
*/
func RecoverWAL(path string) ([]*Record, error) {
	// 1. Check if the file exist, if it does not, nothing to recover
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}

	// 2. Open the file in read-only mode
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []*Record

	// 3. Read records in a loop untill we reach EOF
	for {
		rec, err := DecodeRecord(file)
		if err != nil {
			// io.EOF means we cleanly reached the end of the file
			if err == io.EOF {
				break
			}

			// If we hit corruption or truncation (from a crash)
			// return the records we recovered before the crash occured, and the error
			return records, err
		}
		records = append(records, rec)
	}

	return records, nil
}
