package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"

	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hpcng/warewulf/internal/pkg/api/apiconfig"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"

	gw "github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"

	"path"
)

func run() error {

	log.Println("test0")

	conf := warewulfconf.New()
	// Read the config file.
	config, err := apiconfig.NewClientServer(path.Join(conf.Paths.Sysconfdir, "warewulf/wwapird.conf"))
	if err != nil {
		glog.Fatalf("Failed to read config file, err: %v", err)
	}
	grpcServerEndpoint := fmt.Sprintf("%s:%v", config.ClientApiConfig.Server, config.ClientApiConfig.Port)
	httpServerEndpoint := fmt.Sprintf(":%v", config.ServerApiConfig.Port)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint (we are the client)
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()

	var opts []grpc.DialOption
	if config.ClientTlsConfig.Enabled {

		// Load the client cert and its key
		clientCert, err := tls.LoadX509KeyPair(config.ClientTlsConfig.Cert, config.ClientTlsConfig.Key)
		if err != nil {
			log.Fatalf("Failed to load client cert and key. %s.", err)
		}

		// Load the CA cert.
		var cacert []byte
		cacert, err = os.ReadFile(config.ClientTlsConfig.CaCert)
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
		opts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}

	err = gw.RegisterWWApiHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts)
	if err != nil {
		return err
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	if config.ServerTlsConfig.Enabled {

		return http.ListenAndServeTLS(
			httpServerEndpoint,
			config.ServerTlsConfig.ConcatCert,
			config.ServerTlsConfig.Key,
			mux)
	}

	// Insecure
	return http.ListenAndServe(httpServerEndpoint, mux)
}

func main() {
	flag.Parse() // Pretty sure glog wants this.
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
