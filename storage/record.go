package storage

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

var (
	ErrChecksumMismatch = errors.New("record checksum mismatch: data is corrupted")
	ErrTruncatedRecord  = errors.New("record data is truncated")
)

// OpType represents the operation type (Put or Delete)
type OpType uint8

const (
	OpPut    OpType = 1
	OpDelete OpType = 2 // Tombstone for deleted keys
)

// Header Layout
// - CRC32: 4 bytes (uint32)
// - KeyLen: 4 bytes (uint32)
// - ValLen: 4 bytes (uint32)
// - OpType: 1 byte (uint8)
// = Total Header size: 13 bytes

const HeaderSize = 4 + 4 + 4 + 1 // 13 bytes

// Record represents a single key-value operations.
type Record struct {
	Type  OpType
	Key   []byte
	Value []byte
}

// Encode serialisez a Record into a byte slice with framing and CRC.
func (r *Record) Encode() []byte {
	totalSize := HeaderSize + len(r.Key) + len(r.Value)
	buf := make([]byte, totalSize)

	// 1. Pack KeyLen and ValLen, OpType into head (offsets 4..13)
	binary.BigEndian.PutUint32(buf[4:8], uint32(len(r.Key)))
	binary.BigEndian.PutUint32(buf[8:12], uint32(len(r.Value)))
	buf[12] = byte(r.Type)

	// 2. Cope Key and Value into paload (offsets 13..)
	copy(buf[13:13+len(r.Key)], r.Key)
	copy(buf[13+len(r.Key):], r.Value)

	// 3. Compute CRC32 Checksum over payload (OpType + Key + Value)
	// That is everything from index 12 to the end of the buffer
	checksum := crc32.ChecksumIEEE(buf[12:])
	binary.BigEndian.PutUint32(buf[0:4], checksum)

	return buf
}

// DecodeRecord reads a single framed Record from an io.Reader.
// Returns io.EOF if the stream ended cleanly before reading a new record.
func DecodeRecord(r io.Reader) (*Record, error) {
	headerBuf := make([]byte, HeaderSize)

	// Read the fixed 13-byte header
	_, err := io.ReadFull(r, headerBuf)
	if err != nil {
		return nil, err // Could be io.EOF if clean end of file
	}

	// Parse header fields
	expectedCRC := binary.BigEndian.Uint32(headerBuf[0:4])
	keyLen := binary.BigEndian.Uint32(headerBuf[4:8])
	valLen := binary.BigEndian.Uint32(headerBuf[8:12])
	opType := OpType(headerBuf[12])

	// Read payload: Key + Value
	payload := make([]byte, keyLen+valLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, ErrTruncatedRecord
		}
		return nil, err
	}

	// Verify CRC over (OpType + Key + Value)
	// We construct the checksum input matching Encode: byte(opType) + payload
	crcInput := append([]byte{byte(opType)}, payload...)
	actualCRC := crc32.ChecksumIEEE(crcInput)
	if actualCRC != expectedCRC {
		return nil, ErrChecksumMismatch
	}

	return &Record{
		Type:  opType,
		Key:   payload[:keyLen],
		Value: payload[keyLen:],
	}, nil
}
