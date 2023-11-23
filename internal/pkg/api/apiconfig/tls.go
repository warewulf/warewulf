package apiconfig

// TlsConfig contains TLS configuration parameters for a client or server.
type TlsConfig struct {
	// Enabled is true when secure.
	Enabled bool `yaml:"enabled"`
	// Cert is the path to the client or server certificate file.
	Cert string `yaml:"cert,omitempty"`
	// Key is the path to the client or server key file.
	Key string `yaml:"key,omitempty"`
	// CaCert is the path the CA certificate file.
	CaCert string `yaml:"cacert,omitempty"`
	// ConcatCert is for wwapird. http.ListenAndServeTLS wants the following
	// cert file, so in our case this file contains `cat ${Cert} ${CaCert}`
	//
	// If the certificate is signed by a certificate authority, the certFile
	// should be the concatenation of the server's certificate, any
	// intermediates, and the CA's certificate
	ConcatCert string `yaml:"concatcert,omitempty"`
}
