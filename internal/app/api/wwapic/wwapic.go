package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/warewulf/warewulf/internal/pkg/api/apiconfig"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// wwapic is intended as a sample wwapi client.

func main() {
	log.Println("Client running")
	conf := warewulfconf.Get()

	// Read the config file.
	config, err := apiconfig.NewClient(path.Join(conf.Paths.Sysconfdir, "warewulf/wwapic.conf"))
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	var opts []grpc.DialOption
	if config.TlsConfig.Enabled {

		// Load the client cert and its key
		clientCert, err := tls.LoadX509KeyPair(config.TlsConfig.Cert, config.TlsConfig.Key)
		if err != nil {
			log.Fatalf("Failed to load client cert and key. %s.", err)
		}

		// Load the CA cert.
		var cacert []byte
		cacert, err = os.ReadFile(config.TlsConfig.CaCert)
		if err != nil {
			log.Fatalf("Failed to load cacert. err: %s\n", err)
		}

		// Put the CA cert into the cert pool.
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(cacert) {
			log.Fatalf("Failed to append CA cert to certificate pool. %s.", err)
		}

		// Create the TLS configuration
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
			MinVersion:   tls.VersionTLS13,
			MaxVersion:   tls.VersionTLS13,
		}

		// Create TLS credentials from the TLS configuration
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.DialOption(grpc.WithTransportCredentials(creds)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%v", config.ApiConfig.Server, config.ApiConfig.Port), opts...)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	client := wwapiv1.NewWWApiClient(conn)
	request := &emptypb.Empty{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.Version(ctx, request)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Version Response: %v\n", response)
}
