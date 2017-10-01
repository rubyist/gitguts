package gitguts

import (
	"encoding/hex"
	"fmt"
)

// OID holds the 20 byte SHA-1 hash of a git object.
type OID [20]byte

// String is the string representation of the OID
func (o OID) String() string {
	return fmt.Sprintf("%x", o[:])
}

// ToOID convert a string into an OID
func ToOID(oid string) (OID, error) {
	var o OID

	by, err := hex.DecodeString(oid)
	if err != nil {
		return o, err
	}

	copy(o[:], by)
	return o, nil
}
