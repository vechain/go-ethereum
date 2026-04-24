// Copyright 2022 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package rlp

import (
	"io"
	"math/big"

	"github.com/holiman/uint256"
)

func encBufferFromWriter(w io.Writer) *encBuffer {
	switch w := w.(type) {
	case EncoderBuffer:
		return w.buf
	case *EncoderBuffer:
		return w.buf
	case *encBuffer:
		return w
	default:
		return nil
	}
}

// EncoderBuffer is a buffer for incremental encoding.
//
// The zero value is NOT ready for use. To get a usable buffer,
// create it using NewEncoderBuffer or call Reset.
type EncoderBuffer struct {
	buf *encBuffer
	dst io.Writer

	ownBuffer bool
}

// NewEncoderBuffer creates an encoder buffer.
func NewEncoderBuffer(dst io.Writer) EncoderBuffer {
	var w EncoderBuffer
	w.Reset(dst)
	return w
}

// Reset truncates the buffer and sets the output destination.
func (w *EncoderBuffer) Reset(dst io.Writer) {
	if w.buf != nil && !w.ownBuffer {
		panic("can't Reset derived EncoderBuffer")
	}

	// If the destination writer has an *encBuffer, use it.
	// Note that w.ownBuffer is left false here.
	if dst != nil {
		if outer := encBufferFromWriter(dst); outer != nil {
			*w = EncoderBuffer{outer, nil, false}
			return
		}
	}

	// Get a fresh buffer.
	if w.buf == nil {
		w.buf = encBufferPool.Get().(*encBuffer)
		w.ownBuffer = true
	}
	w.buf.reset()
	w.dst = dst
}

// Flush writes encoded RLP data to the output writer. This can only be called once.
// If you want to re-use the buffer after Flush, you must call Reset.
func (w *EncoderBuffer) Flush() error {
	var err error
	if w.dst != nil {
		err = w.buf.writeTo(w.dst)
	}
	// Release the internal buffer.
	if w.ownBuffer {
		encBufferPool.Put(w.buf)
	}
	*w = EncoderBuffer{}
	return err
}

// ToBytes returns the encoded bytes.
func (w *EncoderBuffer) ToBytes() []byte {
	return w.buf.makeBytes()
}

// AppendToBytes appends the encoded bytes to dst.
func (w *EncoderBuffer) AppendToBytes(dst []byte) []byte {
	out := w.buf.makeBytes()
	return append(dst, out...)
}

// Size returns the total size of the content that was encoded up to this point.
// Note this does not count the size of any lists which are still 'open' (i.e. for
// which ListEnd has not been called yet).
func (w EncoderBuffer) Size() int {
	if w.buf == nil {
		return 0
	}
	return w.buf.size()
}

// Write appends b directly to the encoder output.
func (w EncoderBuffer) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

// WriteBool writes b as the integer 0 (false) or 1 (true).
func (w EncoderBuffer) WriteBool(b bool) {
	w.buf.writeBool(b)
}

// WriteUint64 encodes an unsigned integer.
func (w EncoderBuffer) WriteUint64(i uint64) {
	w.buf.writeUint64(i)
}

// WriteBigInt encodes a big.Int as an RLP string.
// Note: Unlike with Encode, the sign of i is ignored.
func (w EncoderBuffer) WriteBigInt(i *big.Int) {
	w.buf.writeBigInt(i)
}

// WriteUint256 encodes uint256.Int as an RLP string.
func (w EncoderBuffer) WriteUint256(i *uint256.Int) {
	w.buf.writeUint256(i)
}

// WriteBytes encodes b as an RLP string.
func (w EncoderBuffer) WriteBytes(b []byte) {
	w.buf.writeBytes(b)
}

// WriteString encodes s as an RLP string.
func (w EncoderBuffer) WriteString(s string) {
	w.buf.writeString(s)
}

// List starts a list. It returns an internal index. Call ListEnd with
// this index after encoding the content to finish the list.
func (w EncoderBuffer) List() int {
	return w.buf.list()
}

// ListEnd finishes the given list.
func (w EncoderBuffer) ListEnd(index int) {
	w.buf.listEnd(index)
}
