package bot

import (
	"context"
	"errors"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"telkomsel-bot/model"
	"telkomsel-bot/telkomsel"
)

func (h *Handler) editMsg(b *gotgbot.Bot, chatID, msgID int64, text string, kb *gotgbot.InlineKeyboardMarkup) {
	opts := &gotgbot.EditMessageTextOpts{
		ChatId:    chatID,
		MessageId: msgID,
		ParseMode: "Markdown",
	}
	if kb != nil {
		opts.ReplyMarkup = *kb
	}
	_, _, _ = b.EditMessageText(text, opts)
}

func (h *Handler) checkSession(b *gotgbot.Bot, chatID, msgID, userID int64) (*model.Session, bool) {
	session := h.sessions.Get(userID)
	if session == nil || !session.IsLoggedIn() {
		h.showExpiredLogin(b, chatID, msgID, "❌ Belum login.")
		return nil, false
	}

	apiCtx := context.Background()
	_, _, err := h.api.GetBalance(apiCtx, session)
	if err != nil && errors.Is(err, telkomsel.ErrUnauthorized) {
		h.sessions.Delete(userID)
		h.showExpiredLogin(b, chatID, msgID, "⚠️ Sesi expired! Login ulang.")
		return nil, false
	}

	return session, true
}

func (h *Handler) showExpiredLogin(b *gotgbot.Bot, chatID, msgID int64, text string) {
	kb := kbLogin()
	h.editMsg(b, chatID, msgID, text, &kb)
}
