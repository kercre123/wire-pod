package botsetup

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"golang.org/x/crypto/ssh"
)

// this file will be copied to the bot
var SetupScriptPath = "../vector-cloud/pod-bot-install.sh"

// path to copy to
const BotSetupPath = "/data/pod-bot-install.sh"

var SetupSSHStatus string = "not running"
var SSHSettingUp bool = false

func doErr(err error, msg string) error {
	SSHSettingUp = false
	SetupSSHStatus = "not running (last error: " + err.Error() + ", last step: " + msg + ")"
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

func setCPURAMfreq(client *ssh.Client, cpufreq string, ramfreq string, gov string) {
	runCmd(client, "echo "+cpufreq+" > /sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq && echo disabled > /sys/kernel/debug/msm_otg/bus_voting && echo 0 > /sys/kernel/debug/msm-bus-dbg/shell-client/update_request && echo 1 > /sys/kernel/debug/msm-bus-dbg/shell-client/mas && echo 512 > /sys/kernel/debug/msm-bus-dbg/shell-client/slv && echo 0 > /sys/kernel/debug/msm-bus-dbg/shell-client/ab && echo active clk2 0 1 max "+ramfreq+" > /sys/kernel/debug/rpm_send_msg/message && echo "+gov+" > /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor && echo 1 > /sys/kernel/debug/msm-bus-dbg/shell-client/update_request")
}

func SetupBotViaSSH(ip string, key []byte) error {
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		SetupScriptPath = vars.AndroidPath + "/static/pod-bot-install.sh"
	}
	if vars.IsPackagedLinux {
		SetupScriptPath = "./pod-bot-install.sh"
	}
	if !SSHSettingUp {
		logger.Println("Setting up " + ip + " via SSH")
		SetupSSHStatus = "Setting up SSH connection..."
		CreateServerConfig()
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			doErr(err, "parsing priv key")
		}
		config := &ssh.ClientConfig{
			User: "root",
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback:   ssh.InsecureIgnoreHostKey(),
			HostKeyAlgorithms: []string{"ssh-rsa", "ecdsa-sha2-nistp256"},
			Timeout:           time.Second * 5,
		}
		client, err := ssh.Dial("tcp", ip+":22", config)
		if err != nil {
			return doErr(err, "ssh dial")
		}
		SetupSSHStatus = "Checking if device is a Vector..."
		output, err := runCmd(client, "uname -a")
		if err != nil {
			return doErr(err, "checking if vector")
		}
		if !strings.Contains(output, "Vector") {
			return doErr(fmt.Errorf("the remote device is not a vector"), "checking if vector")
		}
		SetupSSHStatus = "Checking if Vector is running CFW..."
		output, err = runCmd(client, "cat /build.prop")
		if err != nil {
			return doErr(err, "checking if cfw")
		}
		// outputWired, _ := runCmd(client, "cat /etc/wired/webroot/index.html")
		var doCloud bool = true
		var initCommand string = "mount -o rw,remount / && mount -o rw,remount,exec /data && systemctl stop anki-robot.target mm-anki-camera mm-qcamera-daemon"
		if strings.Contains(output, "wire_os") {
			//|| strings.Contains(outputWired, "revertDefaultWakeWord") {
			initCommand = "mount -o rw,remount,exec /data && systemctl stop anki-robot.target mm-anki-camera mm-qcamera-daemon"
			// my cfw already has a wire-pod compatible vic-cloud
			doCloud = false
		}
		SetupSSHStatus = "Running initial commands before transfers (screen will go blank, this is normal)..."
		_, err = runCmd(client, initCommand)
		if err != nil {
			if !strings.Contains(err.Error(), "Process exited with status 1") {
				return doErr(err, "initial commands")
			}
		}
		setCPURAMfreq(client, "1267200", "800000", "performance")
		SetupSSHStatus = "Waiting a few seconds for filesystem syncing"
		time.Sleep(time.Second * 3)
		SetupSSHStatus = "Transferring bot setup script and certs..."
		scpClient, err := scp.NewClientBySSH(client)
		if err != nil {
			return doErr(err, "new scp client")
		}
		script, err := os.Open(SetupScriptPath)
		if err != nil {
			return doErr(err, "opening setup script")
		}
		err = scpClient.CopyFile(context.Background(), script, "/data/pod-bot-install.sh", "0755")
		if err != nil {
			return doErr(err, "copying pod-bot-install")
		}
		scpClient.Close()
		serverConfig, err := os.Open(vars.ServerConfigPath)
		if err != nil {
			return doErr(err, "opening server config on disk")
		}
		scpClient, err = scp.NewClientBySSH(client)
		if err != nil {
			return doErr(err, "new scp client 2")
		}
		err = scpClient.CopyFile(context.Background(), serverConfig, "/data/data/server_config.json", "0755")
		if err != nil {
			return doErr(err, "copying server-config.json")
		}
		scpClient.Close()
		if doCloud {
			if runtime.GOOS != "android" && !vars.Packaged {
				cloud, err := os.Open("../vector-cloud/build/vic-cloud")
				if err != nil {
					return doErr(err, "transferring new vic-cloud")
				}
				SetupSSHStatus = "Transferring new vic-cloud..."
				scpClient, err = scp.NewClientBySSH(client)
				if err != nil {
					return doErr(err, "new scp client 3")
				}
				err = scpClient.CopyFile(context.Background(), cloud, "/anki/bin/vic-cloud", "0755")
				if err != nil {
					time.Sleep(time.Second * 1)
					scpClient, err = scp.NewClientBySSH(client)
					if err != nil {
						return doErr(err, "copying vic-cloud")
					}
					err = scpClient.CopyFile(context.Background(), cloud, "/anki/bin/vic-cloud", "0755")
					if err != nil {
						return doErr(err, "copying vic-cloud")
					}
				}
			} else {
				resp, err := http.Get("https://github.com/kercre123/wire-pod/raw/main/vector-cloud/build/vic-cloud")
				if err != nil {
					return doErr(err, "transferring new vic-cloud (download)")
				}
				SetupSSHStatus = "Transferring new vic-cloud..."
				scpClient, err = scp.NewClientBySSH(client)
				if err != nil {
					return doErr(err, "new scp client 3")
				}
				err = scpClient.CopyFile(context.Background(), resp.Body, "/anki/bin/vic-cloud", "0755")
				if err != nil {
					time.Sleep(time.Second * 1)
					scpClient, err = scp.NewClientBySSH(client)
					if err != nil {
						return doErr(err, "copying vic-cloud")
					}
					err = scpClient.CopyFile(context.Background(), resp.Body, "/anki/bin/vic-cloud", "0755")
					if err != nil {
						return doErr(err, "copying vic-cloud")
					}
				}
			}
		}
		scpClient.Close()
		certPath := vars.CertPath
		if vars.APIConfig.Server.EPConfig {
			if runtime.GOOS == "android" || runtime.GOOS == "ios" {
				certPath = vars.AndroidPath + "/static/epod/ep.crt"
			} else {
				certPath = "./epod/ep.crt"
			}
		}
		cert, err := os.Open(certPath)
		if err != nil {
			return doErr(err, "opening cert")
		}
		scpClient, err = scp.NewClientBySSH(client)
		if err != nil {
			return doErr(err, "new scp client 4")
		}
		err = scpClient.CopyFile(context.Background(), cert, "/data/data/wirepod-cert.crt", "0755")
		if err != nil {
			return doErr(err, "copying wire-pod cert")
		}
		scpClient.Close()
		SetupSSHStatus = "Generating new robot certificate (this may take a while)..."
		_, err = runCmd(client, "chmod +rwx /data/data/server_config.json /data/data/wirepod-cert.crt /data/pod-bot-install.sh && /data/pod-bot-install.sh")
		if err != nil {
			return doErr(err, "generating new robot cert")
		}
		setCPURAMfreq(client, "733333", "500000", "interactive")
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
		keyBytes, err := io.ReadAll(key)
		if len(keyBytes) < 5 {
			fmt.Fprint(w, "error: must provide ssh key ("+err.Error()+")")
			return
		}
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
	http.HandleFunc("/api-ssh/", SSHSetup)
}
