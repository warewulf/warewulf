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
	"github.com/warewulf/warewulf/internal/pkg/node"
	nodedb "github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/tpm"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type templateVars struct {
	Message       string
	WaitTime      string
	Hostname      string
	Fqdn          string
	Id            string
	Cluster       string
	ImageName     string
	Ipxe          string
	Hwaddr        string
	Ipaddr        string
	Ipaddr6       string
	Port          string
	Authority     string
	KernelArgs    string
	KernelVersion string
	Root          string
	TLS           bool
	Tags          map[string]string
	NetDevs       map[string]*node.NetDev
}

func HandleProvision(w http.ResponseWriter, req *http.Request) {
	// Parse just enough to determine the stage
	rinfo, err := parseRequest(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		wwlog.ErrorExc(err, "Bad status")
		return
	}

	// Dispatch to the appropriate stage handler
	var handler http.HandlerFunc
	switch rinfo.stage {
	case "ipxe":
		handler = HandleIpxe
	case "kernel":
		handler = HandleKernel
	case "image":
		handler = HandleImage
	case "system":
		handler = HandleSystemOverlay
	case "runtime":
		handler = HandleRuntimeOverlay
	case "grub":
		handler = HandleGrub
	case "initramfs":
		handler = HandleInitramfs
	default:
		w.WriteHeader(http.StatusBadRequest)
		wwlog.Error("Unknown stage: %s", rinfo.stage)
		return
	}
	handler(w, req)
}
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

	var newQuote tpm.Quote
	err = json.Unmarshal(body, &newQuote)
	if err != nil {
		wwlog.Error("Failed to unmarshal JSON quote: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newQuote.ID = wwidRecv
	newQuote.Modified = time.Now()

	tpmStore, err := NewTPMLogStore(node.GetId())
	if err != nil {
		wwlog.Error("Failed to access TPM store: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tpmStore.Save(newQuote); err != nil {
		wwlog.Error("Failed to write TPM quote: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	wwlog.Info("Stored TPM quote for node %s", newQuote.ID)
	w.WriteHeader(http.StatusOK)
}

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

	ekPubBytes, err := base64.StdEncoding.DecodeString(existingQuote.EKPub)
	if err != nil {
		wwlog.Error("Failed to decode EKPub for node %s: %s", node.GetId(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	akPubBytes, err := base64.StdEncoding.DecodeString(existingQuote.AKPub)
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
	if existingQuote.CreateData != "" {
		createData, err = base64.StdEncoding.DecodeString(existingQuote.CreateData)
		if err != nil {
			wwlog.Error("Failed to decode CreateData for node %s: %s", node.GetId(), err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if existingQuote.CreateAttestation != "" {
		createAttestation, err = base64.StdEncoding.DecodeString(existingQuote.CreateAttestation)
		if err != nil {
			wwlog.Error("Failed to decode CreateAttestation for node %s: %s", node.GetId(), err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if existingQuote.CreateSignature != "" {
		createSignature, err = base64.StdEncoding.DecodeString(existingQuote.CreateSignature)
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

	out, err := json.MarshalIndent(newChallenge, "", "  ")
	if err != nil {
		wwlog.Error("Failed to marshal TPM challenge: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	challengePath := filepath.Join(conf.Paths.OverlayProvisiondir(), node.GetId(), "tpm_challenge.json")
	err = os.WriteFile(challengePath, out, 0644)
	if err != nil {
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
		out, err := json.MarshalIndent(tpm.Quote{}, "", "  ")
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

func (s *TPMLogStore) Save(newQuote tpm.Quote) error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return fmt.Errorf("couldn't access storage for quote: %s", err)
	}
	var existingQuote tpm.Quote
	if err := json.Unmarshal(data, &existingQuote); err == nil {
		newQuote.SentLog = existingQuote.SentLog
	}

	out, err := json.MarshalIndent(newQuote, "", "  ")
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

	out, err := json.MarshalIndent(quote, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, out, 0644)
}
