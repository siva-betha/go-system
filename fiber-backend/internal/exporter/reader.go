package exporter

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"os"

	"github.com/klauspost/compress/zstd"
)

type CompressedReader struct {
	file        *os.File
	decoder     *zstd.Decoder
	index       []BlockIndex
	compression bool
}

func NewCompressedReader(file *os.File) (*CompressedReader, error) {
	// Verify magic header
	header := make([]byte, 8)
	if _, err := file.Read(header); err != nil {
		return nil, err
	}

	compression := false
	if string(header) == "PLCEXP1 " {
		compression = true
	} else if string(header) != "PLCRAW1 " {
		return nil, io.ErrUnexpectedEOF
	}

	var decoder *zstd.Decoder
	var err error
	if compression {
		decoder, err = zstd.NewReader(file)
		if err != nil {
			return nil, err
		}
	}

	r := &CompressedReader{
		file:        file,
		decoder:     decoder,
		compression: compression,
	}

	if err := r.readIndex(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *CompressedReader) readIndex() error {
	// Read index from end of file
	// [data blocks] [index size uint32] [index json]

	if _, err := r.file.Seek(-4, io.SeekEnd); err != nil {
		return err
	}

	var indexSize uint32
	if err := binary.Read(r.file, binary.BigEndian, &indexSize); err != nil {
		return err
	}

	if _, err := r.file.Seek(-4-int64(indexSize), io.SeekEnd); err != nil {
		return err
	}

	indexData := make([]byte, indexSize)
	if _, err := io.ReadFull(r.file, indexData); err != nil {
		return err
	}

	return json.Unmarshal(indexData, &r.index)
}

func (r *CompressedReader) ReadBlock(index int) ([]Point, error) {
	if index < 0 || index >= len(r.index) {
		return nil, io.EOF
	}

	block := r.index[index]
	if _, err := r.file.Seek(block.Offset, io.SeekStart); err != nil {
		return nil, err
	}

	var compressedSize uint32
	if err := binary.Read(r.file, binary.BigEndian, &compressedSize); err != nil {
		return nil, err
	}

	compressed := make([]byte, compressedSize)
	if _, err := io.ReadFull(r.file, compressed); err != nil {
		return nil, err
	}

	var decompressed []byte
	if r.compression {
		var err error
		decompressed, err = r.decoder.DecodeAll(compressed, nil)
		if err != nil {
			return nil, err
		}
	} else {
		decompressed = compressed
	}

	var points []Point
	if err := json.Unmarshal(decompressed, &points); err != nil {
		return nil, err
	}

	return points, nil
}

func (r *CompressedReader) Close() error {
	if r.compression && r.decoder != nil {
		r.decoder.Close()
	}
	return r.file.Close()
}
