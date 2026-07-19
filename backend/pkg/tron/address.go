package tron

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
)

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// TronBase58ToBytes decodes a Tron Base58Check address to 21 bytes (0x41 + 20-byte raw address).
// Validates the 4-byte SHA256d checksum.
func TronBase58ToBytes(addr string) ([]byte, error) {
	if len(addr) == 0 {
		return nil, errors.New("empty address")
	}

	// Build lookup table
	var lookup [128]int
	for i := range lookup {
		lookup[i] = -1
	}
	for i, c := range base58Alphabet {
		lookup[c] = i
	}

	result := new(big.Int)
	base := big.NewInt(58)
	for _, c := range addr {
		if int(c) >= 128 || lookup[c] == -1 {
			return nil, fmt.Errorf("invalid base58 character: %c", c)
		}
		result.Mul(result, base)
		result.Add(result, big.NewInt(int64(lookup[c])))
	}

	// Convert to 25-byte slice: 1 version byte + 20 data bytes + 4 checksum bytes
	decoded := result.Bytes()
	if len(decoded) < 25 {
		padded := make([]byte, 25)
		copy(padded[25-len(decoded):], decoded)
		decoded = padded
	}

	payload := decoded[:21]
	checksum := decoded[21:25]

	h1 := sha256.Sum256(payload)
	h2 := sha256.Sum256(h1[:])
	if h2[0] != checksum[0] || h2[1] != checksum[1] || h2[2] != checksum[2] || h2[3] != checksum[3] {
		return nil, errors.New("invalid checksum")
	}
	if payload[0] != 0x41 {
		return nil, fmt.Errorf("unexpected version byte: 0x%02x", payload[0])
	}

	return payload, nil // 21 bytes: 0x41 + 20-byte address
}
