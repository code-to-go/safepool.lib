package core

import (
	"bytes"
	"hash"
	"io"

	"github.com/chmduquesne/rollinghash/buzhash32"
)

const windowSize = 32

type HashBlock struct {
	Hash   Hash256
	Length uint32
}

func getHashBlock(hashFun hash.Hash, length uint32, bufs ...[]byte) HashBlock {
	hashFun.Reset()
	for _, buf := range bufs {
		hashFun.Write(buf)
	}
	block := HashBlock{
		Length: length,
	}
	copy(block.Hash[:], hashFun.Sum(nil))
	return block
}

func HashSplit(hashFun hash.Hash, r io.Reader, splitBits uint) (blocks []HashBlock, err error) {
	var zeroes [windowSize]byte
	h := buzhash32.New()
	h.Write(zeroes[:])
	mask := uint32(0xffffffff)
	mask = mask >> uint32(32-splitBits)

	buf := make([]byte, 0, mask*2)
	inp := make([]byte, 1024)

	for {
		n, err := r.Read(inp)
		if err == io.EOF {
			if len(buf) > 0 {
				blocks = append(blocks, getHashBlock(hashFun, uint32(len(buf)), buf))
			}
			break
		} else if err != nil {
			return nil, err
		}

		step := 1
		for i := 0; i < n; i += step {
			b := inp[i]
			h.Roll(inp[i])
			buf = append(buf, b)

			sum32 := h.Sum32()
			if sum32&mask == mask {
				blocks = append(blocks, getHashBlock(hashFun, uint32(len(buf)), buf))
				buf = buf[:0]
			}
		}
	}
	blocks = append(blocks, getHashBlock(hashFun, uint32(len(buf)), buf))

	return blocks, err
}

type Diff struct {
	Id           uint32
	SourceStart  uint32
	SourceLength uint32
	DestStart    uint32
	DestLength   uint32
}

func HashDiff(source, dest []HashBlock) []Diff {
	sLen := len(source)
	dLen := len(dest)
	column := make([]int, sLen+1)

	var diffs []Diff
	var sOffset, dOffset uint32

	for y := 1; y <= sLen; y++ {
		column[y] = y
	}

	for x := 1; x <= dLen; x++ {
		column[0] = x
		lastkey := x - 1
		for y := 1; y <= sLen; y++ {
			oldkey := column[y]
			var incr int

			if bytes.Compare(source[y-1].Hash[:], dest[x-1].Hash[:]) != 0 {
				incr = 1
			}

			insert := column[y] + 1
			delete := column[y-1] + 1
			if insert < delete && insert < lastkey+incr {
				diffs = append(diffs, Diff{
					Id:           uint32(y),
					SourceStart:  sOffset,
					SourceLength: 0,
					DestStart:    dOffset,
					DestLength:   dest[x-1].Length,
				})
				column[y] = insert
			} else if delete < lastkey+incr {
				diffs = append(diffs, Diff{
					Id:           uint32(y),
					SourceStart:  sOffset,
					SourceLength: source[y-1].Length,
					DestStart:    dOffset,
					DestLength:   0,
				})
				column[y] = delete
				diffs = append(diffs, Diff{})
			} else {
				column[y] = lastkey + incr
				diffs = append(diffs, Diff{
					Id:           uint32(y),
					SourceStart:  sOffset,
					SourceLength: source[y-1].Length,
					DestStart:    dOffset,
					DestLength:   dest[x-1].Length,
				})
			}
			lastkey = oldkey
			sOffset += source[y-1].Length
		}
		dOffset += dest[x-1].Length
	}
	return diffs
}
