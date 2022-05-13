package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/api/apiconfig"
	"github.com/hpcng/warewulf/internal/pkg/api/container"
	"github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type apiServer struct {
	wwapiv1.UnimplementedWWApiServer
}

var apiPrefix string
var apiVersion string

func main() {
	log.Println("Server running")

	// Read the config file.
	config, err := apiconfig.NewServer(path.Join(buildconfig.SYSCONFDIR(), "warewulf/wwapid.conf"))
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	// Pull out config variables.
	apiPrefix = config.ApiConfig.Prefix
	apiVersion = config.ApiConfig.Version
	servicePort := config.ApiConfig.Port
	portString := fmt.Sprintf(":%d", servicePort)

	var opts []grpc.ServerOption
	if !config.TlsConfig.Enabled {
		insecureMode()
	} else {

		// Setup TLS.
		serverCert, err := tls.LoadX509KeyPair(config.TlsConfig.Cert, config.TlsConfig.Key)
		if err != nil {
			log.Fatalf("Failed to load server cert and key. err: %s\n", err)
		}

		// Load the CA cert.
		var cacert []byte
		cacert, err = ioutil.ReadFile(config.TlsConfig.CaCert)
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
			Certificates: []tls.Certificate{serverCert},
			RootCAs:      certPool,
			ClientCAs:    certPool,
			MinVersion:   tls.VersionTLS13,
			MaxVersion:   tls.VersionTLS13,
		}

		// Create TLS credentials from the TLS configuration
		creds := credentials.NewTLS(tlsConfig)
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	listen, err := net.Listen("tcp", portString)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		listen.Close()
	}()

	grpcServer := grpc.NewServer(opts...)
	wwapiv1.RegisterWWApiServer(grpcServer, &apiServer{})
	log.Fatalln(grpcServer.Serve(listen))
}

// private helpers

// insecureMode creates a blocking prompt for customers running wwapid in insecure mode.
// It's a deterrent. Setup TLS.
func insecureMode() {

	fmt.Println("*** Running wwapid in INSECURE mode. *** THIS IS DANGEROUS! *** Enter y to continue. ***")
	reader := bufio.NewReader(os.Stdin)
	result, err := reader.ReadString('\n')

	if err != nil {
		fmt.Printf("Fatal error: %v\n", err)
	}

	if !(result == "y\n") {
		os.Exit(1)
	}

	fmt.Printf("wwapid running IN INSECURE MODE\n")
}

// Api implementation.

// ContainerBuild builds one or more containers.
func (s *apiServer) ContainerBuild(ctx context.Context, request *wwapiv1.ContainerBuildParameter) (response *wwapiv1.ContainerListResponse, err error) {

	// Parameter checks.
	if request == nil {
		return response, status.Errorf(codes.InvalidArgument, "nil request")
	}

	if request.ContainerNames == nil {
		return response, status.Errorf(codes.InvalidArgument, "nil request.ContainerNames")
	}

	// Build the container.
	err = container.ContainerBuild(request)
	if err != nil {
		return
	}

	// Return the built containers. (A REST POST returns what is modified.)
	var containers []*wwapiv1.ContainerInfo
	containers, err = container.ContainerList()
	if err != nil {
		return
	}

	response = &wwapiv1.ContainerListResponse{}
	for i := 0; i < len(containers); i++ {
		for j := 0; j < len(request.ContainerNames); j++ {
			if containers[i].Name == request.ContainerNames[j] {
				response.Containers = append(response.Containers, containers[i])
			}
		}
	}
	return
}

// ContainerDelete deletes one or more containers from Warewulf.
func (s *apiServer) ContainerDelete(ctx context.Context, request *wwapiv1.ContainerDeleteParameter) (response *emptypb.Empty, err error) {

	// Parameter checks.
	if request == nil {
		return response, status.Errorf(codes.InvalidArgument, "nil request")
	}

	if request.ContainerNames == nil {
		return response, status.Errorf(codes.InvalidArgument, "nil request.ContainerNames")
	}

	err = container.ContainerDelete(request)
	return
}

func (s *apiServer) ContainerImport(ctx context.Context, request *wwapiv1.ContainerImportParameter) (response *wwapiv1.ContainerListResponse, err error) {

	// Import the container.
	var containerName string
	containerName, err = container.ContainerImport(request)
	if err != nil {
		return
	}

	// Return the imported container to the client.
	var containers []*wwapiv1.ContainerInfo
	containers, err = container.ContainerList()
	if err != nil {
		return
	}

	// Container name may have been shimmed in during import,
	// which is why ContainerImport returns it.
	for i := 0; i < len(containers); i++ {
		if containerName == containers[i].Name {
			response = &wwapiv1.ContainerListResponse{
				Containers: []*wwapiv1.ContainerInfo{containers[i]},
			}
			return
		}
	}
	return
}

// ContainerList returns details about containers.
func (s *apiServer) ContainerList(ctx context.Context, request *emptypb.Empty) (response *wwapiv1.ContainerListResponse, err error) {

	var containers []*wwapiv1.ContainerInfo
	containers, err = container.ContainerList()
	if err != nil {
		return
	}

	response = &wwapiv1.ContainerListResponse{
		Containers: containers,
	}
	return
}

// ContainerShow returns details about containers.
func (s *apiServer) ContainerShow(ctx context.Context, request *wwapiv1.ContainerShowParameter) (response *wwapiv1.ContainerShowResponse, err error) {

	// Parameter checks.
	if request == nil {
		return response, status.Errorf(codes.InvalidArgument, "nil request")
	}

	return container.ContainerShow(request)
}

// NodeAdd adds one or more nodes for management by Warewulf and returns the added nodes.
func (s *apiServer) NodeAdd(ctx context.Context, request *wwapiv1.NodeAddParameter) (response *wwapiv1.NodeListResponse, err error) {

	// Parameter checks.
	if request == nil {
		return response, status.Errorf(codes.InvalidArgument, "nil request")
	}

	if request.NodeNames == nil {
		return response, status.Errorf(codes.InvalidArgument, "nil request.NodeNames")
	}

	// Add the node(s).
	err = node.NodeAdd(request)
	if err != nil {
		return
	}

	// Return the added nodes as per REST.
	return s.nodeListInternal(request.NodeNames)
}

// NodeDelete deletes one or more nodes for removal of management by Warewulf.
func (s *apiServer) NodeDelete(ctx context.Context, request *wwapiv1.NodeDeleteParameter) (response *emptypb.Empty, err error) {

	response = new(emptypb.Empty)

	// Parameter checks.
	if request == nil {
		return response, status.Errorf(codes.InvalidArgument, "nil request")
	}

	if request.NodeNames == nil {
		return response, status.Errorf(codes.InvalidArgument, "nil request.NodeNames")
	}

	err = node.NodeDelete(request)
	return
}

// NodeList returns details about zero or more nodes.
func (s *apiServer) NodeList(ctx context.Context, request *wwapiv1.NodeNames) (response *wwapiv1.NodeListResponse, err error) {

	// Parameter checks. request.NodeNames can be nil.
	if request == nil {
		return response, status.Errorf(codes.InvalidArgument, "nil request")
	}

	// Perform the list.
	return s.nodeListInternal(request.NodeNames)
}

// NodeSet updates fields for zero or more nodes and returns the updated nodes.
func (s *apiServer) NodeSet(ctx context.Context, request *wwapiv1.NodeSetParameter) (response *wwapiv1.NodeListResponse, err error) {

	// Parameter checks.
	if request == nil {
		return response, status.Errorf(codes.InvalidArgument, "nil request")
	}

	if request.NodeNames == nil {
		return response, status.Errorf(codes.InvalidArgument, "nil request.NodeNames")
	}

	// Perform the NodeSet.
	err = node.NodeSet(request)
	if err != nil {
		return
	}

	// Return the updated nodes as per REST.
	return s.nodeListInternal(request.NodeNames)
}

func (s *apiServer) NodeStatus(ctx context.Context, request *wwapiv1.NodeNames) (response *wwapiv1.NodeStatusResponse, err error) {

	// Parameter checks. request.NodeNames can be nil.
	if request == nil {
		return response, status.Errorf(codes.InvalidArgument, "nil request")
	}

	return node.NodeStatus(request.NodeNames)
}

// Version returns the versions of the wwapiv1 and warewulf as well as the api
// prefix for http routes.
func (s *apiServer) Version(ctx context.Context, request *emptypb.Empty) (response *wwapiv1.VersionResponse, err error) {

	response = &wwapiv1.VersionResponse{
		ApiPrefix:       apiPrefix,
		ApiVersion:      apiVersion,
		WarewulfVersion: version.GetVersion(),
	}
	return
}

// Private helpers.

// nodeListInternal calls NodeList and returns NodeListResponse.
// This does not contain parameter checks.
func (s *apiServer) nodeListInternal(nodeNames []string) (response *wwapiv1.NodeListResponse, err error) {

	var nodes []*wwapiv1.NodeInfo
	nodes, err = node.NodeList(nodeNames)
	if err != nil {
		return
	}

	response = &wwapiv1.NodeListResponse{
		Nodes: nodes,
	}
	return
}
