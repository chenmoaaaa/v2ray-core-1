package mux

import (
	"io"

	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/common/serial"
)

type MetadataReader struct {
	reader io.Reader
	buffer []byte
}

func NewMetadataReader(reader io.Reader) *MetadataReader {
	return &MetadataReader{
		reader: reader,
		buffer: make([]byte, 1024),
	}
}

func (r *MetadataReader) Read() (*FrameMetadata, error) {
	metaLen, err := serial.ReadUint16(r.reader)
	if err != nil {
		return nil, err
	}
	if metaLen > 512 {
		return nil, newError("invalid metalen ", metaLen).AtWarning()
	}

	if _, err := io.ReadFull(r.reader, r.buffer[:metaLen]); err != nil {
		return nil, err
	}
	return ReadFrameFrom(r.buffer)
}

type PacketReader struct {
	reader io.Reader
	eof    bool
}

func NewPacketReader(reader io.Reader) *PacketReader {
	return &PacketReader{
		reader: reader,
		eof:    false,
	}
}

func (r *PacketReader) Read() (buf.MultiBuffer, error) {
	if r.eof {
		return nil, io.EOF
	}

	size, err := serial.ReadUint16(r.reader)
	if err != nil {
		return nil, err
	}

	var b *buf.Buffer
	if size <= buf.Size {
		b = buf.New()
	} else {
		b = buf.NewLocal(int(size))
	}
	if err := b.AppendSupplier(buf.ReadFullFrom(r.reader, int(size))); err != nil {
		b.Release()
		return nil, err
	}
	r.eof = true
	return buf.NewMultiBufferValue(b), nil
}

type StreamReader struct {
	reader   io.Reader
	leftOver int
}

func NewStreamReader(reader io.Reader) *StreamReader {
	return &StreamReader{
		reader:   reader,
		leftOver: -1,
	}
}

func (r *StreamReader) Read() (buf.MultiBuffer, error) {
	if r.leftOver == 0 {
		r.leftOver = -1
		return nil, io.EOF
	}

	if r.leftOver == -1 {
		size, err := serial.ReadUint16(r.reader)
		if err != nil {
			return nil, err
		}
		r.leftOver = int(size)
	}

	mb := buf.NewMultiBuffer()
	for r.leftOver > 0 {
		readLen := buf.Size
		if r.leftOver < readLen {
			readLen = r.leftOver
		}
		b := buf.New()
		if err := b.AppendSupplier(func(bb []byte) (int, error) {
			return r.reader.Read(bb[:readLen])
		}); err != nil {
			mb.Release()
			return nil, err
		}
		r.leftOver -= b.Len()
		mb.Append(b)
		if b.Len() < readLen {
			break
		}
	}
	return mb, nil
}
