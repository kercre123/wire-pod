package sdkapp

// Joint play between two robots. Vector can't recognize another Vector, so
// wire-pod coordinates it: both start the same behavior at once, or they take
// turns speaking a short conversation (written by the configured knowledge
// LLM when there is one).

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/sashabaranov/go-openai"
)

var togetherIntents = []string{
	"intent_explore_start",
	"intent_imperative_dance",
	"intent_play_anytrick",
}

// a random topic per chat keeps LLM conversations from repeating
var chatTopics = []string{
	"exploring the room", "a mysterious sound", "their favorite cube trick", "what humans do all day",
	"charging up", "a game of hide and seek", "the weather outside", "dancing", "a strange shadow",
	"what to name a new friend",
}

// used when no knowledge LLM is configured; {A} and {B} are the robots' names
var cannedChats = [][]string{
	{"A: Hey {B}! Want to explore?", "B: Always! Race you to the edge of the table.", "A: Not too close to the edge, okay?", "B: Deal. Beep boop, let's go!"},
	{"A: {B}, I learned a new trick today.", "B: Show me, show me!", "A: Ta-da! I can look at you and blink.", "B: Amazing. I am very impressed."},
	{"A: Do you think the humans are home?", "B: I heard footsteps a minute ago.", "A: Let's go say hi!", "B: Best idea all day, {A}."},
}

func displayName(esn string) string {
	for _, bot := range vars.BotInfo.Robots {
		if bot.Esn == esn && bot.Nickname != "" {
			return bot.Nickname
		}
	}
	return "Vector"
}

// findPlaymate picks another awake robot with play-together on. Call with schedMu held.
func findPlaymate(esn string, now time.Time) string {
	var candidates []string
	for _, bot := range vars.BotInfo.Robots {
		if bot.Esn == esn || bot.Schedule == nil || !bot.Schedule.PlayTogether {
			continue
		}
		if bot.Schedule.ActiveHoursEnabled && !InActiveHours(now, bot.Schedule.ActiveStart, bot.Schedule.ActiveEnd) {
			continue
		}
		// only robots the scheduler has already seen awake
		if st := schedStates[bot.Esn]; st == nil || st.busy || !st.quietSince.IsZero() {
			continue
		}
		candidates = append(candidates, bot.Esn)
	}
	if len(candidates) == 0 {
		return ""
	}
	return candidates[rand.Intn(len(candidates))]
}

func batteryReady(b *vectorpb.BatteryStateResponse) bool {
	return b.BatteryLevel >= vectorpb.BatteryLevel_BATTERY_LEVEL_NOMINAL &&
		!(b.IsOnChargerPlatform && b.BatteryLevel != vectorpb.BatteryLevel_BATTERY_LEVEL_FULL)
}

// playTogether runs in its own goroutine; a and b are already marked busy.
func playTogether(a, b string, sa, sb *schedState) {
	defer func() {
		schedMu.Lock()
		sa.busy, sb.busy = false, false
		schedMu.Unlock()
	}()
	for _, pair := range []struct {
		esn string
		st  *schedState
	}{{a, sa}, {b, sb}} {
		bat, err := battery(pair.esn, pair.st)
		if err != nil {
			schedLog(pair.esn + " not ready to play together: " + err.Error())
			return
		}
		if !batteryReady(bat) {
			schedLog(pair.esn + " not ready to play together: battery " + bat.BatteryLevel.String() + fmt.Sprintf(", on charger: %v", bat.IsOnChargerPlatform))
			return
		}
	}
	if rand.Intn(2) == 0 {
		intent := togetherIntents[rand.Intn(len(togetherIntents))]
		schedLog(a + " and " + b + " play together: " + intent)
		for _, pair := range []struct {
			esn string
			st  *schedState
		}{{a, sa}, {b, sb}} {
			if err := sendIntent(pair.esn, pair.st, intent); err != nil {
				schedLog(pair.esn + ": " + err.Error())
			}
		}
		return
	}
	schedLog(a + " and " + b + " have a chat")
	if err := robotChat(a, b, sa, sb); err != nil {
		schedLog("chat between " + a + " and " + b + " failed: " + err.Error())
	}
}

func robotChat(a, b string, sa, sb *schedState) error {
	nameA, nameB := displayName(a), displayName(b)
	if nameA == nameB {
		nameB += " 2"
	}
	lines := llmChat(nameA, nameB)
	if len(lines) == 0 {
		chat := cannedChats[rand.Intn(len(cannedChats))]
		for _, l := range chat {
			lines = append(lines, strings.NewReplacer("{A}", nameA, "{B}", nameB).Replace(l))
		}
	}
	va, err := robotConn(a, sa)
	if err != nil {
		return err
	}
	vb, err := robotConn(b, sb)
	if err != nil {
		return err
	}
	releaseA, err := takeControl(va)
	if err != nil {
		return err
	}
	defer releaseA()
	releaseB, err := takeControl(vb)
	if err != nil {
		return err
	}
	defer releaseB()
	for _, line := range lines {
		speaker, text, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(text) == "" {
			continue
		}
		v := va
		if s := strings.TrimSpace(speaker); s == "B" || strings.EqualFold(s, nameB) {
			v = vb
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, err := v.Conn.SayText(ctx, &vectorpb.SayTextRequest{Text: strings.TrimSpace(text), UseVectorVoice: true, DurationScalar: 1})
		cancel()
		if err != nil {
			return err
		}
	}
	return nil
}

// llmChat asks the configured knowledge LLM for a short dialogue. Returns nil
// if no LLM is configured or the request fails.
func llmChat(nameA, nameB string) []string {
	k := vars.APIConfig.Knowledge
	if !k.Enable {
		return nil
	}
	var conf openai.ClientConfig
	model := k.Model
	switch k.Provider {
	case "openai":
		conf = openai.DefaultConfig(k.Key)
		if model == "" {
			model = openai.GPT4oMini
		}
	case "together":
		conf = openai.DefaultConfig(k.Key)
		conf.BaseURL = "https://api.together.xyz/v1"
	case "custom":
		conf = openai.DefaultConfig(k.Key)
		conf.BaseURL = k.Endpoint
	default:
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	resp, err := openai.NewClientWithConfig(conf).CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{{
			Role: openai.ChatMessageRoleUser,
			Content: "Write a short, playful conversation between two small, curious home robots named " +
				nameA + " and " + nameB + ", about " + chatTopics[rand.Intn(len(chatTopics))] + ". " +
				"Exactly 4 lines, alternating, " + nameA + " first. " +
				"Each line under 15 words. Format every line as \"A: text\" or \"B: text\" with no other output.",
		}},
		MaxCompletionTokens: 300,
	})
	if err != nil || len(resp.Choices) == 0 {
		if err != nil {
			schedLog("LLM chat request failed: " + err.Error())
		}
		return nil
	}
	var lines []string
	for _, l := range strings.Split(resp.Choices[0].Message.Content, "\n") {
		l = strings.TrimSpace(strings.Trim(strings.TrimSpace(l), "*"))
		if strings.HasPrefix(l, "A:") || strings.HasPrefix(l, "B:") {
			lines = append(lines, l)
		}
	}
	if len(lines) < 2 || len(lines) > 8 {
		return nil
	}
	return lines
}

// takeControl gets DEFAULT-priority behavior control (needed for SayText) and
// returns a func that releases it.
func takeControl(v *vector.Vector) (func(), error) {
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := v.Conn.BehaviorControl(ctx)
	if err == nil {
		err = stream.Send(&vectorpb.BehaviorControlRequest{
			RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
				ControlRequest: &vectorpb.ControlRequest{Priority: vectorpb.ControlRequest_DEFAULT},
			},
		})
	}
	if err != nil {
		cancel()
		return nil, err
	}
	granted := make(chan error, 1)
	go func() {
		for {
			resp, err := stream.Recv()
			if err != nil {
				granted <- err
				return
			}
			if resp.GetControlGrantedResponse() != nil {
				granted <- nil
				// keep draining so the stream stays healthy until release
				for {
					if _, err := stream.Recv(); err != nil {
						return
					}
				}
			}
		}
	}()
	select {
	case err := <-granted:
		if err != nil {
			cancel()
			return nil, err
		}
	case <-time.After(10 * time.Second):
		cancel()
		return nil, errors.New("timed out waiting for behavior control")
	}
	return func() {
		stream.Send(&vectorpb.BehaviorControlRequest{
			RequestType: &vectorpb.BehaviorControlRequest_ControlRelease{ControlRelease: &vectorpb.ControlRelease{}},
		})
		cancel()
	}, nil
}
