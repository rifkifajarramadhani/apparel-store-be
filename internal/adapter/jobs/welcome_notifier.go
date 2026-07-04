package jobs

import (
	"context"
	"log/slog"

	appmail "github.com/rifkifajarramadhani/golang-clean-architecture/internal/mail"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/user"
)

type WelcomeNotifier struct {
	mailer        *appmail.Mailer
	logger        *slog.Logger
	storefrontURL string
}

func NewWelcomeNotifier(mailer *appmail.Mailer, logger *slog.Logger, storefrontURL string) *WelcomeNotifier {
	return &WelcomeNotifier{mailer: mailer, logger: logger, storefrontURL: storefrontURL}
}

func (n *WelcomeNotifier) NotifyWelcome(ctx context.Context, account user.User) error {
	if _, err := n.mailer.Queue(ctx, appmail.Welcome{
		Username:      account.Username,
		Email:         account.Email,
		StorefrontURL: n.storefrontURL,
	}, appmail.QueueOptions{}); err != nil {
		n.logger.WarnContext(ctx, "queue welcome email failed", "user_id", account.ID, "error", err)
		return err
	}
	return nil
}
