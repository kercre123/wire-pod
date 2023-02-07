package botsetup

import (
	"context"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strings"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/kercre123/chipper/pkg/logger"
	"golang.org/x/crypto/ssh"
)

// this file will be copied to the bot
const SetupScriptPath = "../vector-cloud/pod-bot-install.sh"

// path to copy to
const BotSetupPath = "/data/"

var SetupSSHStatus string = "not running"
var SettingUp bool = false

func doErr(err error) error {
	SettingUp = false
	SetupSSHStatus = "not running (last error: " + err.Error() + ")"
	return err
}

func runCmd(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	output, err := session.Output(cmd)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func SetupBotViaSSH(ip string, key []byte) error {
	if !SettingUp {
		SetupSSHStatus = "Setting up SSH connection..."
		sshCert, _ := pem.Decode(key)
		signer, err := ssh.ParsePrivateKey(sshCert.Bytes)
		if err != nil {
			doErr(err)
		}
		config := &ssh.ClientConfig{
			User: "root",
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			Timeout: 5,
		}
		client, err := ssh.Dial("tcp", ip+":22", config)
		if err != nil {
			return doErr(err)
		}
		SetupSSHStatus = "Checking if device is a Vector..."
		output, err := runCmd(client, "/bin/bash uname -a")
		if err != nil {
			return doErr(err)
		}
		if !strings.Contains(output, "Vector") {
			return doErr(fmt.Errorf("the remote device is not a vector"))
		}
		SetupSSHStatus = "Running initial commands before transfers..."
		_, err = runCmd(client, "mount -o rw,remount / && mount -o rw,remount,exec /data && systemctl stop anki-robot.target && mv /anki/data/assets/cozmo_resources/config/server_config.json /anki/data/assets/cozmo_resources/config/server_config.json.bak")
		if err != nil {
			return doErr(err)
		}
		SetupSSHStatus = "Transferring bot setup script and certs..."
		scpClient, err := scp.NewClientBySSH(client)
		if err != nil {
			return doErr(err)
		}
		script, err := os.Open(SetupScriptPath)
		if err != nil {
			return doErr(err)
		}
		err = scpClient.CopyFile(context.Background(), script, "/data/", "0755")
		if err != nil {
			return doErr(err)
		}
		serverConfig, err := os.Open("../certs/server_config.json")
		if err != nil {
			return doErr(err)
		}
		err = scpClient.CopyFile(context.Background(), serverConfig, "/anki/data/assets/cozmo_resources/config/", "0755")
		if err != nil {
			return doErr(err)
		}
		certPath := "../certs/cert.crt"
		if _, err := os.Stat("./useepod"); err == nil {
			certPath = "./epod/ep.crt"
		}
		cert, err := os.Open(certPath)
		if err != nil {
			return doErr(err)
		}
		err = scpClient.CopyFile(context.Background(), cert, "/anki/etc/wirepod-cert.crt", "0755")
		if err != nil {
			return doErr(err)
		}
		SetupSSHStatus = "Running final commands (this may take a while)..."
		output, err = runCmd(client, "chmod +rwx /anki/data/assets/cozmo_resources/config/server_config.json /anki/bin/vic-cloud /data/data/wirepod-cert.crt /anki/etc/wirepod-cert.crt /data/pod-bot-install.sh && /data/pod-bot-install.sh")
		if err != nil {
			return doErr(err)
		}
		logger.Println(string(output))
		client.Close()
		SetupSSHStatus = "done"
	} else {
		return fmt.Errorf("a bot is already being setup")
	}
	return nil
}

func SSHSetup(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/api-ssh/setup":
		ip := r.FormValue("ip")
		if ip == "" {
			fmt.Fprint(w, "error: must provide ip")
			return
		}
		key, _, err := r.FormFile("key")
		if err != nil {
			fmt.Fprint(w, "error: must provide ssh key ("+err.Error()+")")
			return
		}
		var keyBytes []byte
		key.Read(keyBytes)
		go SetupBotViaSSH(ip, keyBytes)
		fmt.Fprint(w, "running")
		return
	case r.URL.Path == "/api-ssh/get_setup_status":
		fmt.Fprint(w, SetupSSHStatus)
		if SetupSSHStatus == "done" || strings.Contains(SetupSSHStatus, "error") {
			SetupSSHStatus = "not running"
		}
		return
	}
}

func RegisterSSHAPI() {
	http.HandleFunc("/api-ssh", SSHSetup)
}
