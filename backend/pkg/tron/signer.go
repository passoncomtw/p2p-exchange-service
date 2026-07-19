package tron

import (
	"encoding/hex"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// SignRawDataHex signs a Tron raw_data_hex string with the given secp256k1 private key.
// Returns the 65-byte hex signature in Tron format: R(32) + S(32) + V(1), V = recovery_id (0 or 1).
func SignRawDataHex(rawDataHex, privateKeyHex string) (string, error) {
	privKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key hex: %w", err)
	}
	if len(privKeyBytes) != 32 {
		return "", fmt.Errorf("private key must be 32 bytes, got %d", len(privKeyBytes))
	}

	msgHash, err := HashRawData(rawDataHex)
	if err != nil {
		return "", fmt.Errorf("hash raw data: %w", err)
	}

	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)

	// SignCompact returns: [flag(1)][R(32)][S(32)]
	// flag = 31 + recovery_id for compressed key, 27 + recovery_id for uncompressed
	compactSig := ecdsa.SignCompact(privKey, msgHash, true)
	if len(compactSig) != 65 {
		return "", fmt.Errorf("unexpected compact signature length: %d", len(compactSig))
	}

	recoveryID := compactSig[0] - 31 // compressed key offset

	tronSig := make([]byte, 65)
	copy(tronSig[0:32], compactSig[1:33])   // R
	copy(tronSig[32:64], compactSig[33:65]) // S
	tronSig[64] = recoveryID

	return hex.EncodeToString(tronSig), nil
}
