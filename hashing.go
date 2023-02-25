package main

import (
	"crypto/sha1"
	"math/big"
)

func hash(ip string) []byte {

	h := sha1.New()
	h.Write([]byte(ip))
	return h.Sum(nil)

} //end of method

/** Source: https://github.com/fastfn/dendrite/blob/master/ */
func powerOffset(id []byte, exp int, mod int) int64 {

	// Copy the existing slice
	off := make([]byte, len(id))
	copy(off, id)

	// Convert the ID to a bigint
	idInt := big.Int{}
	idInt.SetBytes(id)

	// Get the offset
	two := big.NewInt(2)
	offset := big.Int{}
	offset.Exp(two, big.NewInt(int64(exp)), nil)

	// Sum
	sum := big.Int{}
	sum.Add(&idInt, &offset)

	// Get the ceiling
	ceil := big.Int{}
	ceil.Exp(two, big.NewInt(int64(mod)), nil)

	// Apply the mod
	idInt.Mod(&sum, &ceil)

	return idInt.Int64()

} //end of method

func consistentHashing(id string) int {
	h := hash(id)
	i := powerOffset(h, 1, int(cf.RingSize))
	return int(i)
}
