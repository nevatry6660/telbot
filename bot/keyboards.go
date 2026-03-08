package bot

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
)

func kbLogin() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{{Text: "Login", CallbackData: "login"}},
		},
	}
}

func kbProfile() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "🛒 Beli Paket", CallbackData: "buy"},
				{Text: "📊 Cek Kouta", CallbackData: "check_quota"},
			},
			{{Text: "🔄 Refresh", CallbackData: "refresh"}},
			{{Text: "👋 Logout", CallbackData: "logout"}},
		},
	}
}

func kbMenu() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "📚 Ilmupedia", CallbackData: "pkg_ilmupedia"},
				{Text: "🆔 Custom Id", CallbackData: "pkg_custom"},
			},
			{{Text: "🤖 Beli Otomatis", CallbackData: "auto_buy"}},
			{{Text: "🔙 Kembali", CallbackData: "back_profile"}},
		},
	}
}

func kbPaymentSelect(offerID string) gotgbot.InlineKeyboardMarkup {
	qrisData := "pay_qris"
	pulsaData := "pay_pulsa"
	if offerID != "" {
		qrisData = "pay_qris_" + offerID
		pulsaData = "pay_pulsa_" + offerID
	}
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "QRIS", CallbackData: qrisData},
				{Text: "💰 PULSA", CallbackData: pulsaData},
			},
			{{Text: "🔙 Kembali", CallbackData: "buy"}},
		},
	}
}

func kbAutoMonitor() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "⏱ 20 Menit", CallbackData: "auto_20"},
				{Text: "⏱ 50 Menit", CallbackData: "auto_50"},
			},
			{{Text: "⌨️ Custom", CallbackData: "auto_custom"}},
			{{Text: "🔙 Kembali", CallbackData: "buy"}},
		},
	}
}

func kbAutoPackage() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "📚 Ilmupedia", CallbackData: "auto_pkg_ilmupedia"},
				{Text: "🆔 Custom Id", CallbackData: "auto_pkg_custom"},
			},
			{{Text: "🔙 Kembali", CallbackData: "auto_buy"}},
		},
	}
}

func kbAutoPay() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{{Text: "💰 Pulsa", CallbackData: "auto_pay_pulsa"}},
			{{Text: "🔙 Kembali", CallbackData: "auto_buy"}},
		},
	}
}

func kbAutoRunning() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{{Text: "🛑 Stop Monitor", CallbackData: "auto_stop"}},
			{{Text: "🔙 Kembali", CallbackData: "back_profile"}},
		},
	}
}

func kbConfirmBuy() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "✅ Ya", CallbackData: "confirm_buy"},
				{Text: "❌ Tidak", CallbackData: "buy"},
			},
		},
	}
}

func kbBack(target string) gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{{Text: "🔙 Kembali", CallbackData: target}},
		},
	}
}
