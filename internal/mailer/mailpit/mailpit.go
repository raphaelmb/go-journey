package mailpit

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raphaelmb/go-journey/internal/pgstore"
	"github.com/wneessen/go-mail"
)

type store interface {
	GetTrip(ctx context.Context, tripID uuid.UUID) (pgstore.Trip, error)
}

type Mailpit struct {
	store store
}

func NewMailPit(pool *pgxpool.Pool) *Mailpit {
	return &Mailpit{pgstore.New(pool)}
}

func (mp *Mailpit) SendConfirmTripEmailToTripOwner(tripID uuid.UUID) error {
	ctx := context.Background()
	trip, err := mp.store.GetTrip(ctx, tripID)
	if err != nil {
		return fmt.Errorf("mailpit: failed to get trip for SendConfirmTripEmailToTripOwner: %w", err)
	}

	msg := mail.NewMsg()
	if err := msg.From("mailpit@journey.com"); err != nil {
		return fmt.Errorf("mailpit: failed to set From in email SendConfirmTripEmailToTripOwner: %w", err)
	}

	if err := msg.To(trip.OwnerEmail); err != nil {
		return fmt.Errorf("mailpit: failed to set To in email SendConfirmTripEmailToTripOwner: %w", err)
	}

	msg.Subject("Confirm your trip")
	msg.SetBodyString(mail.TypeTextPlain, fmt.Sprintf(`
		Hello, %s!
		
		Your trip to %s starting at %s needs to be confirmed.

		Click on the button to confirm.
	`, trip.OwnerName, trip.Destination, trip.StartsAt.Time.Format(time.DateOnly)))

	client, err := mail.NewClient(os.Getenv("MAIL_HOST"), mail.WithTLSPortPolicy(mail.NoTLS), mail.WithPort(1025))
	if err != nil {
		return fmt.Errorf("mailpit: failed to create email client SendConfirmTripEmailToTripOwner: %w", err)
	}

	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("mailpit: failed to send email SendConfirmTripEmailToTripOwner: %w", err)
	}

	return nil
}
