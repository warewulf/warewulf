package tls

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/configure"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	conf := config.Get()
	keystore := path.Join(conf.Paths.Sysconfdir, "warewulf", "keys")

	if err := os.MkdirAll(keystore, 0755); err != nil {
		return fmt.Errorf("could not create keystore directory: %w", err)
	}

	keyFile := path.Join(keystore, "warewulf.key")
	certFile := path.Join(keystore, "warewulf.crt")

	if importPath != "" {
		info, err := os.Stat(importPath)
		if err != nil {
			return fmt.Errorf("could not access import path: %w", err)
		}

		var sourceKey, sourceCert string
		if info.IsDir() {
			sourceKey = path.Join(importPath, "warewulf.key")
			sourceCert = path.Join(importPath, "warewulf.crt")
		} else {
			return fmt.Errorf("import path must be a directory containing warewulf.key and warewulf.crt")
		}

		if err := util.CopyFile(sourceKey, keyFile); err != nil {
			return fmt.Errorf("failed to import key: %w", err)
		}
		if err := util.CopyFile(sourceCert, certFile); err != nil {
			return fmt.Errorf("failed to import cert: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Imported keys from %s\n", importPath)

		if err := configure.WAREWULFD(); err != nil {
			return fmt.Errorf("failed to restart warewulfd: %w", err)
		}

		return nil
	}

	if exportPath != "" {
		if err := os.MkdirAll(exportPath, 0755); err != nil {
			return fmt.Errorf("could not create export directory: %w", err)
		}
		if err := util.CopyFile(keyFile, path.Join(exportPath, "warewulf.key")); err != nil {
			return fmt.Errorf("failed to export key: %w", err)
		}
		if err := util.CopyFile(certFile, path.Join(exportPath, "warewulf.crt")); err != nil {
			return fmt.Errorf("failed to export cert: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Exported keys to %s\n", exportPath)
		return nil
	}

	if util.IsFile(keyFile) && util.IsFile(certFile) && !force {
		if create {
			fmt.Fprintf(cmd.OutOrStdout(), "Keys already exist in %s\n", keystore)
		}
	} else {
		if create {
			if err := configure.GenTLSKeys(); err != nil {
				return err
			}
			if err := configure.WAREWULFD(); err != nil {
				return fmt.Errorf("failed to restart warewulfd: %w", err)
			}
		} else {
			return fmt.Errorf("keys not found in: %s", keystore)
		}
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "Private Key:\t%s\n", keyFile)
	fmt.Fprintf(w, "Certificate:\t%s\n", certFile)

	keyBytes, err := os.ReadFile(keyFile)
	if err == nil {
		block, _ := pem.Decode(keyBytes)
		if block != nil {
			if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
				fmt.Fprintf(w, "Key Size:\t%d bits\n", key.N.BitLen())
			}
		}
	}

	certBytes, err := os.ReadFile(certFile)
	if err == nil {
		block, _ := pem.Decode(certBytes)
		if block != nil {
			if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
				fmt.Fprintf(w, "Issuer:\t%s\n", cert.Issuer)
				fmt.Fprintf(w, "Subject:\t%s\n", cert.Subject)
				fmt.Fprintf(w, "Valid From:\t%s\n", cert.NotBefore)
				fmt.Fprintf(w, "Valid Until:\t%s\n", cert.NotAfter)
				fmt.Fprintf(w, "Serial Nr:\t%s\n", cert.SerialNumber)

				var ipStrings []string
				for _, ip := range cert.IPAddresses {
					ipStrings = append(ipStrings, ip.String())
				}
				fmt.Fprintf(w, "IP Addresses:\t%s\n", strings.Join(ipStrings, ", "))
				fmt.Fprintf(w, "DNS Names:\t%s\n", strings.Join(cert.DNSNames, ", "))
			}
		}
	}
	w.Flush()

	return nil
}
