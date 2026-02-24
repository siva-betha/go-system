package storage

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/klauspost/compress/zstd"
)

type ArrayStorage struct {
	basePath      string
	currentFile   *os.File
	currentWriter *zstd.Encoder
	fileSize      int64
	maxFileSize   int64

	mu sync.Mutex
}

type Spectrum struct {
	Timestamp   time.Time
	SequenceNum uint32
	Intensities []uint16
}

func NewArrayStorage(basePath string) (*ArrayStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}
	return &ArrayStorage{
		basePath:    basePath,
		maxFileSize: 100 * 1024 * 1024, // 100 MB
	}, nil
}

func (as *ArrayStorage) StoreSpectrum(spectrum *Spectrum) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Check if need new file
	if as.currentFile == nil || as.fileSize > as.maxFileSize {
		if err := as.rotateFile(); err != nil {
			return err
		}
	}

	// Write spectrum header
	binary.Write(as.currentWriter, binary.BigEndian, spectrum.Timestamp.UnixNano())
	binary.Write(as.currentWriter, binary.BigEndian, spectrum.SequenceNum)
	binary.Write(as.currentWriter, binary.BigEndian, uint16(len(spectrum.Intensities)))

	// Write intensities
	for _, val := range spectrum.Intensities {
		binary.Write(as.currentWriter, binary.BigEndian, val)
	}

	// Note: In a real implementation, we'd update file size by checking the underlying file's size
	// or keeping track of written bytes before compression.
	return nil
}

func (as *ArrayStorage) rotateFile() error {
	// Close current file
	if as.currentWriter != nil {
		as.currentWriter.Close()
	}
	if as.currentFile != nil {
		as.currentFile.Close()
	}

	// Create new file
	filename := filepath.Join(as.basePath,
		fmt.Sprintf("spectra_%s.zst", time.Now().Format("20060102_150405")))

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	encoder, err := zstd.NewWriter(file)
	if err != nil {
		file.Close()
		return err
	}

	as.currentFile = file
	as.currentWriter = encoder
	as.fileSize = 0

	return nil
}
