package api

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

type cert struct {
}

/*
 * NewCert
 */
func NewCert() *cert {
	return &cert{}
}

func (c *cert) readCert(certInput string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certInput))
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

func (c *cert) readPrivateKey(keyInput string) (interface{}, error) {
	var parsedKey interface{}
	var err error

	privPem, _ := pem.Decode([]byte(keyInput))

	if privPem.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("RSA private key is of the wrong type: %s", privPem.Type)
	}

	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPem.Bytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPem.Bytes); err != nil {
			return nil, fmt.Errorf("Unable to parse RSA private key: %v", err)
		}
	}

	return parsedKey, nil
}

func (c *cert) createClientCert(caCert *x509.Certificate, caKey interface{}, subject string) (bytes.Buffer, bytes.Buffer, error) {
	var (
		certOut bytes.Buffer
		keyOut  bytes.Buffer
		priv    interface{}
		err     error
	)

	priv, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return certOut, keyOut, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{os.Getenv("CLIENT_CERT_ORG")},
			CommonName:   subject,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 1, 0),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, c.publicKey(priv), caKey)
	if err != nil {
		return certOut, keyOut, err
	}

	if err != nil {
		return certOut, keyOut, err
	}
	if err := pem.Encode(&certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return certOut, keyOut, err
	}

	if err := pem.Encode(&keyOut, c.pemBlockForKey(priv)); err != nil {
		return certOut, keyOut, err
	}

	return certOut, keyOut, nil
}
func (c *cert) publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}
func (c *cert) pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}
