package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/maks2134/telegram-rag-bot/internal/pipeline"
)

// Telegram runs long-polling support bot handlers.
type Telegram struct {
	api         *tgbotapi.BotAPI
	asker       *pipeline.Asker
	adminChatID int64
	mu          sync.Mutex
	logs        []string
}

// NewTelegram creates a polling bot. token must be non-empty.
func NewTelegram(token string, asker *pipeline.Asker, adminChatID int64) (*Telegram, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	api.Debug = false
	return &Telegram{
		api:         api,
		asker:       asker,
		adminChatID: adminChatID,
	}, nil
}

// Run blocks on getUpdates until ctx is cancelled.
func (t *Telegram) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := t.api.GetUpdatesChan(u)
	log.Printf("telegram: authorized as @%s (polling)", t.api.Self.UserName)

	for {
		select {
		case <-ctx.Done():
			t.api.StopReceivingUpdates()
			return ctx.Err()
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			if update.CallbackQuery != nil {
				t.handleCallback(ctx, update.CallbackQuery)
				continue
			}
			if update.Message == nil || update.Message.Text == "" {
				continue
			}
			t.handleMessage(ctx, update.Message)
		}
	}
}

func mainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Refunds"),
			tgbotapi.NewKeyboardButton("Onboarding"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("SLA"),
			tgbotapi.NewKeyboardButton("Talk to human"),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false
	return kb
}

func inlineSuggestions() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏱ Refund timing", "q:How long do refunds take?"),
			tgbotapi.NewInlineKeyboardButtonData("🚀 Onboarding", "q:How does onboarding work for new teams?"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Business SLA", "q:What is the Business plan SLA?"),
			tgbotapi.NewInlineKeyboardButtonData("👤 Human", "cmd:human"),
		),
	)
}

func (t *Telegram) handleCallback(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	data := strings.TrimSpace(cq.Data)
	callback := tgbotapi.NewCallback(cq.ID, "")
	if _, err := t.api.Request(callback); err != nil {
		log.Printf("telegram callback ack: %v", err)
	}
	if cq.Message == nil {
		return
	}
	chatID := cq.Message.Chat.ID
	user := "callback"
	if cq.From.UserName != "" {
		user = cq.From.UserName
	}

	switch {
	case strings.HasPrefix(data, "q:"):
		q := strings.TrimPrefix(data, "q:")
		t.replyAsk(ctx, chatID, user, q)
	case data == "cmd:human":
		t.sendText(chatID, t.escalate(chatID, user, "/human"), true)
	}
}

func (t *Telegram) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	text := strings.TrimSpace(msg.Text)
	user := msg.From.UserName
	if user == "" {
		user = fmt.Sprintf("id:%d", msg.From.ID)
	}
	t.appendLog(user + ": " + text)
	log.Printf("telegram message chat_id=%d from=%s text=%q", msg.Chat.ID, user, text)

	lower := strings.ToLower(text)
	chatID := msg.Chat.ID

	switch {
	case strings.HasPrefix(lower, "/start"):
		welcome := "👋 *DocPilot* — docs-grounded support\n\n" +
			"I answer only from your knowledge base and cite sources.\n\n" +
			"Try a button below, or ask in your own words.\n" +
			"Need a person\\? Tap *Talk to human* or /human\\.\n\n" +
			"_Tip:_ /id shows your chat id for ADMIN\\_CHAT\\_ID\\."
		out := tgbotapi.NewMessage(chatID, welcome)
		out.ParseMode = "Markdown"
		out.ReplyMarkup = mainKeyboard()
		if _, err := t.api.Send(out); err != nil {
			log.Printf("telegram send: %v", err)
		}
		hint := tgbotapi.NewMessage(chatID, "Quick picks:")
		hint.ReplyMarkup = inlineSuggestions()
		if _, err := t.api.Send(hint); err != nil {
			log.Printf("telegram send: %v", err)
		}
	case strings.HasPrefix(lower, "/id"):
		t.sendText(chatID, fmt.Sprintf("Your chat id: `%d`", chatID), true)
	case strings.HasPrefix(lower, "/human") || lower == "talk to human":
		t.sendText(chatID, t.escalate(chatID, user, text), true)
	case lower == "refunds":
		t.replyAsk(ctx, chatID, user, "How long do refunds take?")
	case lower == "onboarding":
		t.replyAsk(ctx, chatID, user, "How does onboarding work for new teams?")
	case lower == "sla":
		t.replyAsk(ctx, chatID, user, "What is the Business plan SLA?")
	default:
		t.replyAsk(ctx, chatID, user, text)
	}
}

func (t *Telegram) replyAsk(ctx context.Context, chatID int64, user, question string) {
	t.appendLog(user + " → ask: " + question)
	typing := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	_, _ = t.api.Request(typing)

	res := t.asker.Ask(ctx, question)
	body := pipeline.FormatTelegram(res)
	out := tgbotapi.NewMessage(chatID, body)
	out.ReplyMarkup = mainKeyboard()
	if _, err := t.api.Send(out); err != nil {
		log.Printf("telegram send: %v", err)
	}
}

func (t *Telegram) sendText(chatID int64, text string, withKeyboard bool) {
	out := tgbotapi.NewMessage(chatID, text)
	if withKeyboard {
		out.ReplyMarkup = mainKeyboard()
	}
	if strings.Contains(text, "`") {
		out.ParseMode = "Markdown"
	}
	if _, err := t.api.Send(out); err != nil {
		log.Printf("telegram send: %v", err)
	}
}

func (t *Telegram) escalate(chatID int64, user, text string) string {
	note := fmt.Sprintf("🆘 Escalation from @%s (chat %d):\n%s", user, chatID, text)
	if t.adminChatID != 0 {
		msg := tgbotapi.NewMessage(t.adminChatID, note)
		if _, err := t.api.Send(msg); err != nil {
			log.Printf("telegram admin notify: %v", err)
			return "Escalated locally, but admin notify failed. A teammate will still follow up if configured."
		}
		return "✅ Escalated. A teammate will follow up shortly."
	}
	log.Printf("escalation (no ADMIN_CHAT_ID): %s", note)
	return "Escalated. A teammate will follow up. (Set ADMIN_CHAT_ID to forward to an operator chat.)"
}

func (t *Telegram) appendLog(line string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.logs = append(t.logs, line)
	if len(t.logs) > 200 {
		t.logs = t.logs[len(t.logs)-200:]
	}
}
