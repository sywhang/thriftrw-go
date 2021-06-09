// Copyright (c) 2021 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package binary

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"

	"go.uber.org/thriftrw/internal/iface"
	"go.uber.org/thriftrw/protocol/stream"
	"go.uber.org/thriftrw/wire"
)

// StreamReader provides an implementation of a "stream.Reader".
type StreamReader struct {
	iface.Private

	reader io.Reader
	buffer [8]byte
}

// NewStreamReader returns a new StreamReader.
func NewStreamReader(r io.Reader) StreamReader {
	return StreamReader{reader: r}
}

func (sr *StreamReader) read(bs []byte) (int, error) {
	n, err := sr.reader.Read(bs)

	// For consistency with the non-streaming Reader, return an
	// io.ErrUnexpectedEOF if number of bytes read is smaller than what was
	// expected, https://golang.org/pkg/io/#ReaderAt
	if err == io.EOF || n < len(bs) {
		// All EOFs are unexpected when streaming
		err = io.ErrUnexpectedEOF
	}

	return n, err
}

func (sr *StreamReader) discard(n int64) error {
	_, err := io.CopyN(ioutil.Discard, sr.reader, n)
	if err == io.EOF {
		// All EOFs are unexpected when streaming
		err = io.ErrUnexpectedEOF
	}

	return err
}

// ReadBool reads a Thrift encoded bool value, returning a bool.
func (sr *StreamReader) ReadBool() (bool, error) {
	bs := sr.buffer[0:1]
	_, err := sr.read(bs)
	if err != nil {
		return false, err
	}

	switch bs[0] {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("invalid bool value: %q", bs[0])
	}
}

// ReadInt8 reads a Thrift encoded int8 value.
func (sr *StreamReader) ReadInt8() (int8, error) {
	bs := sr.buffer[0:1]
	_, err := sr.read(bs)
	return int8(bs[0]), err
}

// ReadInt16 reads a Thrift encoded int16 value.
func (sr *StreamReader) ReadInt16() (int16, error) {
	bs := sr.buffer[0:2]
	_, err := sr.read(bs)
	return int16(bigEndian.Uint16(bs)), err
}

// ReadInt32 reads a Thrift encoded int32 value.
func (sr *StreamReader) ReadInt32() (int32, error) {
	bs := sr.buffer[0:4]
	_, err := sr.read(bs)
	return int32(bigEndian.Uint32(bs)), err
}

// ReadInt64 reads a Thrift encoded int64 value.
func (sr *StreamReader) ReadInt64() (int64, error) {
	bs := sr.buffer[0:8]
	_, err := sr.read(bs)
	return int64(bigEndian.Uint64(bs)), err
}

// ReadString reads a Thrift encoded string.
func (sr *StreamReader) ReadString() (string, error) {
	bs, err := sr.ReadBinary()
	return string(bs), err
}

// ReadDouble reads a Thrift encoded double, returning a float64.
func (sr *StreamReader) ReadDouble() (float64, error) {
	val, err := sr.ReadInt64()
	return math.Float64frombits(uint64(val)), err
}

// ReadBinary reads a Thrift encoded binary type, returning a byte array.
func (sr *StreamReader) ReadBinary() ([]byte, error) {
	length, err := sr.ReadInt32()
	if err != nil {
		return nil, err
	}

	if length < 0 {
		return nil, fmt.Errorf("negative length %v specified for binary field", length)
	}

	if length == 0 {
		return []byte{}, nil
	}

	if length > bytesAllocThreshold {
		var buf bytes.Buffer
		_, err := io.CopyN(&buf, sr.reader, int64(length))
		if err == io.EOF {
			// All EOFs are unexpected when streaming
			err = io.ErrUnexpectedEOF
		}

		return buf.Bytes(), err
	}

	bs := make([]byte, length)
	_, err = sr.read(bs)
	return bs, err
}

// ReadStructBegin reads the "beginning" of a Thrift encoded struct.  Since
// there is no encoding for the beginning of a struct, this is a noop.
func (sr *StreamReader) ReadStructBegin() error {
	return nil
}

// ReadStructEnd reads the stop field of a Thrift encoded struct.
func (sr *StreamReader) ReadStructEnd() error {
	end, err := sr.ReadInt8()
	if err != nil {
		return err
	}

	if end != 0 {
		return fmt.Errorf("invalid stop field: %v", end)
	}

	return nil
}

// ReadFieldBegin reads off a Thrift encoded field-header.
func (sr *StreamReader) ReadFieldBegin() (stream.FieldHeader, bool, error) {
	fh := stream.FieldHeader{}

	fieldType, err := sr.ReadInt8()
	if err != nil {
		return fh, false, err
	}

	fieldID, err := sr.ReadInt16()
	if err != nil {
		return fh, false, err
	}

	fh.ID = fieldID
	fh.Type = wire.Type(fieldType)

	return fh, true, err
}

// ReadFieldEnd reads the "end" of a Thrift encoded field  Since there is no
// encoding for the end of a field, this is a noop.
func (sr *StreamReader) ReadFieldEnd() error {
	return nil
}

// ReadListBegin reads off the list header of a Thrift encoded list.
func (sr *StreamReader) ReadListBegin() (stream.ListHeader, error) {
	lh := stream.ListHeader{}

	elemType, listSize, err := sr.readTypeSizeHeader()
	if err != nil {
		return lh, err
	}

	lh.Type = wire.Type(elemType)
	lh.Length = int(listSize)
	return lh, nil
}

// ReadListEnd reads the "end" of a Thrift encoded list.  Since there is no
// encoding for the end of a list, this is a noop.
func (sr *StreamReader) ReadListEnd() error {
	return nil
}

// ReadSetBegin reads off the set header of a Thrift encoded set.
func (sr *StreamReader) ReadSetBegin() (stream.SetHeader, error) {
	sh := stream.SetHeader{}

	elemType, setSize, err := sr.readTypeSizeHeader()
	if err != nil {
		return sh, err
	}

	sh.Type = elemType
	sh.Length = setSize
	return sh, nil
}

// ReadSetEnd reads the "end" of a Thrift encoded list.  Since there is no
// encoding for the end of a set, this is a noop.
func (sr *StreamReader) ReadSetEnd() error {
	return nil
}

func (sr *StreamReader) readTypeSizeHeader() (wire.Type, int, error) {
	elemType, err := sr.ReadInt8()
	if err != nil {
		return wire.Type(0), 0, err
	}

	size, err := sr.ReadInt32()
	if err != nil {
		return wire.Type(0), 0, err
	}

	if size < 0 {
		return wire.Type(0), 0, fmt.Errorf("got negative length: %v", size)
	}

	return wire.Type(elemType), int(size), nil
}

// ReadMapBegin reads off the map header of a Thrift encoded map.
func (sr *StreamReader) ReadMapBegin() (stream.MapHeader, error) {
	mh := stream.MapHeader{}

	keyType, err := sr.ReadInt8()
	if err != nil {
		return mh, err
	}

	valueType, err := sr.ReadInt8()
	if err != nil {
		return mh, err
	}

	size, err := sr.ReadInt32()
	if err != nil {
		return mh, err
	}

	if size < 0 {
		return mh, fmt.Errorf("got negative length: %v", size)
	}

	mh.KeyType = wire.Type(keyType)
	mh.ValueType = wire.Type(valueType)
	mh.Length = int(size)
	return mh, nil
}

// ReadMapEnd reads the "end" of a Thrift encoded list.  Since there is no
// encoding for the end of a map, this is a noop.
func (sr *StreamReader) ReadMapEnd() error {
	return nil
}

// Skip skips fully over the provided Thrift type.
func (sr *StreamReader) Skip(t wire.Type) error {
	if w := fixedWidth(t); w > 0 {
		return sr.discard(w)
	}

	switch t {
	case wire.TBinary:
		length, err := sr.ReadInt32()
		if err != nil {
			return err
		}

		if length < 0 {
			return fmt.Errorf("got negative length: %v", length)
		}

		return sr.discard(int64(length))
	case wire.TStruct:
		return sr.skipStruct()
	case wire.TMap:
		return sr.skipMap()
	case wire.TSet:
		return sr.skipList()
	case wire.TList:
		return sr.skipList()
	default:
		return fmt.Errorf("unknown ttype %v", t)
	}
}

func (sr *StreamReader) skipStruct() error {
	fieldType, err := sr.ReadInt8()
	if err != nil {
		return err
	}

	for fieldType != 0 {
		// field id
		if err = sr.discard(int64(2)); err != nil {
			return err
		}

		if err = sr.Skip(wire.Type(fieldType)); err != nil {
			return err
		}

		if fieldType, err = sr.ReadInt8(); err != nil {
			return err
		}
	}

	return nil
}

func (sr *StreamReader) skipMap() error {
	keyRaw, err := sr.ReadInt8()
	if err != nil {
		return err
	}

	valueRaw, err := sr.ReadInt8()
	if err != nil {
		return err
	}

	size, err := sr.ReadInt32()
	if err != nil {
		return err
	}

	if size < 0 {
		return fmt.Errorf("got negative length: %v", size)
	}

	key := wire.Type(keyRaw)
	keyWidth := fixedWidth(key)
	value := wire.Type(valueRaw)
	valueWidth := fixedWidth(value)

	if keyWidth > 0 && valueWidth > 0 {
		length := int64(size) * (keyWidth + valueWidth)
		return sr.discard(length)
	}

	for i := int32(0); i < size; i++ {
		if err := sr.Skip(key); err != nil {
			return err
		}

		if err := sr.Skip(value); err != nil {
			return err
		}
	}

	return nil
}

func (sr *StreamReader) skipList() error {
	elemType, size, err := sr.readTypeSizeHeader()
	if err != nil {
		return err
	}

	width := fixedWidth(elemType)
	if width > 0 {
		length := width * int64(size)
		return sr.discard(length)
	}

	for i := 0; i < size; i++ {
		if err := sr.Skip(elemType); err != nil {
			return err
		}
	}

	return nil
}
