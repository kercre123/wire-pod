package vars

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// a way to create a JSON configuration for wire-pod, rather than the use of env vars

var ApiConfigPath = "./apiConfig.json"

var APIConfig apiConfig

type apiConfig struct {
	Weather struct {
		Enable   bool   `json:"enable"`
		Provider string `json:"provider"`
		Key      string `json:"key"`
		Unit     string `json:"unit"`
	} `json:"weather"`
	Knowledge struct {
		Enable                 bool    `json:"enable"`
		Provider               string  `json:"provider"`
		Key                    string  `json:"key"`
		ID                     string  `json:"id"`
		Model                  string  `json:"model"`
		IntentGraph            bool    `json:"intentgraph"`
		RobotName              string  `json:"robotName"`
		OpenAIPrompt           string  `json:"openai_prompt"`
		OpenAIVoice            string  `json:"openai_voice"`
		OpenAIVoiceWithEnglish bool    `json:"openai_voice_with_english"`
		SaveChat               bool    `json:"save_chat"`
		CommandsEnable         bool    `json:"commands_enable"`
		Endpoint               string  `json:"endpoint"`
		TopP                   float32 `json:"top_p"`
		Temperature            float32 `json:"temp"`
	} `json:"knowledge"`
	STT struct {
		Service  string `json:"provider"`
		Language string `json:"language"`
	} `json:"STT"`
	Server struct {
		// false for ip, true for escape pod
		EPConfig bool   `json:"epconfig"`
		Port     string `json:"port"`
		// Gates the proxy-related UI and the BLE address capture. Set once
		// during first-run setup. Does not gate the accessors: a configured
		// address is always honored, so it can never silently stop applying.
		ProxyMode bool `json:"proxy_mode,omitempty"`
		// Address robots should use to reach wire-pod. Hostname or IP.
		// Empty means derive it from GetOutboundIP(), the existing behavior.
		// Applies to the IP/unlocked path only; EP mode is hardcoded.
		RobotFacingAddress string `json:"robot_facing_address,omitempty"`
		// ESN -> the LAN address wire-pod should dial for that robot. Empty or
		// absent means use the observed peer address, the existing behavior.
		// Kept here rather than on RobotInfo so botSdkInfo.json, which is
		// machine-observed state, keeps its exact on-disk shape.
		RobotLANAddresses map[string]string `json:"robot_lan_addresses,omitempty"`
	} `json:"server"`
	HasReadFromEnv   bool `json:"hasreadfromenv"`
	PastInitialSetup bool `json:"pastinitialsetup"`
}

// NormalizeConfig trims the admin-supplied networking fields and drops entries
// that carry no information. An empty RobotLANAddresses map is reduced to nil
// so that omitempty keeps a default install's apiConfig.json byte-identical.
//
// It also clears a robot-facing address that would not survive use. The setters
// validate, but a hand-edited file does not go through them, and
// CreateServerConfig concatenates ":" + port onto whatever it finds, so a
// stored "host:443" becomes "host:443:443".
//
// Must run before anything reads these fields and before every write.
func NormalizeConfig() {
	APIConfig.Server.RobotFacingAddress = strings.TrimSpace(APIConfig.Server.RobotFacingAddress)
	if err := ValidateRobotFacingAddress(APIConfig.Server.RobotFacingAddress); err != nil {
		logger.Println("Ignoring the configured robot-facing address " +
			strconv.Quote(APIConfig.Server.RobotFacingAddress) + ": " + err.Error() +
			". Falling back to this machine's outbound address.")
		APIConfig.Server.RobotFacingAddress = ""
	}

	robotLANMu.Lock()
	defer robotLANMu.Unlock()

	if len(APIConfig.Server.RobotLANAddresses) == 0 {
		APIConfig.Server.RobotLANAddresses = nil
		return
	}
	// Sorted, because lookups are case-insensitive: two entries differing only
	// in case would otherwise resolve by map iteration order, so the UI could
	// show one address while wire-pod dialled the other, and which one won
	// could change on restart. Sorting makes the survivor the same every load.
	keys := make([]string, 0, len(APIConfig.Server.RobotLANAddresses))
	for esn := range APIConfig.Server.RobotLANAddresses {
		keys = append(keys, esn)
	}
	sort.Strings(keys)

	normalized := make(map[string]string, len(keys))
	for _, key := range keys {
		esn := strings.ToLower(strings.TrimSpace(key))
		addr := strings.TrimSpace(APIConfig.Server.RobotLANAddresses[key])
		// A blank ESN or address carries no intent; dropping it here is what
		// makes "set to empty" clear an entry.
		if esn == "" || addr == "" {
			continue
		}
		if existing, dup := normalized[esn]; dup && existing != addr {
			logger.Println("Duplicate robot LAN address entries for ESN " + esn +
				"; discarding " + existing + " and keeping " + addr)
		}
		normalized[esn] = addr
	}
	if len(normalized) == 0 {
		normalized = nil
	}
	APIConfig.Server.RobotLANAddresses = normalized
}

// ValidateRobotFacingAddress rejects anything that is not a bare host or IPv4
// address. This is not cosmetic: CreateServerConfig concatenates
// addr + ":" + port, so an input of "192.168.1.10:443" would silently yield
// "192.168.1.10:443:443" and produce a robot that cannot connect.
func ValidateRobotFacingAddress(addr string) error {
	if addr == "" {
		// Empty is the default and means "derive from GetOutboundIP()".
		return nil
	}
	if strings.ContainsAny(addr, " \t\r\n") {
		return errors.New("address must not contain whitespace")
	}
	if strings.Contains(addr, "://") {
		return fmt.Errorf("address must be a bare host or IP, not a URL: %q", addr)
	}
	if strings.Contains(addr, "/") {
		return fmt.Errorf("address must not contain a path: %q", addr)
	}
	if strings.Contains(addr, "@") {
		return fmt.Errorf("address must not contain userinfo: %q", addr)
	}
	if ip := net.ParseIP(addr); ip != nil {
		if ip.To4() == nil {
			return fmt.Errorf("address must be IPv4 or a hostname; server_config.json "+
				"has no bracket syntax for IPv6: %q", addr)
		}
		return nil
	}
	if strings.Contains(addr, ":") {
		return fmt.Errorf("address must not include a port, that comes from "+
			"the server port setting: %q", addr)
	}
	return validateHostname(addr)
}

// validateHostname accepts a DNS name of the kind that can appear as a
// certificate DNS SAN, e.g. "escapepod.local" or "wirepod.lan".
func validateHostname(host string) error {
	if len(host) > 253 {
		return fmt.Errorf("hostname is longer than 253 characters: %q", host)
	}
	labels := strings.Split(strings.TrimSuffix(host, "."), ".")
	for _, label := range labels {
		if label == "" {
			return fmt.Errorf("hostname has an empty label: %q", host)
		}
		if len(label) > 63 {
			return fmt.Errorf("hostname label %q is longer than 63 characters", label)
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return fmt.Errorf("hostname label %q must not start or end with a hyphen", label)
		}
		for _, r := range label {
			isLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
			isDigit := r >= '0' && r <= '9'
			if !isLetter && !isDigit && r != '-' {
				return fmt.Errorf("hostname %q contains an invalid character %q", host, r)
			}
		}
	}
	return nil
}

func WriteConfigToDisk() {
	logger.Println("Configuration changed, writing to disk")
	NormalizeConfig()
	writeBytes, _ := json.Marshal(APIConfig)
	os.WriteFile(ApiConfigPath, writeBytes, 0644)
}

func CreateConfigFromEnv() {
	// if no config exists, create it
	if os.Getenv("WEATHERAPI_ENABLED") == "true" {
		APIConfig.Weather.Enable = true
		APIConfig.Weather.Provider = os.Getenv("WEATHERAPI_PROVIDER")
		APIConfig.Weather.Key = os.Getenv("WEATHERAPI_KEY")
		APIConfig.Weather.Unit = os.Getenv("WEATHERAPI_UNIT")
	} else {
		APIConfig.Weather.Enable = false
	}
	if os.Getenv("KNOWLEDGE_ENABLED") == "true" {
		APIConfig.Knowledge.Enable = true
		APIConfig.Knowledge.Provider = os.Getenv("KNOWLEDGE_PROVIDER")
		if os.Getenv("KNOWLEDGE_PROVIDER") == "houndify" {
			APIConfig.Knowledge.ID = os.Getenv("KNOWLEDGE_ID")
		}
		APIConfig.Knowledge.Key = os.Getenv("KNOWLEDGE_KEY")
	} else {
		APIConfig.Knowledge.Enable = false
	}
	WriteSTT()
	APIConfig.HasReadFromEnv = true
	writeBytes, _ := json.Marshal(APIConfig)
	os.WriteFile(ApiConfigPath, writeBytes, 0644)
}

func WriteSTT() {
	// was not part of the original code, so this is its own function
	// launched if stt not found in config
	APIConfig.STT.Service = os.Getenv("STT_SERVICE")
	if os.Getenv("STT_SERVICE") == "vosk" || os.Getenv("STT_SERVICE") == "whisper.cpp" {
		APIConfig.STT.Language = os.Getenv("STT_LANGUAGE")
	}
}

func ReadConfig() {
	if _, err := os.Stat(ApiConfigPath); err != nil {
		CreateConfigFromEnv()
		logger.Println("API config JSON created")
	} else {
		// read config
		configBytes, err := os.ReadFile(ApiConfigPath)
		if err != nil {
			APIConfig.Knowledge.Enable = false
			APIConfig.Weather.Enable = false
			logger.Println("Failed to read API config file")
			logger.Println(err)
			// Normalize on the error paths too: APIConfig may already hold a
			// partially populated value, and no un-normalized field may reach
			// an accessor.
			NormalizeConfig()
			return
		}
		err = json.Unmarshal(configBytes, &APIConfig)
		if err != nil {
			APIConfig.Knowledge.Enable = false
			APIConfig.Weather.Enable = false
			logger.Println("Failed to unmarshal API config JSON")
			logger.Println(err)
			NormalizeConfig()
			return
		}
		NormalizeConfig()
		// stt service is the only thing controlled by shell
		if APIConfig.STT.Service != os.Getenv("STT_SERVICE") {
			WriteSTT()
		}
		if !APIConfig.HasReadFromEnv {
			if APIConfig.Server.Port != os.Getenv("DDL_RPC_PORT") {
				APIConfig.HasReadFromEnv = true
				APIConfig.PastInitialSetup = true
			}
		}

		if APIConfig.Knowledge.Model == "meta-llama/Llama-2-70b-chat-hf" {
			logger.Println("Setting Together model to Llama3")
			APIConfig.Knowledge.Model = "meta-llama/Llama-3-70b-chat-hf"
		}

		writeBytes, _ := json.Marshal(APIConfig)
		os.WriteFile(ApiConfigPath, writeBytes, 0644)
		logger.Println("API config successfully read")
	}
}
