package services

import (
	"context"
	"log/slog"
)

// ConsoleEmailService logs invitation emails to stdout instead of sending them.
type ConsoleEmailService struct{}

func NewConsoleEmailService() *ConsoleEmailService {
	return &ConsoleEmailService{}
}

func (s *ConsoleEmailService) SendInvitation(_ context.Context, to, inviterName, orgName, inviteURL string) error {
	slog.Info("invitation email",
		"to", to,
		"inviter", inviterName,
		"organization", orgName,
		"invite_url", inviteURL,
	)
	return nil
}
