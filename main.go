package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
)

var hostname string

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	servKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("generating random key: %v", err)
	}

	servCertTmpl, err := CertTemplate()
	if err != nil {
		log.Fatalf("creating cert template: %v", err)
	}
	servCertTmpl.KeyUsage = x509.KeyUsageDigitalSignature
	servCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	servCertTmpl.Subject = pkix.Name{CommonName: hostname, Organization: []string{"Self Signed WebServer"}}

	_, servCertPEM, err := CreateCert(servCertTmpl, servCertTmpl, &servKey.PublicKey, servKey)
	if err != nil {
		log.Fatalf("error creating cert: %v", err)
	}

	servKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(servKey),
	})
	serverTLSCert, err := tls.X509KeyPair(servCertPEM, servKeyPEM)
	if err != nil {
		log.Fatalf("invalid key pair: %v", err)
	}

	// SERVER
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLSCert},
	}
	server := http.Server{
		Addr:      getEnv("BIND", ":443"),
		Handler:   logRequest(http.HandlerFunc(httpRequestHandler)),
		TLSConfig: tlsConfig,
	}
	defer server.Close()
	log.Println("Hostname: " + hostname + " | Starting webserver on port " + server.Addr)
	log.Fatal(server.ListenAndServeTLS("", ""))
}

func httpRequestHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Host: " + hostname + " | Path: " + req.URL.Path + "\n"))
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
