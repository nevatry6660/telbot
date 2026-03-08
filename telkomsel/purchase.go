package telkomsel

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"telkomsel-bot/config"
	"telkomsel-bot/model"
)

type PurchaseResult struct {
	OrderID  string
	Status   string
	Message  string
	QRURL    string
	Eligible bool
	RawData  json.RawMessage
}

type PaymentStatus struct {
	Status  string
	Message string
}

type PackageDetails struct {
	ID        string
	Name      string
	ShortDesc string
	LongDesc  string
	Price     string
	Validity  string
	Quota     string
}

func (c *Client) GetPackageDetails(ctx context.Context, session *model.Session, offerID string) (*PackageDetails, error) {
	if offerID == "" {
		offerID = config.OfferId
	}

	endpoint := fmt.Sprintf("/api/paket-details/v2/%s", offerID)

	resp, err := c.doGet(ctx, endpoint, session)
	if err != nil {
		return nil, err
	}

	if resp.Status != "00000" {
		return nil, fmt.Errorf("failed to get details: %s", resp.Message)
	}

	var data struct {
		Offer struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			ShortDesc     string `json:"shortdesc"`
			LongDesc      string `json:"longdesc"`
			Price         string `json:"price"`
			ProductLength string `json:"productlength"`
			Highlight     string `json:"highlightvalue"`
		} `json:"offer"`
	}

	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parse package details: %w", err)
	}

	return &PackageDetails{
		ID:        data.Offer.ID,
		Name:      data.Offer.Name,
		ShortDesc: data.Offer.ShortDesc,
		LongDesc:  data.Offer.LongDesc,
		Price:     data.Offer.Price,
		Validity:  data.Offer.ProductLength,
		Quota:     data.Offer.Highlight,
	}, nil
}

func (c *Client) BuyIlmupedia(ctx context.Context, session *model.Session, offerID, paymentMethod string) (*PurchaseResult, error) {
	if offerID == "" {
		offerID = config.OfferId
	}
	if paymentMethod == "" {
		paymentMethod = config.DefaultPayment
	}

	body := map[string]interface{}{
		"mode":              "SELF",
		"productType":       "PACKAGE",
		"msisdninitiator":   session.FullPhone,
		"toBeSubscribedTo":  false,
		"platform":          "web",
		"type":              "purchase",
		"offerID":           offerID,
		"paymentMethod":     paymentMethod,
		"businessproductid": offerID,
		"isCampaignOffer":   "false",
		"cart_id":           "",
	}

	resp, err := c.doPost(ctx, config.BuyEndpoint, session, body)
	if err != nil {
		return nil, err
	}

	result := &PurchaseResult{
		Status:  resp.Status,
		Message: resp.Message,
	}

	if resp.Status != "00000" {
		return result, fmt.Errorf("purchase failed: %s", resp.Message)
	}

	var orderData struct {
		OrderID           string `json:"orderid"`
		TransactionStatus string `json:"transactionstatus"`
		TransactionDesc   string `json:"transactiondesc"`
		IsEligible        string `json:"iseligible"`
		Reason            string `json:"reason"`
		QRURL             string `json:"qrURL"`
	}

	if err := json.Unmarshal(resp.Data, &orderData); err == nil {
		result.OrderID = orderData.OrderID
		result.QRURL = orderData.QRURL
		result.Eligible = orderData.IsEligible == "true"
		result.RawData = resp.Data
	}

	return result, nil
}

func (c *Client) CheckPaymentStatus(ctx context.Context, session *model.Session, orderID string) (*PaymentStatus, error) {
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"payment_transaction_id": orderID,
			"type":                   "package",
		},
	}

	resp, err := c.doPost(ctx, config.StatusEndpoint, session, body)
	if err != nil {
		return nil, err
	}

	status := &PaymentStatus{
		Message: resp.Message,
	}

	if resp.Status == "00000" {
		var statusData struct {
			Payment struct {
				Status string `json:"status"`
			} `json:"payment"`
		}
		if err := json.Unmarshal(resp.Data, &statusData); err == nil {
			status.Status = statusData.Payment.Status
		}
	}

	return status, nil
}

func (c *Client) PollPaymentStatus(ctx context.Context, session *model.Session, orderID string, maxPolls int, interval time.Duration) (*PaymentStatus, error) {
	if maxPolls == 0 {
		maxPolls = 60
	}
	if interval == 0 {
		interval = 5 * time.Second
	}

	for i := 0; i < maxPolls; i++ {
		status, err := c.CheckPaymentStatus(ctx, session, orderID)
		if err != nil {
			time.Sleep(interval)
			continue
		}

		switch status.Status {
		case "paid":
			return status, nil
		case "expired", "cancelled":
			return status, fmt.Errorf("payment %s", status.Status)
		}

		time.Sleep(interval)
	}

	return nil, fmt.Errorf("payment polling timeout after %d attempts", maxPolls)
}

func FormatPackageDetails(pkg *PackageDetails) string {
	if pkg == nil {
		return "❌ Detail paket tidak ditemukan."
	}

	msg := "ℹ️ *Informasi Paket*\n\n"
	msg += fmt.Sprintf("📦 *Nama:* %s\n", pkg.Name)
	msg += fmt.Sprintf("💰 *Harga:* Rp%s\n", pkg.Price)
	msg += fmt.Sprintf("⏳ *Masa Aktif:* %s\n", pkg.Validity)
	if pkg.Quota != "" {
		msg += fmt.Sprintf("📊 *Kuota Utama:* %s\n", pkg.Quota)
	}
	desc := stripHTML(pkg.LongDesc)
	if desc != "" {
		msg += fmt.Sprintf("\n📝 *Deskripsi:*\n%s\n\n", desc)
	}

	return msg
}

func FormatPurchaseResult(result *PurchaseResult, paymentMethod string) string {
	if result == nil {
		return "❌ No purchase result available."
	}

	msg := "🛒 *Ilmupedia Purchase Result*\n\n"
	msg += fmt.Sprintf("📋 Status: %s\n", result.Message)

	if result.OrderID != "" {
		msg += fmt.Sprintf("🆔 Order ID: `%s`\n", result.OrderID)
	}

	if result.QRURL != "" {
		msg += fmt.Sprintf("\n📱 *Scan QR to pay (QRIS):*\n%s\n", result.QRURL)
		msg += "\nUse `/status <order-id>` to check payment status."
	} else if paymentMethod == "AIRTIME" && result.OrderID != "" {
		msg += "\n💰 *Pembayaran berhasil menggunakan Pulsa.*\n"
	}

	return msg
}

func stripHTML(s string) string {

	s = strings.ReplaceAll(s, "<li>", "• ")
	s = strings.ReplaceAll(s, "</li>", "\n")

	re := regexp.MustCompile(`<[^>]*>`)
	s = re.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	return s
}
