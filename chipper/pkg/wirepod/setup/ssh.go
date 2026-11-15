package botsetup

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"golang.org/x/crypto/ssh"
)

//go:embed robot_sshkey
var defaultSSHKey []byte

const ErrSSHKeyRejected = "ssh key rejected"

var SetupScriptPath = "../vector-cloud/pod-bot-install.sh"

const BotSetupPath = "/data/pod-bot-install.sh"

const vicCloudURL = "https://github.com/kercre123/wire-pod/raw/main/vector-cloud/build/vic-cloud"

var (
	sshStatusMu sync.Mutex
	sshStatus   = "not running"
	sshRunning  atomic.Bool
	sshDoneIPs  = map[string]bool{}
)

func sshDoneFor(ip string) bool {
	sshStatusMu.Lock()
	defer sshStatusMu.Unlock()
	return sshDoneIPs[ip]
}

func setSSHStatus(s string) {
	sshStatusMu.Lock()
	sshStatus = s
	sshStatusMu.Unlock()
}

func takeSSHStatus() string {
	sshStatusMu.Lock()
	defer sshStatusMu.Unlock()
	s := sshStatus
	if s == "done" || strings.Contains(s, "error") {
		sshStatus = "not running"
	}
	return s
}

func doErr(err error, msg string) error {
	setSSHStatus("not running (last error: " + err.Error() + ", last step: " + msg + ")")
	return err
}

func runCmd(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	output, err := session.Output(cmd)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func exitedWithStatus1(err error) bool {
	var exitErr *ssh.ExitError
	return errors.As(err, &exitErr) && exitErr.ExitStatus() == 1
}

func setCPURAMfreq(client *ssh.Client, cpufreq string, ramfreq string, gov string) {
	runCmd(client, "echo "+cpufreq+" > /sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq && echo disabled > /sys/kernel/debug/msm_otg/bus_voting && echo 0 > /sys/kernel/debug/msm-bus-dbg/shell-client/update_request && echo 1 > /sys/kernel/debug/msm-bus-dbg/shell-client/mas && echo 512 > /sys/kernel/debug/msm-bus-dbg/shell-client/slv && echo 0 > /sys/kernel/debug/msm-bus-dbg/shell-client/ab && echo active clk2 0 1 max "+ramfreq+" > /sys/kernel/debug/rpm_send_msg/message && echo "+gov+" > /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor && echo 1 > /sys/kernel/debug/msm-bus-dbg/shell-client/update_request")
}

func copyToBot(client *ssh.Client, data []byte, remotePath string, perm string) error {
	var err error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Second)
		}
		var scpClient scp.Client
		scpClient, err = scp.NewClientBySSH(client)
		if err != nil {
			continue
		}
		err = scpClient.CopyFile(context.Background(), bytes.NewReader(data), remotePath, perm)
		scpClient.Close()
		if err == nil {
			return nil
		}
	}
	return err
}

func checkELF(data []byte) error {
	if !bytes.HasPrefix(data, []byte("\x7fELF")) {
		return errors.New("vic-cloud is not an ELF binary")
	}
	return nil
}

func downloadVicCloud() ([]byte, error) {
	cl := &http.Client{Timeout: 2 * time.Minute}
	resp, err := cl.Get(vicCloudURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status downloading vic-cloud: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func loadVicCloud() ([]byte, error) {
	var data []byte
	var err error
	if runtime.GOOS != "android" && !vars.Packaged {
		data, err = os.ReadFile("../vector-cloud/build/vic-cloud")
	} else {
		data, err = downloadVicCloud()
	}
	if err != nil {
		return nil, err
	}
	return data, checkELF(data)
}

func setupScriptPath() string {
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		return vars.AndroidPath + "/static/pod-bot-install.sh"
	}
	if vars.IsPackagedLinux {
		return "./pod-bot-install.sh"
	}
	return SetupScriptPath
}

func certPathForBot() string {
	if !vars.Config().Server.EPConfig {
		return vars.CertPath
	}
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		return vars.AndroidPath + "/static/epod/ep.crt"
	}
	return "./epod/ep.crt"
}

func sshSigner(key []byte) (ssh.Signer, error) {
	if len(key) == 0 {
		key = defaultSSHKey
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("invalid ssh key: %w", err)
	}
	return signer, nil
}

func checkSSHReady(key []byte) error {
	if _, err := sshSigner(key); err != nil {
		return err
	}
	if sshRunning.Load() {
		return errors.New("a bot is already being setup")
	}
	return nil
}

func StartSSHSetup(ip string, key []byte) error {
	signer, err := sshSigner(key)
	if err != nil {
		return err
	}
	if !sshRunning.CompareAndSwap(false, true) {
		return errors.New("a bot is already being setup")
	}
	go func() {
		defer sshRunning.Store(false)
		defer func() {
			if r := recover(); r != nil {
				logger.Println("SSH setup panic: ", r)
				doErr(fmt.Errorf("panic: %v", r), "unknown")
			}
		}()
		if err := setupBotViaSSH(ip, signer); err != nil {
			logger.Println("SSH setup error: ", err)
		}
	}()
	return nil
}

func setupBotViaSSH(ip string, signer ssh.Signer) error {
	logger.Println("Setting up " + ip + " via SSH")
	setSSHStatus("Setting up SSH connection...")
	CreateServerConfig()
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback:   ssh.InsecureIgnoreHostKey(),
		HostKeyAlgorithms: []string{"ssh-rsa", "ecdsa-sha2-nistp256"},
		Timeout:           time.Second * 5,
	}
	client, err := ssh.Dial("tcp", net.JoinHostPort(ip, "22"), config)
	if err != nil {
		if strings.Contains(err.Error(), "unable to authenticate") {
			return doErr(errors.New(ErrSSHKeyRejected), "ssh dial")
		}
		return doErr(err, "ssh dial")
	}
	defer client.Close()

	setSSHStatus("Checking if device is a Vector...")
	output, err := runCmd(client, "uname -a")
	if err != nil {
		return doErr(err, "checking if vector")
	}
	if !strings.Contains(output, "Vector") {
		return doErr(fmt.Errorf("the remote device is not a vector"), "checking if vector")
	}

	setSSHStatus("Checking if Vector is running CFW...")
	doCloud := true
	initCommand := "mount -o rw,remount / && mount -o rw,remount,exec /data && systemctl stop anki-robot.target mm-anki-camera mm-qcamera-daemon"
	_, err = runCmd(client, "head -n1 /anki/bin/vic-gateway")
	if err != nil {
		if !exitedWithStatus1(err) {
			return doErr(err, "checking if cfw")
		}
		logger.Println("SSH setup: modern CFW detected, not copying vic-cloud")
		initCommand = "mount -o rw,remount,exec /data && systemctl stop anki-robot.target mm-anki-camera mm-qcamera-daemon"
		doCloud = false
	}

	script, err := os.ReadFile(setupScriptPath())
	if err != nil {
		return doErr(err, "opening setup script")
	}
	serverConfig, err := os.ReadFile(vars.ServerConfigPath)
	if err != nil {
		return doErr(err, "opening server config on disk")
	}
	cert, err := os.ReadFile(certPathForBot())
	if err != nil {
		return doErr(err, "opening cert")
	}
	var vicCloud []byte
	if doCloud {
		setSSHStatus("Getting new vic-cloud...")
		vicCloud, err = loadVicCloud()
		if err != nil {
			return doErr(err, "getting new vic-cloud")
		}
	}

	setSSHStatus("Running initial commands before transfers (screen will go blank, this is normal)...")
	_, err = runCmd(client, initCommand)
	if err != nil && !exitedWithStatus1(err) {
		return doErr(err, "initial commands")
	}
	setCPURAMfreq(client, "1267200", "800000", "performance")
	defer setCPURAMfreq(client, "733333", "500000", "interactive")
	setSSHStatus("Waiting a few seconds for filesystem syncing")
	time.Sleep(time.Second * 3)

	setSSHStatus("Transferring bot setup script and certs...")
	if err := copyToBot(client, script, BotSetupPath, "0755"); err != nil {
		return doErr(err, "copying pod-bot-install")
	}
	if err := copyToBot(client, serverConfig, "/data/data/server_config.json", "0755"); err != nil {
		return doErr(err, "copying server-config.json")
	}
	if doCloud {
		setSSHStatus("Transferring new vic-cloud...")
		if err := copyToBot(client, vicCloud, "/anki/bin/vic-cloud", "0755"); err != nil {
			return doErr(err, "copying vic-cloud")
		}
	}
	if err := copyToBot(client, cert, "/data/data/wirepod-cert.crt", "0755"); err != nil {
		return doErr(err, "copying wire-pod cert")
	}

	setSSHStatus("Generating new robot certificate (this may take a while)...")
	_, err = runCmd(client, "chmod +rwx /data/data/server_config.json /data/data/wirepod-cert.crt "+BotSetupPath+" && "+BotSetupPath)
	if err != nil {
		return doErr(err, "generating new robot cert")
	}
	sshStatusMu.Lock()
	sshDoneIPs[ip] = true
	sshStatus = "done"
	sshStatusMu.Unlock()
	return nil
}

func SSHSetup(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api-ssh/setup":
		ip := r.FormValue("ip")
		if ip == "" {
			fmt.Fprint(w, "error: must provide ip")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		keyBytes, err := readOptionalKey(r)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		if err := StartSSHSetup(ip, keyBytes); err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, "running")
	case "/api-ssh/get_setup_status":
		fmt.Fprint(w, takeSSHStatus())
	}
}

func readOptionalKey(r *http.Request) ([]byte, error) {
	key, _, err := r.FormFile("key")
	if errors.Is(err, http.ErrMissingFile) || errors.Is(err, http.ErrNotMultipart) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("couldn't read ssh key (%w)", err)
	}
	defer key.Close()
	keyBytes, err := io.ReadAll(key)
	if err != nil {
		return nil, fmt.Errorf("couldn't read ssh key (%w)", err)
	}
	return keyBytes, nil
}

func RegisterSSHAPI() {
	http.HandleFunc("/api-ssh/", SSHSetup)
}
