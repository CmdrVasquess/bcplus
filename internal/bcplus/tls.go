package bcplus

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/uuid"
)

const (
	certFile      = "bcplus-cert.pem"
	keyFile       = "bcplus-key.pem"
	TLSCommonName = "de.fractalqb.bcplus.app"
)

func newTLSCert(dir, commonName string) (err error) {
	log.Infoa("create new TLS certificate in `dir`", dir)
	pKey, err := ecdsa.GenerateKey(elliptic.P384(), crand.Reader)
	if err != nil {
		return err
	}
	validStart := time.Now()
	validTil := validStart.Add(10 * 365 * 24 * time.Hour) // ~ 10 years
	serNo, err := uuid.NewV4()
	if err != nil {
		return err
	}
	cerTmpl := x509.Certificate{
		SerialNumber:          new(big.Int).SetBytes(serNo.Bytes()),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             validStart,
		NotAfter:              validTil,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	cerDer, err := x509.CreateCertificate(
		crand.Reader,
		&cerTmpl, &cerTmpl,
		pKey.Public(), pKey)
	if err != nil {
		return fmt.Errorf("create cert: %s", err)
	}

	fnm := filepath.Join(dir, certFile)
	wr, err := os.Create(fnm)
	if err != nil {
		return fmt.Errorf("create cert-file '%s': %s", fnm, err)
	}
	defer wr.Close()
	err = pem.Encode(wr, &pem.Block{Type: "CERTIFICATE", Bytes: cerDer})
	if err != nil {
		return fmt.Errorf("write cert to '%s': %s", fnm, err)
	}
	err = wr.Close()
	if err != nil {
		return fmt.Errorf("close cert-file '%s': %s", fnm, err)
	}

	ecpem, err := x509.MarshalECPrivateKey(pKey)
	if err != nil {
		return fmt.Errorf("marshal private key: %s", err)
	}
	block := &pem.Block{Type: "EC PRIVATE KEY", Bytes: ecpem}
	fnm = filepath.Join(dir, keyFile)
	wr, err = os.OpenFile(fnm, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create key-file '%s': %s", fnm, err)
	}
	defer wr.Close()
	err = pem.Encode(wr, block)
	if err != nil {
		return fmt.Errorf("write key-file '%s': %s", fnm, err)
	}
	err = wr.Close()
	if err != nil {
		return fmt.Errorf("close key-file '%s': %s", fnm, err)
	}

	return nil
}

func mustTLSCert(dir string) error {
	certNm := filepath.Join(dir, certFile)
	keyNm := filepath.Join(dir, keyFile)
	if _, err := os.Stat(certNm); os.IsNotExist(err) {
		err = newTLSCert(dir, TLSCommonName)
		if err != nil {
			return err
		}
	} else if _, err = os.Stat(keyNm); os.IsNotExist(err) {
		err = newTLSCert(dir, TLSCommonName)
		if err != nil {
			return err
		}
	}
	//	res, err := tls.LoadX509KeyPair(certNm, keyNm)
	return nil
}
