package bot

import (
	"context"
	"errors"
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"telkomsel-bot/telkomsel"
)

func (h *Handler) cbShowProfile(b *gotgbot.Bot, chatID, msgID, userID int64) {
	h.editMsg(b, chatID, msgID, "⏳ Mengambil profil...", nil)

	session, ok := h.checkSession(b, chatID, msgID, userID)
	if !ok {
		return
	}

	apiCtx := context.Background()
	profile, err := h.api.GetFullProfile(apiCtx, session)
	if err != nil {
		if errors.Is(err, telkomsel.ErrUnauthorized) {
			h.sessions.Delete(userID)
			h.showExpiredLogin(b, chatID, msgID, "⚠️ Sesi expired!")
			return
		}
		kb := kbBack("back_profile")
		h.editMsg(b, chatID, msgID, fmt.Sprintf("❌ Gagal: %s", err.Error()), &kb)
		return
	}

	text := telkomsel.FormatProfile(profile) + "\nPilih aksi:"
	kb := kbProfile()
	h.editMsg(b, chatID, msgID, text, &kb)
}

func (h *Handler) cbShowMenu(b *gotgbot.Bot, chatID, msgID, userID int64) {
	h.editMsg(b, chatID, msgID, "⏳ Mengambil paket rekomendasi...", nil)

	session, ok := h.checkSession(b, chatID, msgID, userID)
	if !ok {
		return
	}

	apiCtx := context.Background()
	offers, err := h.api.GetRecommendedOffers(apiCtx, session)
	if err != nil {
		kb := kbMenu()
		h.editMsg(b, chatID, msgID, "📦 Pilih paket:", &kb)
		return
	}

	text := fmt.Sprintf("📦 *Paket Rekomendasi* (%d paket)\nPilih paket:", len(offers))
	kb := kbMenuWithOffers(offers)
	h.editMsg(b, chatID, msgID, text, &kb)
}

func (h *Handler) cbCheckQuota(b *gotgbot.Bot, chatID, msgID, userID int64) {
	h.editMsg(b, chatID, msgID, "⏳ Mengambil kuota...", nil)

	session, ok := h.checkSession(b, chatID, msgID, userID)
	if !ok {
		return
	}

	apiCtx := context.Background()
	quota, err := h.api.CheckQuota(apiCtx, session)
	if err != nil {
		if errors.Is(err, telkomsel.ErrUnauthorized) {
			h.sessions.Delete(userID)
			h.showExpiredLogin(b, chatID, msgID, "⚠️ Sesi expired!")
			return
		}
		kb := kbBack("back_profile")
		h.editMsg(b, chatID, msgID, fmt.Sprintf("❌ Gagal: %s", err.Error()), &kb)
		return
	}

	text := telkomsel.FormatQuota(quota)
	kb := kbBack("back_profile")
	h.editMsg(b, chatID, msgID, text, &kb)
}
