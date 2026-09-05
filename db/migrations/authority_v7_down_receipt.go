package migrations

import "time"

func authorityV7DownConsumeTimeValid(transactionStartedAt, consumedAt time.Time) bool {
	return !transactionStartedAt.IsZero() &&
		!consumedAt.IsZero() &&
		!consumedAt.Before(transactionStartedAt)
}
