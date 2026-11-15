package sdkapp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

var transCfg = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

func updateSettings(robot *Robot, settings map[string]interface{}) error {
	body, err := json.Marshal(map[string]interface{}{
		"update_settings": true,
		"settings":        settings,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "https://"+robot.Target+"/v1/update_settings", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+robot.GUID)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func setCustomEyeColor(robot *Robot, hue string, sat string) error {
	h, err := strconv.ParseFloat(hue, 64)
	if err != nil {
		return fmt.Errorf("invalid hue: %w", err)
	}
	s, err := strconv.ParseFloat(sat, 64)
	if err != nil {
		return fmt.Errorf("invalid saturation: %w", err)
	}
	return updateSettings(robot, map[string]interface{}{
		"custom_eye_color": map[string]interface{}{
			"enabled":    true,
			"hue":        h,
			"saturation": s,
		},
	})
}

func setPresetEyeColor(robot *Robot, value string) error {
	color, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid eye color: %w", err)
	}
	return updateSettings(robot, map[string]interface{}{
		"custom_eye_color": map[string]interface{}{"enabled": false},
		"eye_color":        color,
	})
}

func setSettingSDKstring(robot *Robot, setting string, value string) error {
	return updateSettings(robot, map[string]interface{}{setting: value})
}

func setSettingSDKintbool(robot *Robot, setting string, value string) error {
	if i, err := strconv.Atoi(value); err == nil {
		return updateSettings(robot, map[string]interface{}{setting: i})
	}
	if b, err := strconv.ParseBool(value); err == nil {
		return updateSettings(robot, map[string]interface{}{setting: b})
	}
	return fmt.Errorf("invalid value for %s: %q", setting, value)
}
