package tpm

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-attestation/attest"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// FileLog struct to hold checksums of files sent to the node
type FileLog struct {
	Filename string `json:"filename" yaml:"filename"`
	Checksum string `json:"checksum" yaml:"checksum"`
}

// Quote struct to hold EK certificate and attestation data
type Quote struct {
	EKCert    string            `json:"ek_cert" yaml:"ek_cert"`
	EKPub     string            `json:"ek_pub" yaml:"ek_pub"`
	AKPub     string            `json:"ak_pub" yaml:"ak_pub"`
	Quote     string            `json:"quote" yaml:"quote"`
	Signature string            `json:"signature" yaml:"signature"`
	PCRs      map[string]string `json:"pcrs" yaml:"pcrs"`
	Nonce     string            `json:"nonce" yaml:"nonce"`

	CreateData        string `json:"create_data,omitempty" yaml:"create_data,omitempty"`
	CreateAttestation string `json:"create_attestation,omitempty" yaml:"create_attestation,omitempty"`
	CreateSignature   string `json:"create_signature,omitempty" yaml:"create_signature,omitempty"`

	EventLog string    `json:"eventlog,omitempty" yaml:"eventlog,omitempty"`
	Token    string    `json:"token,omitempty" yaml:"token,omitempty"`
	ID       string    `json:"id" yaml:"id"`
	Modified time.Time `json:"modified" yaml:"modified"`
	SentLog  []FileLog `json:"sentlogs,omitempty" yaml:"logs,omitempty"`
}

// Challenge struct to hold encrypted credentials and secrets for TPM challenges

type Challenge struct {
	EncryptedCredential attest.EncryptedCredential `json:"encrypted_credential" yaml:"encrypted_credential"`

	Secret []byte `json:"secret" yaml:"secret"`

	ID string `json:"id" yaml:"id"`
}

// TPMConfig struct to hold both Quotes and Challenges

type TPMConfig struct {
	Quotes map[string]Quote `yaml:"quotes"`

	Challenges map[string]Challenge `yaml:"challenges"`
}

var (
	ErrDecodeAKPub = errors.New("decoding AKPub failed")

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

	akPubObj, err := attest.ParseAKPublic(akPubBytes)
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
		wwlog.Verbose("pcr[%d]: %x", idx, digest)
		pcrs = append(pcrs, attest.PCR{
			Index:     idx,
			Digest:    digest,
			DigestAlg: crypto.SHA256,
		})
	}

	verifier := &attest.AKPublic{
		Public: akPubObj.Public,
		Hash:   crypto.SHA256,
	}
	wwlog.Verbose("Quote: %x", q.Quote)
	wwlog.Verbose("Signature: %x", q.Signature)
	wwlog.Verbose("nonceBytes: %x", nonceBytes)
	wwlog.Verbose("akPub: %x", akPubObj.Public)
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

func (quote *Quote) VerifyGrubBinary() error {
	sentReceived := []FileLog{}
	if quote.EventLog != "" {
		logBytes, err := base64.StdEncoding.DecodeString(quote.EventLog)
		if err == nil {
			el, err := attest.ParseEventLog(logBytes)
			if err == nil {
				events := el.Events(attest.HashSHA256)
				for _, event := range events {
					if event.Index != 9 {
						continue
					}
					found := false
					for _, log := range quote.SentLog {
						sum, err := hex.DecodeString(log.Checksum)
						if err == nil && bytes.Equal(sum, event.Digest) {
							found = true
							sentReceived = append(sentReceived, log)
							break
						}
					}
					if !found {
						wwlog.Warn("Event not found in tpm.json: Digest=%x Data=%s", event.Digest, FormatEventData(event))
					}
				}
			}
		}
	} else {
		sentReceived = quote.SentLog
	}

	if quote.PCRs == nil {
		return fmt.Errorf("no PCRs in quote")
	}

	// We expect PCR9
	pcr9Hex, ok := quote.PCRs["9"]
	if !ok {
		return fmt.Errorf("PCR9 not present in quote")
	}

	pcr9, err := hex.DecodeString(pcr9Hex)
	if err != nil {
		return fmt.Errorf("failed to decode PCR9: %v", err)
	}

	// Start with empty SHA256 (32 bytes of zeros)
	pcr := make([]byte, 32)

	for _, log := range sentReceived {
		sum, err := hex.DecodeString(log.Checksum)
		if err != nil {
			return fmt.Errorf("failed to decode checksum for %s: %v", log.Filename, err)
		}
		// TPM Extend: NewPCR = SHA256(OldPCR || DataHash)
		hasher := sha256.New()
		hasher.Write(pcr)
		hasher.Write(sum)
		pcr = hasher.Sum(nil)
	}

	if !bytes.Equal(pcr, pcr9) {
		return fmt.Errorf("PCR9 mismatch: expected %x, got %x", pcr, pcr9)
	}

	return nil
}
