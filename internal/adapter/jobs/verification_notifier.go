package jobs

import (
	"context"
	"log/slog"
	"net/url"

	appmail "github.com/rifkifajarramadhani/golang-clean-architecture/internal/mail"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/user"
)

type VerificationNotifier struct {
	mailer    *appmail.Mailer
	logger    *slog.Logger
	publicURL string
}

func NewVerificationNotifier(mailer *appmail.Mailer, logger *slog.Logger, publicURL string) *VerificationNotifier {
	return &VerificationNotifier{mailer: mailer, logger: logger, publicURL: publicURL}
}

func (n *VerificationNotifier) NotifyVerification(ctx context.Context, account user.User, token string) error {
	if _, err := n.mailer.Queue(ctx, appmail.EmailVerification{
		Username:        account.Username,
		Email:           account.Email,
		VerificationURL: n.publicURL + "/api/auth/verify-email?token=" + url.QueryEscape(token),
	}, appmail.QueueOptions{}); err != nil {
		n.logger.WarnContext(ctx, "queue email verification failed", "user_id", account.ID, "error", err)
		return err
	}

	return nil
}
