package identity

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"talenro.local/platform/internal/secret"
)

// LocalEmailSender is a capture-only fixture. Production composition must use
// an external provider and never this adapter.
type LocalEmailSender struct {
	mu       sync.Mutex
	accepted map[uuid.UUID]Delivery
	order    []uuid.UUID
}

var _ EmailSender = (*LocalEmailSender)(nil)

// NewLocalEmailSender creates an empty provider-idempotent capture fixture.
func NewLocalEmailSender() *LocalEmailSender {
	return &LocalEmailSender{accepted: make(map[uuid.UUID]Delivery)}
}

// Send captures one defensive copy and treats an accepted DeliveryID as success.
func (sender *LocalEmailSender) Send(ctx context.Context, delivery Delivery) error {
	if nilIdentityValue(ctx) || sender == nil || ctx.Err() != nil || delivery.DeliveryID == uuid.Nil {
		return ErrEmailDelivery
	}
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if sender.accepted == nil {
		sender.accepted = make(map[uuid.UUID]Delivery)
	}
	if _, exists := sender.accepted[delivery.DeliveryID]; exists {
		return nil
	}
	if !validDelivery(delivery) {
		return ErrEmailDelivery
	}
	owned := cloneDelivery(delivery)
	sender.accepted[owned.DeliveryID] = owned
	sender.order = append(sender.order, owned.DeliveryID)
	return nil
}

// Captured returns provider-acceptance order with no aliases to internal data.
func (sender *LocalEmailSender) Captured() []Delivery {
	if sender == nil {
		return nil
	}
	sender.mu.Lock()
	defer sender.mu.Unlock()
	result := make([]Delivery, 0, len(sender.order))
	for _, deliveryID := range sender.order {
		result = append(result, cloneDelivery(sender.accepted[deliveryID]))
	}
	return result
}

func validDelivery(delivery Delivery) bool {
	if delivery.DeliveryID == uuid.Nil || !validTemplateID(delivery.TemplateID) || !validIdentityLocale(delivery.Locale) {
		return false
	}
	canonical, err := CanonicalizeEmail(delivery.Recipient)
	if err != nil || string(canonical.Bytes()) != delivery.Recipient {
		return false
	}
	token := delivery.OneTimeToken.Copy()
	defer clear(token)
	return len(token) == 32
}

func cloneDelivery(delivery Delivery) Delivery {
	token := delivery.OneTimeToken.Copy()
	defer clear(token)
	return Delivery{
		DeliveryID: delivery.DeliveryID, TemplateID: delivery.TemplateID, Locale: delivery.Locale,
		Recipient: delivery.Recipient, OneTimeToken: secret.NewBytes(token),
	}
}

func validTemplateID(template TemplateID) bool {
	switch template {
	case VerifyEmailTemplate, ResetPasswordTemplate:
		return true
	default:
		return false
	}
}
