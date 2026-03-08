package telkomsel

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"telkomsel-bot/config"
	"telkomsel-bot/model"
)

type RecommendedOffer struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Price         string  `json:"price"`
	ProductLength string  `json:"productlength"`
	HighlightVal  string  `json:"highlightvalue"`
	Category      string  `json:"category"`
	Subcategory   string  `json:"subcategory"`
	IsLoan        bool    `json:"isLoan"`
	IsSubscribe   bool    `json:"isSubscribe"`
	Bonuses       []Bonus `json:"bonus"`
}

type Bonus struct {
	Name     string `json:"name"`
	Quota    string `json:"quota"`
	Class    string `json:"class"`
	Validity string `json:"validity"`
}

func (c *Client) GetRecommendedOffers(ctx context.Context, session *model.Session) ([]RecommendedOffer, error) {
	body := map[string]interface{}{
		"isPrepaid":               true,
		"page":                    "dashboard",
		"isCorporate":             false,
		"fetchSpecialAreaPackage": true,
		"isRoaming":               false,
	}

	resp, err := c.doPost(ctx, config.OffersEndpoint, session, body)
	if err != nil {
		return nil, err
	}

	if resp.Status != "00000" {
		return nil, fmt.Errorf("failed to get offers: %s", resp.Message)
	}

	var data struct {
		OfferGroup []struct {
			Name   string             `json:"name"`
			Offers []RecommendedOffer `json:"offer"`
		} `json:"offerGroup"`
	}

	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parse recommended offers: %w", err)
	}

	seen := make(map[string]bool)
	var all []RecommendedOffer
	for _, group := range data.OfferGroup {
		for _, o := range group.Offers {
			if !seen[o.ID] {
				seen[o.ID] = true
				all = append(all, o)
			}
		}
	}

	return all, nil
}

func FormatRecommendedOffers(offers []RecommendedOffer) string {
	if len(offers) == 0 {
		return "❌ Tidak ada paket rekomendasi yang tersedia."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📦 *Paket Rekomendasi* (%d paket)\n\n", len(offers)))

	for i, o := range offers {
		sb.WriteString(fmt.Sprintf("*%d. %s*\n", i+1, o.Name))
		sb.WriteString(fmt.Sprintf("   💰 Rp%s • ⏳ %s\n", o.Price, o.ProductLength))
		if o.HighlightVal != "" {
			sb.WriteString(fmt.Sprintf("   📊 %s\n", o.HighlightVal))
		}

		if len(o.Bonuses) > 0 {
			sb.WriteString("   🎁 Bonus: ")
			parts := make([]string, 0, len(o.Bonuses))
			for _, b := range o.Bonuses {
				parts = append(parts, fmt.Sprintf("%s %s", b.Name, b.Quota))
			}
			sb.WriteString(strings.Join(parts, ", "))
			sb.WriteString("\n")
		}

		if o.IsLoan {
			sb.WriteString("   🏷️ Bayar Nanti (Loan)\n")
		}
		if o.IsSubscribe {
			sb.WriteString("   🔄 Berlangganan\n")
		}

		sb.WriteString(fmt.Sprintf("   \U0001F194 ID: `%s`\n", o.ID))
		sb.WriteString("\n")
	}

	sb.WriteString("Gunakan ID paket di atas untuk membeli dengan `buy_package`.")
	return sb.String()
}
