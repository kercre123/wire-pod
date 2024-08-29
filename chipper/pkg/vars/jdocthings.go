package vars

import (
	"encoding/json"
	"os"
)

const (
	ROBOT_SETTINGS_JDOC = "vic.RobotSettings"
	APP_TOKENS_JDOC     = "vic.AppTokens"
)

type RobotSettings struct {
	ButtonWakeword int  `json:"button_wakeword"`
	Clock24Hour    bool `json:"clock_24_hour"`
	CustomEyeColor struct {
		Enabled    bool    `json:"enabled"`
		Hue        float64 `json:"hue"`
		Saturation float64 `json:"saturation"`
	} `json:"custom_eye_color"`
	DefaultLocation  string `json:"default_location"`
	DistIsMetric     bool   `json:"dist_is_metric"`
	EyeColor         int    `json:"eye_color"`
	Locale           string `json:"locale"`
	MasterVolume     int    `json:"master_volume"`
	TempIsFahrenheit bool   `json:"temp_is_fahrenheit"`
	TimeZone         string `json:"time_zone"`
}

type AJdoc struct {
	DocVersion     uint64 `protobuf:"varint,1,opt,name=doc_version,json=docVersion,proto3" json:"doc_version,omitempty"`            // first version = 1; 0 => invalid or doesn't exist
	FmtVersion     uint64 `protobuf:"varint,2,opt,name=fmt_version,json=fmtVersion,proto3" json:"fmt_version,omitempty"`            // first version = 1; 0 => invalid
	ClientMetadata string `protobuf:"bytes,3,opt,name=client_metadata,json=clientMetadata,proto3" json:"client_metadata,omitempty"` // arbitrary client-defined string, eg a data fingerprint (typ "", 32 chars max)
	JsonDoc        string `protobuf:"bytes,4,opt,name=json_doc,json=jsonDoc,proto3" json:"json_doc,omitempty"`
}

type botjdoc struct {
	// vic:00000000
	Thing string `json:"thing"`
	// vic.RobotSettings, etc
	Name string `json:"name"`
	// actual jdoc
	Jdoc AJdoc `json:"jdoc"`
}

func WriteJdocs() {
	writeBytes, _ := json.Marshal(BotJdocs)
	os.WriteFile(JdocsPath, writeBytes, 0644)
}

// removes a bot from jdocs file
func DeleteData(thing string) {
	var newdocs []botjdoc
	for _, jdocentry := range BotJdocs {
		if jdocentry.Thing != thing {
			newdocs = append(newdocs, jdocentry)
		}
	}
	BotJdocs = newdocs
	WriteJdocs()
}

func ESNToThing(esn string) (thing string) {
	return "vic:" + esn

}

func GetJdoc(thing, jdocname string) (AJdoc, bool) {
	for _, botJdoc := range BotJdocs {
		if botJdoc.Name == jdocname && botJdoc.Thing == thing {
			return botJdoc.Jdoc, true
		}
	}
	return AJdoc{}, false
}

//    DocVersion     uint64 `protobuf:"varint,1,opt,name=doc_version,json=docVersion,proto3" json:"doc_version,omitempty"`            // first version = 1; 0 => invalid or doesn't exist
// FmtVersion     uint64 `protobuf:"varint,2,opt,name=fmt_version,json=fmtVersion,proto3" json:"fmt_version,omitempty"`            // first version = 1; 0 => invalid
// ClientMetadata string `protobuf:"bytes,3,opt,name=client_metadata,json=clientMetadata,proto3" json:"client_metadata,omitempty"` // arbitrary client-defined string, eg a data fingerprint (typ "", 32 chars max)
// JsonDoc        string

func AddJdoc(thing string, name string, jdoc AJdoc) uint64 {
	var latestVersion uint64 = 0
	matched := false
	for index, jdocentry := range BotJdocs {
		if jdocentry.Thing == thing && jdocentry.Name == name {
			BotJdocs[index].Jdoc = jdoc
			latestVersion = BotJdocs[index].Jdoc.DocVersion
			matched = true
			break
		}
	}
	if !matched {
		var newbot botjdoc
		newbot.Thing = thing
		newbot.Name = name
		newbot.Jdoc = jdoc
		BotJdocs = append(BotJdocs, newbot)
	}
	WriteJdocs()
	return latestVersion
}
