package tpm

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/google/go-tpm/legacy/tpm2"
)

// Quote struct to hold EK certificate and attestation data
type Quote struct {
	EKCert    string            `json:"ek_cert"`
	EKPub     string            `json:"ek_pub"`
	AKPub     string            `json:"ak_pub"`
	Quote     string            `json:"quote"`
	Signature *tpm2.Signature   `json:"signature"`
	PCRs      map[string]string `json:"pcrs"`
	Nonce     string            `json:"nonce"`
	EventLog  string            `json:"eventlog,omitempty"`
}

func VerifyQuote(quote *Quote) error {
	// 1. Parse AK Public Key
	akPubBytes, err := base64.StdEncoding.DecodeString(quote.AKPub)
	if err != nil {
		return fmt.Errorf("decoding AKPub: %v", err)
	}

	// Decode TPM public key
	akPubTPM, err := tpm2.DecodePublic(akPubBytes)
	if err != nil {
		return fmt.Errorf("decoding TPM public key: %v", err)
	}

	if akPubTPM.Type != tpm2.AlgECC {
		return fmt.Errorf("AK is not ECC")
	}

	var curve elliptic.Curve
	switch akPubTPM.ECCParameters.CurveID {
	case tpm2.CurveNISTP256:
		curve = elliptic.P256()
	default:
		return fmt.Errorf("unsupported curve: %v", akPubTPM.ECCParameters.CurveID)
	}

	ecdsaPub := &ecdsa.PublicKey{
		Curve: curve,
		X:     akPubTPM.ECCParameters.Point.X(),
		Y:     akPubTPM.ECCParameters.Point.Y(),
	}

	// 2. Decode Attestation Data (Quote)
	quoteBytes, err := base64.StdEncoding.DecodeString(quote.Quote)
	if err != nil {
		return fmt.Errorf("decoding quote: %v", err)
	}

	attest, err := tpm2.DecodeAttestationData(quoteBytes)
	if err != nil {
		return fmt.Errorf("decoding attestation data: %v", err)
	}

	// 3. Verify Signature
	if quote.Signature == nil {
		return fmt.Errorf("missing signature")
	}
	if quote.Signature.ECC == nil {
		return fmt.Errorf("signature is not ECC")
	}

	// Hash the quote data
	hash := sha256.Sum256(quoteBytes)

	if !ecdsa.Verify(ecdsaPub, hash[:], quote.Signature.ECC.R, quote.Signature.ECC.S) {
		return fmt.Errorf("signature verification failed")
	}

	// 4. Verify Nonce (Qualification)
	nonceBytes, err := base64.StdEncoding.DecodeString(quote.Nonce)
	if err != nil {
		return fmt.Errorf("decoding nonce: %v", err)
	}

	if !bytes.Equal(attest.ExtraData, nonceBytes) {
		return fmt.Errorf("nonce mismatch: expected %x, got %x", nonceBytes, attest.ExtraData)
	}

	// 5. Verify Magic
	if attest.Magic != 0xff544347 { // TPM_GENERATED_VALUE
		return fmt.Errorf("invalid magic: %x", attest.Magic)
	}

	// 6. Verify Type
	if attest.Type != tpm2.TagAttestQuote {
		return fmt.Errorf("invalid type: %v", attest.Type)
	}

	return nil
}