// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package compression

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"sync"
)

// gzipCompressor implements Compressor
type gzipCompressor struct {
	lock sync.Mutex

	writeBuffer *bytes.Buffer
	gzipWriter  *gzip.Writer

	bytesReader *bytes.Reader
	gzipReader  *gzip.Reader
}

// Compress [msg] and returns the compressed bytes.
func (g *gzipCompressor) Compress(msg []byte) ([]byte, error) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.resetWriter()
	if _, err := g.gzipWriter.Write(msg); err != nil {
		return nil, err
	}
	if err := g.gzipWriter.Close(); err != nil {
		return nil, err
	}
	compressed := g.writeBuffer.Bytes()
	compressedCopy := make([]byte, len(compressed))
	copy(compressedCopy, compressed)
	return compressedCopy, nil
}

// Decompress decompresses [msg].
func (g *gzipCompressor) Decompress(msg []byte) ([]byte, error) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if err := g.resetReader(msg); err != nil {
		return nil, err
	}
	decompressed, err := ioutil.ReadAll(g.gzipReader)
	if err != nil {
		return nil, err
	}
	if err = g.gzipReader.Close(); err != nil {
		return nil, err
	}
	return decompressed, nil
}

func (g *gzipCompressor) resetWriter() {
	g.writeBuffer.Reset()
	g.gzipWriter.Reset(g.writeBuffer)
}

func (g *gzipCompressor) resetReader(msg []byte) error {
	g.bytesReader.Reset(msg)
	return g.gzipReader.Reset(g.bytesReader)
}

// NewGzipCompressor returns a new gzip Compressor that compresses
func NewGzipCompressor() Compressor {
	var buf bytes.Buffer
	return &gzipCompressor{
		writeBuffer: &buf,
		gzipWriter:  gzip.NewWriter(&buf),

		bytesReader: &bytes.Reader{},
		gzipReader:  &gzip.Reader{},
	}
}
