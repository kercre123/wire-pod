package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// handleTestLLM answers the web UI's "Test connection" button: it asks the AI server for its model
// list with the key that was typed in (not yet saved) and says in a sentence whether that worked.
// It only reports a status and a model count; the response body is never passed on.

var llmTestClient = &http.Client{Timeout: 8 * time.Second}

type testLLMResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func handleTestLLM(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Provider string `json:"provider"`
		Key      string `json:"key"`
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testLLM(req.Provider, strings.TrimSpace(req.Key), strings.TrimSpace(req.Endpoint)))
}

func testLLM(provider, key, endpoint string) testLLMResult {
	var base string
	switch provider {
	case "openai":
		base = "https://api.openai.com/v1"
	case "together":
		base = "https://api.together.xyz/v1"
	case "custom":
		base = strings.TrimRight(endpoint, "/")
		u, err := url.Parse(base)
		if base == "" || err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return testLLMResult{false, "Enter the server address first, for example http://localhost:11434/v1."}
		}
	case "houndify":
		return testLLMResult{false, "Houndify can’t be tested from here. Save it and ask a question."}
	default:
		return testLLMResult{false, "Choose an AI provider first."}
	}
	req, err := http.NewRequest("GET", base+"/models", nil)
	if err != nil {
		return testLLMResult{false, "That server address isn’t valid."}
	}
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := llmTestClient.Do(req)
	if err != nil {
		return testLLMResult{false, "Couldn’t reach the server. Check the address and that it is running."}
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return testLLMResult{false, fmt.Sprintf("The server rejected the key (HTTP %d).", resp.StatusCode)}
	case resp.StatusCode != http.StatusOK:
		return testLLMResult{false, fmt.Sprintf("The server answered with HTTP %d. Check the address.", resp.StatusCode)}
	}
	var models struct {
		Data []json.RawMessage `json:"data"`
	}
	if json.NewDecoder(resp.Body).Decode(&models) == nil && len(models.Data) > 0 {
		return testLLMResult{true, fmt.Sprintf("Connected. %d model%s available.", len(models.Data), plural(len(models.Data)))}
	}
	return testLLMResult{true, "Connected."}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
