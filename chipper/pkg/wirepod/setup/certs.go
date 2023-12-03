package botsetup

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
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

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", vars.OutboundIPTester)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

// creates and exports a priv/pub key combo generated with IP address
func CreateCertCombo() error {
	// get preferred IP address of machine
	ipAddr := GetOutboundIP()

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
		IPAddresses:  []net.IP{ipAddr},
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
		ip := GetOutboundIP()
		ipString := ip.String()
		url := ipString + ":" + vars.APIConfig.Server.Port
		config.Jdocs = url
		config.Token = url
		config.Chipper = url
		config.Check = ipString + "/ok"
		config.Logfiles = "s3://anki-device-logs-prod/victor"
		config.Appkey = "oDoa0quieSeir6goowai7f"
	}
	writeBytes, _ := json.Marshal(config)
	os.WriteFile(vars.ServerConfigPath, writeBytes, 0777)
}
