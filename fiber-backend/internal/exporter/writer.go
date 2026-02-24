package exporter

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/klauspost/compress/zstd"
)

type CompressedWriter struct {
	file           *os.File
	encoder        *zstd.Encoder
	blockSize      int
	currentBlock   []Point
	blockCount     int
	index          []BlockIndex
	compressedSize int64
	compression    bool
	mu             sync.Mutex
}

func NewCompressedWriter(file *os.File, compression bool) (*CompressedWriter, error) {
	// Magic header
	header := "PLCEXP1 "
	if !compression {
		header = "PLCRAW1 "
	}
	if _, err := file.Write([]byte(header)); err != nil {
		return nil, err
	}

	var encoder *zstd.Encoder
	var err error
	if compression {
		encoder, err = zstd.NewWriter(file, zstd.WithEncoderLevel(zstd.SpeedDefault))
		if err != nil {
			return nil, err
		}
	}

	return &CompressedWriter{
		file:         file,
		encoder:      encoder,
		blockSize:    1000,
		currentBlock: make([]Point, 0, 1000),
		index:        make([]BlockIndex, 0),
		compression:  compression,
	}, nil
}

func (w *CompressedWriter) WriteBatch(points []Point) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, p := range points {
		w.currentBlock = append(w.currentBlock, p)
		if len(w.currentBlock) >= w.blockSize {
			if err := w.flushBlock(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *CompressedWriter) flushBlock() error {
	if len(w.currentBlock) == 0 {
		return nil
	}

	data, err := json.Marshal(w.currentBlock)
	if err != nil {
		return err
	}

	// Record position before block data
	pos, _ := w.file.Seek(0, io.SeekCurrent)

	// Note: The encoder handles the actual compression and writing to the file
	// However, we want to know the compressed size for indexing.
	// Zstd Encoder doesn't expose written bytes easily without a wrapper.
	// For simplicity in this implementation, we'll flush the encoder to ensure data is written.

	// Write compressed block
	var compressed []byte
	if w.compression {
		compressed = w.encoder.EncodeAll(data, nil)
	} else {
		compressed = data
	}

	// Write block size first
	if err := binary.Write(w.file, binary.BigEndian, uint32(len(compressed))); err != nil {
		return err
	}

	n, err := w.file.Write(compressed)
	if err != nil {
		return err
	}
	w.compressedSize += int64(n)

	// Update index
	w.index = append(w.index, BlockIndex{
		Offset:     pos,
		Length:     int64(n) + 4, // data + size header
		PointCount: len(w.currentBlock),
		StartTime:  w.currentBlock[0].Timestamp,
		EndTime:    w.currentBlock[len(w.currentBlock)-1].Timestamp,
	})

	w.currentBlock = w.currentBlock[:0]
	w.blockCount++
	return nil
}

func (w *CompressedWriter) Close() error {
	if err := w.flushBlock(); err != nil {
		return err
	}

	// Write Index at the end
	indexData, err := json.Marshal(w.index)
	if err != nil {
		return err
	}

	// Write index size
	if err := binary.Write(w.file, binary.BigEndian, uint32(len(indexData))); err != nil {
		return err
	}

	// Write index data
	if _, err := w.file.Write(indexData); err != nil {
		return err
	}

	if w.compression && w.encoder != nil {
		return w.encoder.Close()
	}
	return nil
}
