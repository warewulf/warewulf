package tpm

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyGrubBinary(t *testing.T) {
	// 1. Create dummy checksums
	sum1 := sha256.Sum256([]byte("file1"))
	sum2 := sha256.Sum256([]byte("file2"))

	sum1Hex := hex.EncodeToString(sum1[:])
	sum2Hex := hex.EncodeToString(sum2[:])

	// 2. Calculate expected PCR9
	pcr := make([]byte, 32)

	// Extend sum1
	h := sha256.New()
	h.Write(pcr)
	h.Write(sum1[:])
	pcr = h.Sum(nil)

	// Extend sum2
	h.Reset()
	h.Write(pcr)
	h.Write(sum2[:])
	pcr = h.Sum(nil)

	pcrHex := hex.EncodeToString(pcr)

	quote := Quote{
		SentLog: []FileLog{
			{Filename: "file1", Checksum: sum1Hex},
			{Filename: "file2", Checksum: sum2Hex},
		},
		PCRs: map[string]string{
			"9": pcrHex,
		},
	}

	if err := quote.VerifyGrubBinary(); err != nil {
		t.Errorf("VerifyGrubBinary failed: %v", err)
	}

	// Test failure
	quote.PCRs["9"] = "0000000000000000000000000000000000000000000000000000000000000000"
	if err := quote.VerifyGrubBinary(); err == nil {
		t.Error("VerifyGrubBinary should have failed with wrong PCR")
	}
}
