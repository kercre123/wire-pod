package botsetup

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

type ClientServerConfig struct {
	Jdocs    string `json:"jdocs"`
	Token    string `json:"tms"`
	Chipper  string `json:"chipper"`
	Check    string `json:"check"`
	Logfiles string `json:"logfiles"`
	Appkey   string `json:"appkey"`
}

// creates and exports a priv/pub key combo generated with IP address
func CreateCertCombo() error {
	// Build the SAN set. The robot pins this leaf into its root pool and
	// derives ServerName from the dial authority, so whatever address ends up
	// in server_config.json must appear here or the handshake fails.
	host := vars.GetRobotFacingAddress()
	var ips []net.IP
	var dnsNames []string

	// Retain the outbound IP so an unconfigured install generates exactly the
	// SAN set it always has. Skip it when the UDP probe failed and returned
	// the unspecified address, which would only pin a useless SAN.
	if outbound := vars.GetOutboundIP(); !outbound.IsUnspecified() {
		ips = append(ips, outbound)
	}
	if parsed := net.ParseIP(host); parsed != nil {
		if len(ips) == 0 || !parsed.Equal(ips[0]) {
			ips = append(ips, parsed)
		}
	} else {
		dnsNames = append(dnsNames, host)
	}
	if len(ips) == 0 && len(dnsNames) == 0 {
		return errors.New("cannot generate a certificate with no SANs: no outbound IP and no robot-facing address configured")
	}

	// ca certificate
	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(2019),
		Subject:               pkix.Name{},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(30, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 1028)
	if err != nil {
		return err
	}

	// create actual certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject:      pkix.Name{},
		IPAddresses:  ips,
		DNSNames:     dnsNames,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certPrivKey, err := rsa.GenerateKey(rand.Reader, 1028)
	if err != nil {
		return err
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}
	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	// export certificates
	os.MkdirAll(vars.Certs, 0777)
	logger.Println("Outputting certificate to " + vars.CertPath)
	err = os.WriteFile(vars.CertPath, certPEM.Bytes(), 0777)
	if err != nil {
		return err
	}
	logger.Println("Outputting private key to " + vars.KeyPath)
	err = os.WriteFile(vars.KeyPath, certPrivKeyPEM.Bytes(), 0777)
	if err != nil {
		return err
	}
	vars.ChipperCert = certPEM.Bytes()
	vars.ChipperKey = certPrivKeyPEM.Bytes()
	vars.ChipperKeysLoaded = true

	return nil
}

// WarnIfCertMissesAddress is a read-only coherence check. The robot pins
// wire-pod's leaf certificate and verifies the hostname it dials against that
// leaf's SANs, so a server_config.json that names an address the certificate
// does not cover produces a robot which cannot complete a TLS handshake.
//
// This only warns. Regenerating the certificate to fix the mismatch would
// de-pair every already-provisioned robot, so that has to stay a deliberate
// admin action via /api/generate_certs.
func WarnIfCertMissesAddress(host string) {
	certBytes, err := os.ReadFile(vars.CertPath)
	if err != nil {
		// No certificate yet is normal during first-run setup.
		return
	}
	block, _ := pem.Decode(certBytes)
	if block == nil {
		logger.Println("WARNING: " + vars.CertPath + " is not valid PEM, cannot check it covers " + host)
		return
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		logger.Println("WARNING: could not parse " + vars.CertPath + " to check it covers " + host + ": " + err.Error())
		return
	}
	if cert.VerifyHostname(host) == nil {
		return
	}

	var covered []string
	for _, ip := range cert.IPAddresses {
		covered = append(covered, ip.String())
	}
	covered = append(covered, cert.DNSNames...)
	logger.Println("Note: the existing certificate does not cover " + host +
		" (it covers " + strings.Join(covered, ", ") + "). Robots already paired " +
		"are unaffected and keep working. A robot set up against " + host + " from " +
		"now on would fail the TLS handshake, which regenerating the certificate " +
		"would fix -- but that de-pairs every already-paired robot, so only do it " +
		"if you are prepared to set them all up again.")
}

// outputs a server config to ../certs/server_config.json
func CreateServerConfig() {
	os.MkdirAll(vars.Certs, 0777)
	var config ClientServerConfig
	//{"jdocs": "escapepod.local:443", "tms": "escapepod.local:443", "chipper": "escapepod.local:443", "check": "escapepod.local/ok:80", "logfiles": "s3://anki-device-logs-prod/victor", "appkey": "oDoa0quieSeir6goowai7f"}
	if vars.APIConfig.Server.EPConfig {
		config.Jdocs = "escapepod.local:443"
		config.Token = "escapepod.local:443"
		config.Chipper = "escapepod.local:443"
		config.Check = "escapepod.local/ok"
		config.Logfiles = "s3://anki-device-logs-prod/victor"
		config.Appkey = "oDoa0quieSeir6goowai7f"
	} else {
		// Robot-facing address: the configured value when set, else the
		// outbound IP, which is exactly the previous behavior.
		host := vars.GetRobotFacingAddress()
		url := host + ":" + vars.APIConfig.Server.Port
		config.Jdocs = url
		config.Token = url
		config.Chipper = url
		// The check endpoint is plaintext HTTP on port 80 and carries no port.
		config.Check = host + "/ok"
		WarnIfCertMissesAddress(host)
		config.Logfiles = "s3://anki-device-logs-prod/victor"
		config.Appkey = "oDoa0quieSeir6goowai7f"
	}
	writeBytes, _ := json.Marshal(config)
	os.WriteFile(vars.ServerConfigPath, writeBytes, 0777)
}
