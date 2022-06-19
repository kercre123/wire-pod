package robot

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
)

// DeviceCertRecord is a record of useful facts about an individual device
// certificate. The device may or may not be a robot.
type DeviceCertRecord struct {
	CommonName             string `json:"CommonName"` // CN value in cert Distinguished Name
	KeysDigest             string `json:"KeysDigest"`
	CertDigest             string `json:"CertDigest"`
	CertSignatureAlgorithm string `json:"CertSignatureAlgorithm"`
	CertSignature          string `json:"CertSignature"`
}

// Filenames and default directory for certificate/keys data
var DefaultCloudDir = "/factory/cloud"

const (
	CertFilename = "AnkiRobotDeviceCert.pem"
	KeysFilename = "AnkiRobotDeviceKeys.pem"
)

// CheckFactoryCloudFiles checks that the factory-programmed "Cloud" directory
// has been populated with the expected files, and checks that the information
// in those files looks correct and is consistent with the robot's Electronic
// Serial Number (ESN).
//
// cloudDir - path of the directory with cloud files. Eg "/factory/cloud".
//
// canonicalESN - the ESN of the robot, in "canonical" string form with
// lowercase hexadecimal digits. Eg "00e00012" for the 32-bit ESN value
// 14680082.
func CheckFactoryCloudFiles(cloudDir, canonicalESN string) error {
	expCommonName := fmt.Sprintf("vic:%s", canonicalESN)
	expOrganization := "Anki Inc."

	infoFile := filepath.Join(cloudDir, fmt.Sprintf("Info%s.json", canonicalESN))
	if bytes, err := ioutil.ReadFile(infoFile); err != nil {
		return err
	} else {
		var dcr DeviceCertRecord
		if err := json.Unmarshal(bytes, &dcr); err != nil {
			return fmt.Errorf("ERROR. Parsing JSON: %q", err)
		}
		if dcr.CommonName != expCommonName {
			return fmt.Errorf("CommonName mismatch in file %s. Exp=%s, Got=%s", infoFile, expCommonName, dcr.CommonName)
		}
	}

	keysFile := filepath.Join(cloudDir, KeysFilename)
	if keysBytesPEM, err := ioutil.ReadFile(keysFile); err != nil {
		return err
	} else {
		block, _ := pem.Decode(keysBytesPEM)
		expType := "PRIVATE KEY"
		if block.Type != expType {
			return fmt.Errorf("PEM Type mismatch in file %s. Exp=%s, Got=%s", keysFile, expType, block.Type)
		}
		if _, err := x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
			return fmt.Errorf("ERROR. Parsing private key in file %s: %s", keysFile, err)
		}
	}

	certFile := filepath.Join(cloudDir, CertFilename)
	if certBytesPEM, err := ioutil.ReadFile(certFile); err != nil {
		return err
	} else {
		block, _ := pem.Decode(certBytesPEM)
		expType := "CERTIFICATE"
		if block.Type != expType {
			return fmt.Errorf("PEM Type mismatch in file %s. Exp=%s, Got=%s", certFile, expType, block.Type)
		}
		if cert, err := x509.ParseCertificate(block.Bytes); err != nil {
			return fmt.Errorf("ERROR. Parsing certificate in file %s: %s", certFile, err)
		} else {
			if cert.Subject.CommonName != expCommonName {
				return fmt.Errorf("CommonName mismatch in file %s. Exp=%s, Got=%s", certFile, expCommonName, cert.Subject.CommonName)
			}
			expNumOrgs := 1
			gotNumOrgs := len(cert.Subject.Organization)
			if gotNumOrgs != expNumOrgs {
				return fmt.Errorf("Wrong number of Organizations in cert file %s. Exp=%d, Got=%d", certFile, expNumOrgs, gotNumOrgs)
			}
			if cert.Subject.Organization[0] != expOrganization {
				return fmt.Errorf("Organization mismatch in file %s. Exp=%s, Got=%s", certFile, expOrganization, cert.Subject.Organization[0])
			}
			invalidTooSoon, _ := time.Parse("2006/01/02", "2100/01/01")
			if !cert.NotAfter.After(invalidTooSoon) {
				return fmt.Errorf("Cert NotAfter date is too soon: %v", cert.NotAfter)
			}
		}
	}

	return nil
}

// ParseX509Certificate parses an x509 certificate from the given
// cloud factory directory
func ParseX509Certificate(cloudDir string) (*x509.Certificate, error) {
	certBytesPEM, err := PEMData(cloudDir)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(certBytesPEM)
	return x509.ParseCertificate(block.Bytes)
}

// PEMData returns the PEM data located in the given cloud
// factory directory's certificate file
func PEMData(cloudDir string) ([]byte, error) {
	certFile := filepath.Join(cloudDir, CertFilename)
	certBytesPEM, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	return certBytesPEM, nil
}

// TLSKeyPair returns the public and private key in the given
// factory directory as a tls.Certificate
func TLSKeyPair(cloudDir string) (tls.Certificate, error) {
	certFile := filepath.Join(cloudDir, CertFilename)
	keysFile := filepath.Join(cloudDir, KeysFilename)
	return tls.LoadX509KeyPair(certFile, keysFile)
}

var certCommonName string

// CertCommonName returns the CommonName field stored in the certificate
// in the given factory directory
func CertCommonName(cloudDir string) (string, error) {
	if certCommonName != "" {
		return certCommonName, nil
	}
	cert, err := ParseX509Certificate(cloudDir)
	if err != nil {
		return "", err
	}
	certCommonName = cert.Subject.CommonName
	return certCommonName, nil
}
