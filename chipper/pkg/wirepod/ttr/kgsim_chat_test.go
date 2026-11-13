package wirepod_ttr

import (
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/sashabaranov/go-openai"
)

// a saved image message has parts and no text; it must stay in the context, while a truly empty message goes
func TestRememberedImageMessagesStayInContext(t *testing.T) {
	oldChats, oldSave := vars.RememberedChats, vars.APIConfig.Knowledge.SaveChat
	t.Cleanup(func() { vars.RememberedChats, vars.APIConfig.Knowledge.SaveChat = oldChats, oldSave })
	vars.APIConfig.Knowledge.SaveChat = true
	vars.RememberedChats = []vars.RememberedChat{{ESN: "aaaa1111", Chats: []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, MultiContent: []openai.ChatMessagePart{{Type: openai.ChatMessagePartTypeText, Text: "what is this?"}}},
		{Role: openai.ChatMessageRoleAssistant, Content: "a cube"},
		{Role: openai.ChatMessageRoleAssistant, Content: "  "},
	}}}
	msgs := CreateAIReq("and now?", "aaaa1111", false, false).Messages
	var parts, empty int
	for _, m := range msgs {
		if len(m.MultiContent) > 0 {
			parts++
		}
		if m.Content == "  " {
			empty++
		}
	}
	if parts != 1 || empty != 0 {
		t.Fatalf("want the image message kept and the blank one dropped, got %d and %d in %+v", parts, empty, msgs)
	}
}
