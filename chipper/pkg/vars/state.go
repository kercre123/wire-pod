package vars

import (
	"slices"
	"sync"
)

var (
	chatsMu         sync.Mutex
	rememberedChats []RememberedChat

	rinfoMu       sync.Mutex
	recurringInfo []RecurringInfoStore
)

func GetRememberedChat(esn string) RememberedChat {
	chatsMu.Lock()
	defer chatsMu.Unlock()
	for _, chat := range rememberedChats {
		if chat.ESN == esn {
			chat.Chats = slices.Clone(chat.Chats)
			return chat
		}
	}
	return RememberedChat{ESN: esn}
}

func UpdateRememberedChat(esn string, fn func(chat *RememberedChat)) {
	chatsMu.Lock()
	defer chatsMu.Unlock()
	for i := range rememberedChats {
		if rememberedChats[i].ESN == esn {
			fn(&rememberedChats[i])
			return
		}
	}
	chat := RememberedChat{ESN: esn}
	fn(&chat)
	rememberedChats = append(rememberedChats, chat)
}

func ClearRememberedChats() {
	chatsMu.Lock()
	rememberedChats = nil
	chatsMu.Unlock()
}

func FindRecurringInfoByID(id string) (RecurringInfoStore, bool) {
	rinfoMu.Lock()
	defer rinfoMu.Unlock()
	for _, rinf := range recurringInfo {
		if rinf.ID == id {
			return rinf, true
		}
	}
	return RecurringInfoStore{}, false
}

func AddToRInfo(esn string, id string, ip string) {
	rinfoMu.Lock()
	defer rinfoMu.Unlock()
	for i := range recurringInfo {
		if recurringInfo[i].ESN == esn {
			recurringInfo[i].ID = id
			recurringInfo[i].IP = ip
			return
		}
	}
	recurringInfo = append(recurringInfo, RecurringInfoStore{ESN: esn, ID: id, IP: ip})
}
