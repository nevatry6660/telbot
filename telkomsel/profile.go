package telkomsel

import (
	"context"
	"encoding/json"
	"fmt"

	"telkomsel-bot/model"
)

type ProfileInfo struct {
	Name          string
	Email         string
	Phone         string
	LoyaltyPoints string
	LoyaltyTier   string
	Balance       string
	BalanceExpiry string
	AccountStatus string
}

func (c *Client) GetProfile(ctx context.Context, session *model.Session) (*ProfileInfo, error) {
	resp, err := c.doGet(ctx, "/api/attributes/getprofile", session)
	if err != nil {
		return nil, err
	}

	if resp.Status != "00000" {
		return nil, fmt.Errorf("profile API error: %s", resp.Message)
	}

	var data struct {
		Name struct {
			FirstName string `json:"firstName"`
			SurrName  string `json:"surrName"`
		} `json:"name"`
		Email struct {
			Email    string `json:"email"`
			IsVerify bool   `json:"isVerify"`
		} `json:"email"`
		CIAM struct {
			AccountStatus string `json:"accountStatus"`
			GivenName     string `json:"givenName"`
		} `json:"ciam"`
	}

	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parse profile: %w", err)
	}

	name := data.Name.FirstName
	if data.Name.SurrName != "" {
		name += " " + data.Name.SurrName
	}
	if name == "" {
		name = data.CIAM.GivenName
	}

	return &ProfileInfo{
		Name:          name,
		Email:         data.Email.Email,
		Phone:         session.FullPhone,
		AccountStatus: data.CIAM.AccountStatus,
	}, nil
}

func (c *Client) GetLoyaltyInfo(ctx context.Context, session *model.Session) (points, tier string, err error) {
	resp, err := c.doGet(ctx, "/api/subscriber/loyalty-info", session)
	if err != nil {
		return "", "", err
	}

	if resp.Status != "00000" {
		return "", "", fmt.Errorf("loyalty API error: %s", resp.Message)
	}

	var data struct {
		Profiles struct {
			LoyaltyPoints   string `json:"loyalty_points"`
			LoyaltyCategory string `json:"loyalty_points_category"`
		} `json:"profiles"`
	}

	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", "", fmt.Errorf("parse loyalty: %w", err)
	}

	return data.Profiles.LoyaltyPoints, data.Profiles.LoyaltyCategory, nil
}

func (c *Client) GetBalance(ctx context.Context, session *model.Session) (balance, expiry string, err error) {
	resp, err := c.doGet(ctx, "/api/subscriber/profile-balance", session)
	if err != nil {
		return "", "", err
	}

	if resp.Status != "00000" {
		return "", "", fmt.Errorf("balance API error: %s", resp.Message)
	}

	var data struct {
		Balance    string `json:"balance"`
		ExpiryDate string `json:"expiry_date"`
	}

	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", "", fmt.Errorf("parse balance: %w", err)
	}

	return data.Balance, data.ExpiryDate, nil
}

func (c *Client) GetFullProfile(ctx context.Context, session *model.Session) (*ProfileInfo, error) {
	profile, err := c.GetProfile(ctx, session)
	if err != nil {
		return nil, err
	}

	points, tier, err := c.GetLoyaltyInfo(ctx, session)
	if err == nil {
		profile.LoyaltyPoints = points
		profile.LoyaltyTier = tier
	}

	balance, expiry, err := c.GetBalance(ctx, session)
	if err == nil {
		profile.Balance = balance
		profile.BalanceExpiry = expiry
	}

	return profile, nil
}

func FormatProfile(p *ProfileInfo) string {
	if p == nil {
		return ""
	}

	msg := "👤 *Profil*\n\n"
	msg += fmt.Sprintf("• *Nama:* %s\n", p.Name)
	msg += fmt.Sprintf("• *Nomor:* +%s\n", p.Phone)
	if p.Email != "" {
		msg += fmt.Sprintf("• *Email:* %s\n", p.Email)
	}
	if p.Balance != "" {
		msg += fmt.Sprintf("• *Pulsa:* Rp%s\n", p.Balance)
	}
	if p.BalanceExpiry != "" {
		msg += fmt.Sprintf("• *Masa Aktif:* %s\n", p.BalanceExpiry)
	}
	if p.LoyaltyTier != "" {
		msg += fmt.Sprintf("• *Tier:* %s\n", p.LoyaltyTier)
	}
	if p.LoyaltyPoints != "" {
		msg += fmt.Sprintf("• *Poin:* %s\n", p.LoyaltyPoints)
	}

	return msg
}
