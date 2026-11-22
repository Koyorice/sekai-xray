package proxy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"io/fs"
	"math/big"
	"os"
	"time"
)

var (
	certFile = "ca.crt"
	keyFile  = "ca.key"
)

func exists() bool {
	_, err := os.Stat(certFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false
		}

		panic(err)
	}

	_, err = os.Stat(keyFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false
		}

		panic(err)
	}

	return true
}

func loadCA() *tls.Certificate {
	certDER, err := os.ReadFile(certFile)
	if err != nil {
		panic(err)
	}

	keyDER, err := os.ReadFile(keyFile)
	if err != nil {
		panic(err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		panic(err)
	}

	key, err := x509.ParsePKCS1PrivateKey(keyDER)
	if err != nil {
		panic(err)
	}

	return &tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  key,
		Leaf:        cert,
	}
}

func ensureCA() *tls.Certificate {
	if exists() {
		return loadCA()
	}

	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, limit)
	if err != nil {
		panic(err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "PJSK Proxy MITM CA",
			Organization: []string{"PJSK Proxy MITM CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	raw, err := x509.CreateCertificate(
		rand.Reader,
		tmpl,
		tmpl,
		private.Public(),
		private,
	)
	if err != nil {
		panic(err)
	}

	cert, err := x509.ParseCertificate(raw)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(certFile, raw, 0600)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(keyFile, x509.MarshalPKCS1PrivateKey(private), 0600)
	if err != nil {
		panic(err)
	}

	return &tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  private,
		Leaf:        cert,
	}
}
