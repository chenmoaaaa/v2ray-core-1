package crypto

import (
	"io"

	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/common/serial"
)

type ChunkSizeDecoder interface {
	SizeBytes() int
	Decode([]byte) (uint16, error)
}

type ChunkSizeEncoder interface {
	SizeBytes() int
	Encode(uint16, []byte) []byte
}

type PlainChunkSizeParser struct{}

func (PlainChunkSizeParser) SizeBytes() int {
	return 2
}

func (PlainChunkSizeParser) Encode(size uint16, b []byte) []byte {
	return serial.Uint16ToBytes(size, b)
}

func (PlainChunkSizeParser) Decode(b []byte) (uint16, error) {
	return serial.BytesToUint16(b), nil
}

type ChunkStreamReader struct {
	sizeDecoder ChunkSizeDecoder
	reader      buf.Reader

	buffer       []byte
	leftOver     buf.MultiBuffer
	leftOverSize int
}

func NewChunkStreamReader(sizeDecoder ChunkSizeDecoder, reader io.Reader) *ChunkStreamReader {
	return &ChunkStreamReader{
		sizeDecoder: sizeDecoder,
		reader:      buf.NewReader(reader),
		buffer:      make([]byte, sizeDecoder.SizeBytes()),
	}
}

func (r *ChunkStreamReader) readAtLeast(size int) error {
	mb := r.leftOver
	r.leftOver = nil
	for mb.Len() < size {
		extra, err := r.reader.Read()
		if err != nil {
			mb.Release()
			return err
		}
		mb.AppendMulti(extra)
	}
	r.leftOver = mb

	return nil
}

func (r *ChunkStreamReader) readSize() (uint16, error) {
	if r.sizeDecoder.SizeBytes() > r.leftOver.Len() {
		if err := r.readAtLeast(r.sizeDecoder.SizeBytes() - r.leftOver.Len()); err != nil {
			return 0, err
		}
	}
	r.leftOver.Read(r.buffer)
	return r.sizeDecoder.Decode(r.buffer)
}

func (r *ChunkStreamReader) Read() (buf.MultiBuffer, error) {
	size := r.leftOverSize
	if size == 0 {
		nextSize, err := r.readSize()
		if err != nil {
			return nil, err
		}
		if nextSize == 0 {
			return nil, io.EOF
		}
		size = int(nextSize)
	}

	if r.leftOver.IsEmpty() {
		if err := r.readAtLeast(1); err != nil {
			return nil, err
		}
	}

	if size >= r.leftOver.Len() {
		mb := r.leftOver
		r.leftOverSize = size - r.leftOver.Len()
		r.leftOver = nil
		return mb, nil
	}

	mb := r.leftOver.SliceBySize(size)
	if mb.Len() != size {
		b := buf.New()
		b.AppendSupplier(buf.ReadFullFrom(&r.leftOver, size-mb.Len()))
		mb.Append(b)
	}

	r.leftOverSize = 0
	return mb, nil
}

type ChunkStreamWriter struct {
	sizeEncoder ChunkSizeEncoder
	writer      buf.Writer
}

func NewChunkStreamWriter(sizeEncoder ChunkSizeEncoder, writer io.Writer) *ChunkStreamWriter {
	return &ChunkStreamWriter{
		sizeEncoder: sizeEncoder,
		writer:      buf.NewWriter(writer),
	}
}

func (w *ChunkStreamWriter) Write(mb buf.MultiBuffer) error {
	mb2Write := buf.NewMultiBuffer()
	const sliceSize = 8192

	for {
		slice := mb.SliceBySize(sliceSize)

		b := buf.New()
		b.AppendSupplier(func(buffer []byte) (int, error) {
			w.sizeEncoder.Encode(uint16(slice.Len()), buffer[:0])
			return w.sizeEncoder.SizeBytes(), nil
		})
		mb2Write.Append(b)
		mb2Write.AppendMulti(slice)

		if mb.IsEmpty() {
			break
		}
	}

	return w.writer.Write(mb2Write)
}
