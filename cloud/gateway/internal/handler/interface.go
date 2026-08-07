// Package handler dispatches parsed device messages to NATS and the database.
package handler

import (
	"context"

	"eregen.dev/gateway/internal/model"
)

// MessageHandler processes a specific device message type.
type MessageHandler interface {
	// Handles returns the message type this handler processes.
	Handles() model.UpstreamMessageType
	// Process handles a validated device message.
	Process(ctx context.Context, msg *model.DeviceMessage) error
}
