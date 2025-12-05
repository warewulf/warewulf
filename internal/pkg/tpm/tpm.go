package tpm

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/go-attestation/attest"
)

// Quote struct to hold EK certificate and attestation data
type Quote struct {
	EKCert    string            `json:"ek_cert"`
	EKPub     string            `json:"ek_pub"`
	AKPub     string            `json:"ak_pub"`
	Quote     string            `json:"quote"`
	Signature string            `json:"signature"`
	PCRs      map[string]string `json:"pcrs"`
	Nonce     string            `json:"nonce"`
	EventLog  string            `json:"eventlog,omitempty"`
	Name      string            `json:"name"`
	ID        string            `json:"id"`
}

var (
	ErrDecodeAKPub     = errors.New("decoding AKPub failed")
	ErrParseAKPub      = errors.New("parsing TPM public key failed")
	ErrDecodeQuote     = errors.New("decoding quote failed")
	ErrDecodeSignature = errors.New("decoding signature failed")
	ErrDecodeNonce     = errors.New("decoding nonce failed")
	ErrQuoteVerify     = errors.New("quote verification failed")
	ErrNoEventLog      = errors.New("no event log present")
	ErrEventLogVerify  = errors.New("event log verification failed")
)

func (quote *Quote) Verify() (bool, error) {
	// 1. Parse AK Public Key
	akPubBytes, err := base64.StdEncoding.DecodeString(quote.AKPub)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrDecodeAKPub, err)
	}

	akPub, err := x509.ParsePKIXPublicKey(akPubBytes)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrParseAKPub, err)
	}

	// 2. Decode Quote
	quoteBytes, err := base64.StdEncoding.DecodeString(quote.Quote)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrDecodeQuote, err)
	}

	sigBytes, err := base64.StdEncoding.DecodeString(quote.Signature)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrDecodeSignature, err)
	}

	nonceBytes, err := base64.StdEncoding.DecodeString(quote.Nonce)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrDecodeNonce, err)
	}

	// Construct go-attestation Quote object
	q := attest.Quote{
		Version:   attest.TPMVersion20,
		Quote:     quoteBytes,
		Signature: sigBytes,
	}

	// Reconstruct PCRs
	var pcrs []attest.PCR
	for idxStr, digestHex := range quote.PCRs {
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			continue
		}
		digest, err := hex.DecodeString(digestHex)
		if err != nil {
			continue
		}
		pcrs = append(pcrs, attest.PCR{
			Index:     idx,
			Digest:    digest,
			DigestAlg: crypto.SHA256,
		})
	}

	verifier := &attest.AKPublic{
		Public: akPub,
		Hash:   crypto.SHA256,
	}
	if err := verifier.Verify(q, pcrs, nonceBytes); err != nil {
		return false, fmt.Errorf("%w: %v", ErrQuoteVerify, err)
	}

	return true, nil
}

func (quote *Quote) VerifyEventLog() (bool, error) {
	if quote.EventLog == "" {
		return false, ErrNoEventLog
	}

	logBytes, err := base64.StdEncoding.DecodeString(quote.EventLog)
	if err != nil {
		return false, fmt.Errorf("decoding event log: %v", err)
	}

	el, err := attest.ParseEventLog(logBytes)
	if err != nil {
		return false, fmt.Errorf("parsing event log: %v", err)
	}

	// Reconstruct PCRs
	var pcrs []attest.PCR
	for idxStr, digestHex := range quote.PCRs {
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			continue
		}
		digest, err := hex.DecodeString(digestHex)
		if err != nil {
			continue
		}
		pcrs = append(pcrs, attest.PCR{
			Index:     idx,
			Digest:    digest,
			DigestAlg: crypto.SHA256,
		})
	}

	if _, err := el.Verify(pcrs); err != nil {
		var replayErr attest.ReplayError
		if errors.As(err, &replayErr) {
			return false, fmt.Errorf("%w: invalid PCRs %v", ErrEventLogVerify, replayErr.InvalidPCRs)
		}
		return false, fmt.Errorf("%w: %v", ErrEventLogVerify, err)
	}

	return true, nil
}
