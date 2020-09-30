package certs

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"time"
)

const CertsPath = "certs"

var organization string
var country string
var province string
var locality string
var streetAddress string
var postalCode string

func init() {
	flag.StringVar(&organization, "organization", "goasicminer", "CA cert organization")
	flag.StringVar(&country, "country", "US", "CA cert country")
	flag.StringVar(&province, "province", "Ohio", "CA cert province")
	flag.StringVar(&locality, "locality", "Flavor Town", "CA cert locality")
	flag.StringVar(&streetAddress, "street-address", "42 Central Dr", "CA cert street address")
	flag.StringVar(&postalCode, "postal-code", "86753-09", "CA cert postal code")
}

func GetCertsPath(name string) (string, string, bool) {
	certsFolder := utils.GetSubFolder(CertsPath)
	certPath := path.Join(certsFolder, fmt.Sprintf("%s.crt", name))
	certKeyPath := path.Join(certsFolder, fmt.Sprintf("%s-key.pem", name))
	_, certErr := os.Stat(certPath)
	_, certKeyErr := os.Stat(certKeyPath)
	return certPath, certKeyPath, os.IsNotExist(certErr) || os.IsNotExist(certKeyErr)
}

func LoadCert(certName string) (tls.Certificate, error) {
	certPath, certKeyPath, _ := GetCertsPath(certName)
	return tls.LoadX509KeyPair(certPath, certKeyPath)
}

func GetCACert() (tls.Certificate, error) {
	certPath, certKeyPath, missing := GetCertsPath("ca")
	if missing {
		now := time.Now()
		cert := &x509.Certificate{
			SerialNumber: big.NewInt(int64(now.Year())),
			Subject: pkix.Name{
				Organization:  []string{organization},
				Country:       []string{country},
				Province:      []string{province},
				Locality:      []string{locality},
				StreetAddress: []string{streetAddress},
				PostalCode:    []string{postalCode},
			},
			NotBefore:             now,
			NotAfter:              now.AddDate(1, 0, 0),
			IsCA:                  true,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
			BasicConstraintsValid: true,
		}

		privKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return tls.Certificate{}, err
		}

		certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privKey.PublicKey, privKey)
		if err != nil {
			return tls.Certificate{}, err
		}

		if err := WriteCert(certBytes, certPath, privKey, certKeyPath); err != nil {
			return tls.Certificate{}, err
		}
	}
	return LoadCert("ca")
}

func WriteCert(certBytes []byte, certPath string, privKey *rsa.PrivateKey, certKeyPath string) error {
	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return err
	}

	if err := ioutil.WriteFile(certPath, certPEM.Bytes(), 0600); err != nil {
		return err
	}

	privKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(privKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	}); err != nil {
		return err
	}

	if err := ioutil.WriteFile(certKeyPath, privKeyPEM.Bytes(), 0600); err != nil {
		return err
	}
	return nil
}

func GetCert(name string) (tls.Certificate, error) {
	certPath, certKeyPath, missing := GetCertsPath(name)
	if missing {
		caCert, err := GetCACert()
		if err != nil {
			return tls.Certificate{}, err
		}
		localIps, err := utils.GetLocalIPs()
		if err != nil {
			return tls.Certificate{}, err
		}
		now := time.Now()
		cert := &x509.Certificate{
			SerialNumber: big.NewInt(now.Unix()),
			Subject: pkix.Name{
				Organization:  []string{organization},
				Country:       []string{country},
				Province:      []string{province},
				Locality:      []string{locality},
				StreetAddress: []string{streetAddress},
				PostalCode:    []string{postalCode},
			},
			IPAddresses:  localIps,
			NotBefore:    now,
			NotAfter:     now.AddDate(1, 0, 0),
			SubjectKeyId: []byte{1, 2, 3, 4, 6},
			ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			KeyUsage:     x509.KeyUsageDigitalSignature,
		}

		privKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return tls.Certificate{}, err
		}
		rawCACert, err := x509.ParseCertificate(caCert.Certificate[0])
		if err != nil {
			return tls.Certificate{}, err
		}
		certBytes, err := x509.CreateCertificate(rand.Reader, cert, rawCACert, &privKey.PublicKey,
			caCert.PrivateKey)
		if err != nil {
			return tls.Certificate{}, err
		}

		if err := WriteCert(certBytes, certPath, privKey, certKeyPath); err != nil {
			return tls.Certificate{}, err
		}
	}
	return LoadCert(name)
}
