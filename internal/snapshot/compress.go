package snapshot

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
)

type Compressor struct {
	level int
}

func NewCompressor(level int) *Compressor {
	if level < gzip.DefaultCompression || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}
	return &Compressor{level: level}
}

func NewDefaultCompressor() *Compressor {
	return &Compressor{level: gzip.DefaultCompression}
}

func (c *Compressor) Compress(snap *Snapshot) (*Snapshot, error) {
	if snap == nil {
		return nil, fmt.Errorf("snapshot is nil")
	}

	data, err := json.Marshal(snap)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}

	var buf bytes.Buffer
	gw, err := gzip.NewWriterLevel(&buf, c.level)
	if err != nil {
		return nil, fmt.Errorf("create gzip writer: %w", err)
	}

	if _, err := gw.Write(data); err != nil {
		return nil, fmt.Errorf("write compressed data: %w", err)
	}

	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("close gzip writer: %w", err)
	}

	compressed := buf.Bytes()

	checksum := sha256.Sum256(compressed)

	snap.Size = int64(len(compressed))
	snap.Compressed = true
	snap.Checksum = fmt.Sprintf("%x", checksum)

	return snap, nil
}

func (c *Compressor) Decompress(data []byte) (*Snapshot, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create gzip reader: %w", err)
	}
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		return nil, fmt.Errorf("read decompressed data: %w", err)
	}

	var snap Snapshot
	if err := json.Unmarshal(decompressed, &snap); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}

	return &snap, nil
}

func (c *Compressor) VerifyChecksum(snap *Snapshot, data []byte) (bool, error) {
	checksum := sha256.Sum256(data)
	computed := fmt.Sprintf("%x", checksum)
	return computed == snap.Checksum, nil
}

func (c *Compressor) CompressRaw(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw, err := gzip.NewWriterLevel(&buf, c.level)
	if err != nil {
		return nil, fmt.Errorf("create gzip writer: %w", err)
	}

	if _, err := gw.Write(data); err != nil {
		return nil, fmt.Errorf("write compressed data: %w", err)
	}

	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

func (c *Compressor) DecompressRaw(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create gzip reader: %w", err)
	}
	defer gr.Close()

	return io.ReadAll(gr)
}

func (c *Compressor) CompressAndChecksum(data []byte) (compressed []byte, checksum string, err error) {
	compressed, err = c.CompressRaw(data)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(compressed)
	checksum = fmt.Sprintf("%x", sum)
	return compressed, checksum, nil
}
