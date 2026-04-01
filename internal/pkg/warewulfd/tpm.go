package warewulfd

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-attestation/attest"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	nodedb "github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/tpm"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// Handler for receiving tpm quote. As nothing is known about
// the node at this point, we just store what we are getting
func TPMReceive(w http.ResponseWriter, req *http.Request) {
	wwlog.Debug("Requested URL: %s", req.URL.String())

	wwidRecv := req.URL.Query().Get("wwid")
	if wwidRecv == "" {
		wwlog.Error("TPM receive: wwid parameter missing")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate node exists
	nodes, err := nodedb.New()
	if err != nil {
		wwlog.Error("Failed to load node configuration: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Check if the node exists by ID, IP or HW address
	node, err := nodes.GetNodeOnly(wwidRecv)
	if err != nil {
		if node, err = nodes.FindByIpaddr(wwidRecv); err != nil {
			if node, err = nodes.FindByHwaddr(wwidRecv); err != nil {
				wwlog.Error("Node not found: %s", wwidRecv)
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		wwlog.Error("Failed to read request body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var upload tpm.TpmUpload
	err = json.Unmarshal(body, &upload)
	if err != nil {
		wwlog.Error("Failed to unmarshal JSON quote: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	upload.ID = wwidRecv

	tpmStore, err := NewTPMLogStore(node.GetId())
	if err != nil {
		wwlog.Error("Failed to access TPM store: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tpmStore.Save(upload); err != nil {
		wwlog.Error("Failed to write TPM quote: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	wwlog.Info("Stored TPM quote for node %s (Manufacturer: %s)", upload.ID, upload.TpmData.GetManufacturer())
	w.WriteHeader(http.StatusOK)
}

// Send the challenge to the node. Challenge itself can only be
// decrypted by the TPM of the node.
func TPMChallengeSend(w http.ResponseWriter, req *http.Request) {
	wwlog.Debug("Requested URL: %s", req.URL.String())

	wwidRecv := req.URL.Query().Get("wwid")
	if wwidRecv == "" {
		wwlog.Error("TPM challenge send: wwid parameter missing")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodes, err := nodedb.New()
	if err != nil {
		wwlog.Error("Failed to load node configuration: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	node, err := nodes.GetNodeOnly(wwidRecv)
	if err != nil {
		if node, err = nodes.FindByIpaddr(wwidRecv); err != nil {
			if node, err = nodes.FindByHwaddr(wwidRecv); err != nil {
				wwlog.Error("Node not found: %s", wwidRecv)
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
	}

	conf := warewulfconf.Get()
	tpmPath := filepath.Join(conf.Paths.OverlayProvisiondir(), node.GetId(), "tpm.json")

	if !util.IsFile(tpmPath) {
		wwlog.Error("No TPM quote found for node %s", node.GetId())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data, err := os.ReadFile(tpmPath)
	if err != nil {
		wwlog.Error("Failed to read TPM quote: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var existingQuote tpm.Quote
	err = json.Unmarshal(data, &existingQuote)
	if err != nil {
		wwlog.Error("Failed to unmarshal TPM quote: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if verify, err := existingQuote.Verify(); !verify && err != nil {
		wwlog.Error("Failed to verify TPM quote: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if verify, err := existingQuote.VerifyEventLog(); !verify && err != nil {
		wwlog.Error("Failed to reply TPM log: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := existingQuote.VerifyGrubBinary(); err != nil {
		wwlog.Error("grub events on server do no match events in TPM: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ekPubBytes, err := base64.StdEncoding.DecodeString(existingQuote.Current.EKPub)
	if err != nil {
		wwlog.Error("Failed to decode EKPub for node %s: %s", node.GetId(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	akPubBytes, err := base64.StdEncoding.DecodeString(existingQuote.Current.AKPub)
	if err != nil {
		wwlog.Error("Failed to decode AKPub for node %s: %s", node.GetId(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ekPub, err := x509.ParsePKIXPublicKey(ekPubBytes)
	if err != nil {
		wwlog.Error("Failed to parse EK public key for node %s: %s", node.GetId(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var createData []byte
	var createAttestation []byte
	var createSignature []byte
	if existingQuote.Current.CreateData != "" {
		createData, err = base64.StdEncoding.DecodeString(existingQuote.Current.CreateData)
		if err != nil {
			wwlog.Error("Failed to decode CreateData for node %s: %s", node.GetId(), err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if existingQuote.Current.CreateAttestation != "" {
		createAttestation, err = base64.StdEncoding.DecodeString(existingQuote.Current.CreateAttestation)
		if err != nil {
			wwlog.Error("Failed to decode CreateAttestation for node %s: %s", node.GetId(), err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if existingQuote.Current.CreateSignature != "" {
		createSignature, err = base64.StdEncoding.DecodeString(existingQuote.Current.CreateSignature)
		if err != nil {
			wwlog.Error("Failed to decode CreateSignature for node %s: %s", node.GetId(), err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	akAttestParams := attest.AttestationParameters{
		Public:            akPubBytes,
		CreateData:        createData,
		CreateAttestation: createAttestation,
		CreateSignature:   createSignature,
	}

	activationParams := attest.ActivationParameters{
		EK: ekPub,
		AK: akAttestParams,
	}

	secret, encryptedCredential, err := activationParams.Generate()
	if err != nil {
		wwlog.Error("Error generating Credential Activation Challenge for node %s: %s", node.GetId(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	wwlog.Debug("secret %s: %x", node.GetId(), secret)

	newChallenge := tpm.Challenge{
		EncryptedCredential: *encryptedCredential,
		Secret:              secret,
		ID:                  node.GetId(),
	}

	tpmStore, err := NewTPMLogStore(node.GetId())
	if err != nil {
		wwlog.Error("Failed to access TPM store: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tpmStore.SaveChallenge(newChallenge); err != nil {
		wwlog.Error("Failed to write TPM challenge: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newChallenge.EncryptedCredential)

	wwlog.Info("Sent TPM challenge for node %s", node.GetId())
}

type TPMLogStore struct {
	path string
}

func NewTPMLogStore(nodeId string) (*TPMLogStore, error) {
	conf := warewulfconf.Get()
	path := filepath.Join(conf.Paths.OverlayProvisiondir(), nodeId, "tpm.json")

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	if !util.IsFile(path) {
		out, err := json.MarshalIndent(tpm.Quote{Modified: time.Now()}, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, out, 0644); err != nil {
			return nil, err
		}
	}

	return &TPMLogStore{
		path: path,
	}, nil
}

func (s *TPMLogStore) Save(upload tpm.TpmUpload) error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return fmt.Errorf("couldn't access storage for quote: %s", err)
	}
	var quote tpm.Quote
	_ = json.Unmarshal(data, &quote)

	if !quote.Current.HasQuote() {
		quote.Current = upload.TpmData
	} else if !quote.Current.Equal(&upload.TpmData) {
		quote.New = upload.TpmData
	} else {
		quote.Current = upload.TpmData
	}
	quote.EventLog = upload.EventLog
	quote.ID = upload.ID
	quote.Modified = time.Now()

	out, err := json.MarshalIndent(quote, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, out, 0644)
}

func (s *TPMLogStore) SaveChallenge(challenge tpm.Challenge) error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return fmt.Errorf("couldn't access storage for challenge: %s", err)
	}
	var quote tpm.Quote
	if err := json.Unmarshal(data, &quote); err != nil {
		return err
	}
	quote.Challenge = &challenge
	quote.Modified = time.Now()

	out, err := json.MarshalIndent(quote, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, out, 0644)
}

func (s *TPMLogStore) SetFilename(filename string) {
	s.path = filename
}

func (s *TPMLogStore) ClearLogs() error {
	if !util.IsFile(s.path) {
		return nil
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var quote tpm.Quote
	err = json.Unmarshal(data, &quote)
	if err != nil {
		return err
	}
	quote.SentLog = []tpm.FileLog{}
	quote.Modified = time.Now()
	out, err := json.MarshalIndent(quote, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, out, 0644)
}

func (s *TPMLogStore) Update(filename, checksum string) error {
	if checksum == "" {
		fileBytes, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("couldn't access file for quote: %s", err)
		}
		sum := sha256.Sum256(fileBytes)
		checksum = fmt.Sprintf("%x", sum)
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return fmt.Errorf("couldn't access storage for quote: %s", err)
	}
	quote := tpm.Quote{}
	err = json.Unmarshal(data, &quote)
	if err != nil {
		return err
	}

	for _, entry := range quote.SentLog {
		if entry.Filename == filename && entry.Checksum == checksum {
			return nil
		}
	}

	quote.SentLog = append(quote.SentLog, tpm.FileLog{
		Filename: filename,
		Checksum: checksum,
	})
	quote.Modified = time.Now()

	out, err := json.MarshalIndent(quote, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, out, 0644)
}

func (s *TPMLogStore) GetSecret() string {
	if !util.IsFile(s.path) {
		return ""
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return ""
	}
	var quote tpm.Quote
	err = json.Unmarshal(data, &quote)
	if err != nil {
		return ""
	}
	return string(quote.Challenge.Secret)
}
