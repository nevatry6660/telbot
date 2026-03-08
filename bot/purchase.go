package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"telkomsel-bot/telkomsel"
)

func (h *Handler) cbShowPayment(b *gotgbot.Bot, chatID, msgID, userID int64, offerID string) {
	_, ok := h.checkSession(b, chatID, msgID, userID)
	if !ok {
		return
	}

	kb := kbPaymentSelect(offerID)
	h.editMsg(b, chatID, msgID, "💳 Pembayaran Via:", &kb)
}

func (h *Handler) cbBuy(b *gotgbot.Bot, chatID, msgID, userID int64, paymentMethod, offerID string) {
	h.editMsg(b, chatID, msgID, "🔄 Mengambil info paket...", nil)

	session, ok := h.checkSession(b, chatID, msgID, userID)
	if !ok {
		return
	}

	apiCtx := context.Background()

	details, err := h.api.GetPackageDetails(apiCtx, session, offerID)
	if err != nil {
		if errors.Is(err, telkomsel.ErrUnauthorized) {
			h.sessions.Delete(userID)
			h.showExpiredLogin(b, chatID, msgID, "⚠️ Sesi expired!")
			return
		}
		kb := kbBack("buy")
		h.editMsg(b, chatID, msgID, fmt.Sprintf("❌ Gagal: %s", err.Error()), &kb)
		return
	}

	session.PendingOfferID = offerID
	session.PendingPayment = paymentMethod
	h.sessions.Set(userID, session)

	text := telkomsel.FormatPackageDetails(details) + "\n💳 Bayar: *" + paymentMethod + "*"

	if paymentMethod == "AIRTIME" {
		balance, _, balErr := h.api.GetBalance(apiCtx, session)
		if balErr == nil {
			bal, _ := strconv.ParseInt(balance, 10, 64)
			price, _ := strconv.ParseInt(details.Price, 10, 64)
			if bal < price {
				kb := kbBack("buy")
				h.editMsg(b, chatID, msgID, fmt.Sprintf("❌ Pulsa tidak cukup!\n\n💰 Pulsa: Rp%s\n💵 Harga: Rp%s", balance, details.Price), &kb)
				return
			}
			text += fmt.Sprintf("\n💰 Pulsa: Rp%s", balance)
		}
	}

	text += "\n\n❓ Apakah anda ingin melanjutkan untuk membeli paket ini?"
	kb := kbConfirmBuy()
	h.editMsg(b, chatID, msgID, text, &kb)
}

func (h *Handler) cbConfirmBuy(b *gotgbot.Bot, chatID, msgID, userID int64) {
	session, ok := h.checkSession(b, chatID, msgID, userID)
	if !ok {
		return
	}

	offerID := session.PendingOfferID
	paymentMethod := session.PendingPayment

	if paymentMethod == "" {
		kb := kbMenu()
		h.editMsg(b, chatID, msgID, "⚠️ Tidak ada pembelian yang tertunda.", &kb)
		return
	}

	session.PendingOfferID = ""
	session.PendingPayment = ""
	h.sessions.Set(userID, session)

	h.editMsg(b, chatID, msgID, "🛒 Memproses pembelian...", nil)

	apiCtx := context.Background()

	result, err := h.api.BuyIlmupedia(apiCtx, session, offerID, paymentMethod)
	if err != nil {
		if errors.Is(err, telkomsel.ErrUnauthorized) {
			h.sessions.Delete(userID)
			h.showExpiredLogin(b, chatID, msgID, "⚠️ Sesi expired!")
			return
		}
		kb := kbBack("buy")
		h.editMsg(b, chatID, msgID, fmt.Sprintf("❌ Gagal beli: %s", err.Error()), &kb)
		return
	}

	kb := kbBack("back_profile")
	if result.QRURL != "" {
		h.editMsg(b, chatID, msgID, fmt.Sprintf("✅ Order dibuat!\n🆔 `%s`\n\n📱 Scan QRIS di bawah...", result.OrderID), nil)

		caption := fmt.Sprintf("📱 Scan QRIS untuk bayar\n🆔 Order: `%s`", result.OrderID)
		_, sendErr := b.SendPhoto(chatID, gotgbot.InputFileByURL(result.QRURL), &gotgbot.SendPhotoOpts{
			Caption:     caption,
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		if sendErr != nil {
			log.Printf("[Buy] Failed to send QRIS image: %v", sendErr)
			h.editMsg(b, chatID, msgID, telkomsel.FormatPurchaseResult(result, paymentMethod), &kb)
		}

		go h.pollPaymentStatus(b, chatID, userID, result.OrderID)
	} else {
		h.editMsg(b, chatID, msgID, telkomsel.FormatPurchaseResult(result, paymentMethod), &kb)
	}
}

func (h *Handler) pollPaymentStatus(b *gotgbot.Bot, chatID, userID int64, orderID string) {
	session := h.sessions.Get(userID)
	if session == nil {
		return
	}

	log.Printf("[Payment] Starting auto-poll for order %s", orderID)

	apiCtx := context.Background()
	status, err := h.api.PollPaymentStatus(apiCtx, session, orderID, 60, 5*time.Second)

	var msg string
	if err != nil {
		msg = fmt.Sprintf("⏰ Pembayaran timeout/gagal\n🆔 Order: `%s`\n\n%v", orderID, err)
	} else {
		switch status.Status {
		case "paid":
			msg = fmt.Sprintf("✅ Pembayaran berhasil!\n🆔 Order: `%s`", orderID)
		case "expired":
			msg = fmt.Sprintf("⏰ Pembayaran expired\n🆔 Order: `%s`", orderID)
		case "cancelled":
			msg = fmt.Sprintf("🚫 Pembayaran dibatalkan\n🆔 Order: `%s`", orderID)
		default:
			msg = fmt.Sprintf("❓ Status: %s\n🆔 Order: `%s`", status.Status, orderID)
		}
	}

	kb := kbProfile()
	_, _ = b.SendMessage(chatID, msg, &gotgbot.SendMessageOpts{
		ParseMode:   "Markdown",
		ReplyMarkup: kb,
	})

	log.Printf("[Payment] Poll complete for order %s", orderID)
}
