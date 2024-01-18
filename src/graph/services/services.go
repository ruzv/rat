package services

import (
	"context"
)

// Service describes the common methods all rat services must implement.
type Service interface {
	// Starts the service and blocks until it is stopped or an unrecoverable
	// error occurs.
	Run() error
	// Stops the service. Retruns error on failure to stop service. Either
	// with error or without must stop the service.
	Stop(context.Context) error
}
