package certs

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestGetCACert(t *testing.T) {
	if caCert, err := GetCACert(); err != nil {
		log.WithError(err).Fatal("Error getting CA certificate")
	} else {
		log.WithField("CA cert", caCert).Println("CA Certificate")
	}
}

func TestGetCerts(t *testing.T) {
	if cert, err := GetCert("rpc"); err != nil {
		log.WithError(err).Fatal("Error getting certificate")
	} else {
		log.WithField("cert", cert).Println("Certificate")
	}
}
