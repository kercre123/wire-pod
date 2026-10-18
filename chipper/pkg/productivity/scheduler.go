package productivity

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

const (
	ProductivityImgPath = "./productivity-images"
	TodoistAPIEndpoint  = "https://api.todoist.com/rest/v2/tasks"
)

type ManualReminder struct {
	ID                  string                 `json:"id"`
	Enabled             bool                   `json:"enabled"`
	Image               string                 `json:"image"`
	Phrases             []string               `json:"phrases"`
	RequireConfirmation bool                   `json:"require_confirmation"`
	SnoozeMinutes       int                    `json:"snooze_minutes"`
	Schedule            ManualReminderSchedule `json:"schedule"`
}

type ManualReminderSchedule struct {
	Type       string   `json:"type"`
	Time       string   `json:"time"`
	Minute     int      `json:"minute"`
	MinMinutes int      `json:"min_minutes"`
	MaxMinutes int      `json:"max_minutes"`
	Days       []string `json:"days"`
	Hours      []int    `json:"hours"`
}

type TodoistDue struct {
	Datetime string `json:"datetime"`
	Timezone string `json:"timezone"`
}

type TodoistTask struct {
	ID      string      `json:"id"`
	Content string      `json:"content"`
	Due     *TodoistDue `json:"due"`
}

var (
	nextRandomRun      = make(map[string]time.Time)
	lastProcessedTasks = make(map[string]bool)
	lastManualRun      = make(map[string]string)
	schedulerQuit      = make(chan bool)
	externalApiClient  = &http.Client{Timeout: 15 * time.Second}
)

func StartScheduler() {
	logger.Println("Productivity: Starting Scheduler")
	rand.Seed(time.Now().UnixNano())
	go schedulerLoop()
	go executorLoop()
}

func StopScheduler() {
	schedulerQuit <- true
}

func schedulerLoop() {
	ticker := time.NewTicker(30 * time.Second)
	externalApiTicker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	defer externalApiTicker.Stop()

	for {
		select {
		case <-schedulerQuit:
			return
		case <-externalApiTicker.C:
			provider := vars.APIConfig.Productivity.Provider
			if provider == "todoist" {
				checkTodoistTasks()
			}
		case <-ticker.C:
			configStr := vars.APIConfig.Productivity.ManualConfig
			if configStr == "" || configStr == "[]" {
				continue
			}

			targetBot := vars.APIConfig.Productivity.TargetRobot
			if targetBot == "" || targetBot == "None" {
				if len(vars.BotInfo.Robots) > 0 {
					targetBot = vars.BotInfo.Robots[0].Esn
				} else {
					continue
				}
			}
			checkManualReminders(targetBot, configStr)
		}
	}
}

func checkTodoistTasks() {
	token := vars.APIConfig.Productivity.Key
	if token == "" {
		return
	}

	req, err := http.NewRequest("GET", TodoistAPIEndpoint, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := externalApiClient.Do(req)
	if err != nil {
		logger.Println("Productivity: Todoist connection failed")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var allTasks []TodoistTask
	if err := json.Unmarshal(body, &allTasks); err != nil {
		return
	}

	now := time.Now()
	y, m, d := now.Date()
	startOfNextDay := time.Date(y, m, d+1, 0, 0, 0, 0, now.Location())

	targetBot := vars.APIConfig.Productivity.TargetRobot
	if targetBot == "" || targetBot == "None" {
		if len(vars.BotInfo.Robots) > 0 {
			targetBot = vars.BotInfo.Robots[0].Esn
		} else {
			return
		}
	}

	for _, task := range allTasks {
		if task.Due == nil || task.Due.Datetime == "" {
			continue
		}

		taskTime, err := time.Parse(time.RFC3339, task.Due.Datetime)
		if err != nil {
			continue
		}

		if taskTime.After(now) && taskTime.Before(startOfNextDay) {
			if !lastProcessedTasks[task.ID] {
				logger.Println("Productivity: Todoist task matched: " + task.Content)
				taskQueue <- Task{
					ID:       task.ID,
					RobotESN: targetBot,
					Phrases:  []string{task.Content},
					Source:   "todoist",
				}
				lastProcessedTasks[task.ID] = true
			}
		}
	}
}

func checkManualReminders(esn string, configStr string) {
	var reminders []ManualReminder
	if err := json.Unmarshal([]byte(configStr), &reminders); err != nil {
		return
	}
	now := time.Now()
	currentDay := now.Format("Mon")
	currentHHMM := now.Format("15:04")
	currentMinute := now.Minute()
	currentHour := now.Hour()

	for _, r := range reminders {
		if !r.Enabled {
			continue
		}

		if len(r.Schedule.Days) > 0 {
			dayMatch := false
			for _, d := range r.Schedule.Days {
				if strings.EqualFold(d, currentDay) {
					dayMatch = true
					break
				}
			}
			if !dayMatch {
				continue
			}
		}

		if r.Schedule.Type == "hourly" || r.Schedule.Type == "random_interval" {
			if len(r.Schedule.Hours) > 0 {
				hourMatch := false
				for _, h := range r.Schedule.Hours {
					if h == currentHour {
						hourMatch = true
						break
					}
				}
				if !hourMatch {
					continue
				}
			}
		}

		shouldRun := false
		var runKey string
		switch r.Schedule.Type {
		case "daily":
			if r.Schedule.Time == currentHHMM {
				runKey = r.ID + "_" + now.Format("2006-01-02") + "_" + currentHHMM
				if lastManualRun[r.ID] != runKey {
					shouldRun = true
				}
			}
		case "hourly":
			if r.Schedule.Minute == currentMinute {
				runKey = r.ID + "_" + now.Format("2006-01-02-15") + "_" + fmt.Sprint(currentMinute)
				if lastManualRun[r.ID] != runKey {
					shouldRun = true
				}
			}
		case "random_interval":
			shouldRun = handleRandomInterval(r.ID, r.Schedule.MinMinutes, r.Schedule.MaxMinutes)
		}

		if shouldRun {
			select {
			case taskQueue <- Task{
				ID:                  r.ID,
				RobotESN:            esn,
				Phrases:             r.Phrases,
				Image:               r.Image,
				Source:              "manual",
				RequireConfirmation: r.RequireConfirmation,
				SnoozeMinutes:       r.SnoozeMinutes,
			}:
				logger.Println("Productivity: Scheduled manual task " + r.ID)
				if runKey != "" {
					lastManualRun[r.ID] = runKey
				}
			default:
				logger.Println("Productivity: Queue full, skipping task " + r.ID)
			}
		}
	}
}

func handleRandomInterval(id string, min int, max int) bool {
	now := time.Now()
	nextRun, exists := nextRandomRun[id]
	if !exists {
		nextRandomRun[id] = calculateNextRandomTime(min, max)
		return false
	}
	if now.After(nextRun) {
		nextRandomRun[id] = calculateNextRandomTime(min, max)
		return true
	}
	return false
}

func calculateNextRandomTime(min int, max int) time.Time {
	if min < 1 {
		min = 1
	}
	if max < min {
		max = min
	}
	interval := min
	if max > min {
		interval = min + rand.Intn(max-min)
	}
	return time.Now().Add(time.Duration(interval) * time.Minute)
}