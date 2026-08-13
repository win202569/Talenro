package controlapi

import (
	"encoding/base64"
	"io"
)

const traceBytes = 16

func (h *Handler) traceID() string {
	if h == nil || unavailableApplication(h.random) {
		return "trace-unavailable"
	}
	raw := make([]byte, traceBytes)
	defer clear(raw)
	if _, err := io.ReadFull(h.random, raw); err != nil {
		return "trace-unavailable"
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}
