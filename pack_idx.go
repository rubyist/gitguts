package gitguts

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

var (
	packIndexSignature = []byte{0xFF, 0x74, 0x4F, 0x63}

	errInvalidPackIndex   = errors.New("invalid pack index")
	errInvalidPackVersion = errors.New("invalid pack version")
	errObjectNotFound     = errors.New("object not found")
)

// PackIndex describes a git pack file index.
type PackIndex struct {
	packIndexHeader
	objects  []OID
	offsets  []uint32
	offsets2 []uint64
}

type packIndexHeader struct {
	Signature [4]byte
	Version   uint32
	Fanout    [256]uint32
}

// OpenPackIndex parses a git pack index, returning a PackIndex.
func OpenPackIndex(path string) (*PackIndex, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	idx := &PackIndex{}

	// Read in the pack header and the fanout table
	if err := binary.Read(f, binary.BigEndian, &idx.packIndexHeader); err != nil {
		return nil, err
	}

	// Verify the signature
	if bytes.Compare(idx.Signature[:], packIndexSignature) != 0 {
		return nil, errInvalidPackIndex
	}

	// We're just doing version 2 packs
	if idx.Version != 2 {
		return nil, errInvalidPackVersion
	}

	// The last entry of the fanout table tells us how many
	// objects are in the pack.
	numObjects := idx.Fanout[255]

	// Read in the object names table
	idx.objects = make([]OID, numObjects)
	if err := binary.Read(f, binary.BigEndian, idx.objects); err != nil {
		return nil, err
	}

	// Skip the CRC table and go to the offsets table
	if _, err := f.Seek(int64(numObjects*4), io.SeekCurrent); err != nil {
		return nil, err
	}

	// Read in the offsets table
	idx.offsets = make([]uint32, numObjects)
	if err := binary.Read(f, binary.BigEndian, idx.offsets); err != nil {
		return nil, err
	}

	// Read in the fourth fan out layer, if necessary
	bigOffsets := 0
	for i := 0; i < len(idx.offsets); i++ {
		if idx.offsets[i]&0x8000 > 0 {
			bigOffsets++
		}
	}

	// Pack files larger than 2G have a secondary offsets
	// table. If an offset value has the MSB set to 1, the
	// remaining bits are an offset into a secondary table of 8
	// byte offsets.
	if bigOffsets > 0 {
		idx.offsets2 = make([]uint64, bigOffsets)
		if err := binary.Read(f, binary.BigEndian, idx.offsets2); err != nil {
			return nil, err
		}
	}

	// The end of the file contains a 20 byte SHA-1 checksum of
	// the associated pack file and a 20 byte SHA-1 checksum of
	// the index file.

	return idx, nil
}

// OffsetOf locate the offset of the object.
func (p *PackIndex) OffsetOf(oid OID) (int, error) {
	// Calculate the number of names starting with oid[0]
	oid0 := int(oid[0])
	before := uint32(0)
	if oid0 > 0 {
		before = p.Fanout[oid0-1]
	}
	numOIDs := p.Fanout[oid0] - before

	offsetIndex := searchOIDs(p.objects[before:], 0, numOIDs-1, oid)
	if offsetIndex < 0 {
		return 0, errObjectNotFound
	}

	offsetIndex += int(before)

	offset := p.offsets[offsetIndex]
	if offset&0x8000 > 0 {
		return int(p.offsets2[offset&0x7FFF]), nil
	}

	return int(p.offsets[offsetIndex]), nil
}

// searchOIDs perform a binary search on the set of OIDs to find the
// index of the oid.
func searchOIDs(oids []OID, l, r uint32, oid OID) int {
	for l <= r {
		m := l + (r-l)/2
		switch bytes.Compare(oids[m][:], oid[:]) {
		case 0:
			return int(m)
		case 1:
			r = m - 1
		default:
			l = m + 1
		}
	}
	return -1
}
