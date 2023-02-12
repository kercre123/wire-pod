package ble

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	updateAccountSettingsPath = "/v1/update_account_settings"
	updateSettingsPath        = "/v1/update_settings"
	sendOnboardingInputPath   = "/v1/send_onboarding_input"
)

// VectorSettings is a list of settings for the robot
type VectorSettings struct {
	Timezone           string `json:"timezone,omitempty"`
	DefaultLocation    string `json:"default_location,omitempty"`
	Locale             string `json:"locale,omitempty"`
	AllowDataAnalytics bool   `json:"allow_data_analytics,omitempty"`
	MetricDistance     bool   `json:"metric_distance,omitempty"`
	MetricTemperature  bool   `json:"metric_temperature,omitempty"`
	ButtonWakeword     bool   `json:"button_wakeword,omitempty"`
	Clock24Hour        bool   `json:"clock_24_hour,omitempty"`
	AlexaOptIn         bool   `json:"alexaOptIn,omitempty"`
}

type updateAccountSettings struct {
	AccountSettings accountsettings `json:"account_settings"`
}
type accountsettings struct {
	DataCollection bool   `json:"data_collection"`
	AppLocale      string `json:"app_locale"`
}

func (u *updateAccountSettings) String() string {
	str, _ := json.Marshal(u)
	return string(str)
}

type updateSettings struct {
	Settings settings `json:"settings"`
}

type settings struct {
	TimeZone         string `json:"time_zone"`
	DefaultLocation  string `json:"default_location,omitempty"`
	Locale           string `json:"locale"`
	DistIsMetric     bool   `json:"dist_is_metric"`
	TempIsFahrenheit bool   `json:"temp_is_fahrenheit"`
}

func (u *updateSettings) String() string {
	str, _ := json.Marshal(u)
	return string(str)
}

type markComplete struct {
	OnboardingMarkCompleteAndExit struct {
	} `json:"onboarding_mark_complete_and_exit"`
}

func (u *markComplete) String() string {
	str, _ := json.Marshal(u)
	return string(str)
}

// ConfigureSettings sends the settings to the robot
func (v *VectorBLE) ConfigureSettings(settings *VectorSettings) error {
	if !v.state.authorized {
		return errors.New(errNotAuthorized)
	}

	if v.ble.Version() != rtsV5 {
		return errors.New("unsupported version")
	}

	if err := v.updateAccountSettings(settings); err != nil {
		return err
	}

	if err := v.updateSettings(settings); err != nil {
		return err
	}

	if err := v.onboardMarkComplete(settings); err != nil {
		return err
	}

	return nil
}

func (v *VectorBLE) updateAccountSettings(s *VectorSettings) error {
	as := updateAccountSettings{
		AccountSettings: accountsettings{
			DataCollection: s.AllowDataAnalytics,
			AppLocale:      s.Locale,
		},
	}

	r, err := v.SDKProxy(
		&SDKProxyRequest{
			URLPath: updateAccountSettingsPath,
			Body:    as.String(),
		},
	)
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("update account settings failure: %s", r.ResponseBody)
	}
	return nil
}

func (v *VectorBLE) updateSettings(s *VectorSettings) error {
	as := updateSettings{
		Settings: settings{
			DefaultLocation:  s.DefaultLocation,
			TimeZone:         s.Timezone,
			Locale:           s.Locale,
			DistIsMetric:     s.MetricDistance,
			TempIsFahrenheit: (!s.MetricTemperature),
		},
	}

	r, err := v.SDKProxy(
		&SDKProxyRequest{
			URLPath: updateSettingsPath,
			Body:    as.String(),
		},
	)
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("update settings failure: %s", r.ResponseBody)
	}
	return nil
}

func (v *VectorBLE) onboardMarkComplete(s *VectorSettings) error {
	as := markComplete{}

	r, err := v.SDKProxy(
		&SDKProxyRequest{
			URLPath: sendOnboardingInputPath,
			Body:    as.String(),
		},
	)
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("onboarding completion failure: %s", r.ResponseBody)
	}

	return nil
}
