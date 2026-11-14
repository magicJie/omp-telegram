package telegram

import (
	"encoding/json"
	"strings"
	"testing"
)

// A forwarded update arrives over the wire with forward_origin and via_bot;
// the raw JSON stored in inbox.raw (json.Marshal(u)) must still contain them.
func TestMessageForwardFieldsRoundTrip(t *testing.T) {
	raw := `{
		"update_id": 1,
		"message": {
			"message_id": 42,
			"from": {"id": 7501215763, "username": "jay", "is_bot": false},
			"chat": {"id": 100, "type": "supergroup"},
			"text": "hi",
			"forward_origin": {
				"type": "hidden_user",
				"date": 1731600000,
				"sender_user_name": "Ghost"
			},
			"via_bot": {"id": 9285871424, "username": "relaybot", "is_bot": true}
		}
	}`

	var u Update
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	m := u.Message
	if m == nil || m.ForwardOrigin == nil {
		t.Fatal("forward_origin was dropped on unmarshal")
	}
	if m.ForwardOrigin.Type != "hidden_user" || m.ForwardOrigin.SenderUserName != "Ghost" {
		t.Fatalf("forward_origin mismatch: %+v", m.ForwardOrigin)
	}
	if m.ViaBot == nil || m.ViaBot.ID != 9285871424 || !m.ViaBot.IsBot {
		t.Fatalf("via_bot mismatch: %+v", m.ViaBot)
	}

	out, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(out)
	for _, want := range []string{`"forward_origin"`, `"hidden_user"`, `"via_bot"`, `9285871424`} {
		if !strings.Contains(s, want) {
			t.Fatalf("marshalled raw missing %s: %s", want, s)
		}
	}
}

// Legacy forward fields (Bot API < 7) are preserved too, for historical
// migrated messages that still carry forward_from / forward_from_chat.
func TestMessageLegacyForwardFieldsRoundTrip(t *testing.T) {
	raw := `{
		"message_id": 1,
		"chat": {"id": 1, "type": "private"},
		"forward_from": {"id": 9, "username": "old", "is_bot": false},
		"forward_from_chat": {"id": 88, "type": "channel"}
	}`

	var m Message
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m.ForwardFrom == nil || m.ForwardFrom.ID != 9 {
		t.Fatalf("forward_from dropped: %+v", m.ForwardFrom)
	}
	if m.ForwardFromChat == nil || m.ForwardFromChat.ID != 88 {
		t.Fatalf("forward_from_chat dropped: %+v", m.ForwardFromChat)
	}
}

// Messages without forward/via_bot fields must round-trip unchanged: the new
// fields are omitted and do not alter existing behaviour.
func TestMessageWithoutForwardFieldsUnchanged(t *testing.T) {
	m := Message{
		MessageID: 7,
		From:      &User{ID: 1, Username: "a", IsBot: false},
		Chat:      Chat{ID: 2, Type: "private"},
		Text:      "plain",
	}
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(out)
	for _, absent := range []string{"forward_origin", "forward_from", "forward_from_chat", "via_bot"} {
		if strings.Contains(s, absent) {
			t.Fatalf("unexpected %s in plain message: %s", absent, s)
		}
	}
	if m.ForwardOrigin != nil || m.ForwardFrom != nil || m.ForwardFromChat != nil || m.ViaBot != nil {
		t.Fatal("forward fields unexpectedly set on plain message")
	}
}
