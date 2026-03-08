package telkomsel

import (
	"context"
	"errors"
	"log"

	"telkomsel-bot/model"
)

func ValidateSessions(sessions *model.SessionManager, api *Client) {
	all := sessions.All()
	if len(all) == 0 {
		return
	}

	log.Printf("🔍 Validating %d saved session(s)...", len(all))
	apiCtx := context.Background()

	for userID, session := range all {
		if !session.IsLoggedIn() {
			log.Printf("  ⏭ User %d: not logged in (state=%s), skipping", userID, session.State)
			continue
		}

		_, _, err := api.GetBalance(apiCtx, session)
		if err != nil {
			if errors.Is(err, ErrUnauthorized) {
				log.Printf("  ❌ User %d: token expired, removing session", userID)
				sessions.Delete(userID)
			} else {
				log.Printf("  ⚠️ User %d: API error (%v), keeping session", userID, err)
			}
		} else {
			log.Printf("  ✅ User %d: session valid (+%s)", userID, session.FullPhone)
		}
	}
}
