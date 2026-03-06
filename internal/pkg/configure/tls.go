package configure

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path"
	"time"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// TLS ensures TLS keys exist if TLS is enabled.
// If force is true, regenerates even if keys already exist.
// Returns true if new keys were generated.
func TLS(force bool) (bool, error) {
	conf := warewulfconf.Get()
	if !conf.Warewulf.TLSEnabled() {
		return false, nil
	}

	keystore := path.Join(conf.Paths.Sysconfdir, "warewulf", "tls")
	keyFile := path.Join(keystore, "warewulf.key")
	certFile := path.Join(keystore, "warewulf.crt")

	if !force && util.IsFile(keyFile) && util.IsFile(certFile) {
		wwlog.Info("TLS keys already exist in %s", keystore)
		return false, nil
	}

	if err := GenTLSKeys(); err != nil {
		return false, err
	}
	wwlog.Info("TLS keys generated in %s", keystore)
	return true, nil
}

// GenTLSKeys generates new TLS keys and certificate unconditionally.
func GenTLSKeys() error {
	conf := warewulfconf.Get()
	keystore := path.Join(conf.Paths.Sysconfdir, "warewulf", "tls")

	if err := os.MkdirAll(keystore, 0755); err != nil {
		return fmt.Errorf("could not create keystore directory: %w", err)
	}

	keyFile := path.Join(keystore, "warewulf.key")
	certFile := path.Join(keystore, "warewulf.crt")
	wwlog.Verbose("Generating new x509 keys in %s", keystore)
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate rsa key: %w", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "Warewulf Server",
			Organization: []string{"Warewulf"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365 * 10),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if ip := net.ParseIP(conf.Ipaddr); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	}
	if ip := net.ParseIP(conf.Ipaddr6); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	}
	if conf.Fqdn != "" {
		template.DNSNames = append(template.DNSNames, conf.Fqdn)
	}
	template.DNSNames = append(template.DNSNames, "warewulf")

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open key file for writing: %w", err)
	}
	defer keyOut.Close()
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return fmt.Errorf("failed to write data to key file: %w", err)
	}

	certOut, err := os.OpenFile(certFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open cert file for writing: %w", err)
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write data to cert file: %w", err)
	}

	return nil
}
