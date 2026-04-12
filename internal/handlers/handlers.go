package handlers

import (
	"oolio-backend-challenge/internal/config"
	"oolio-backend-challenge/internal/idempotency"
	"oolio-backend-challenge/internal/promo"
	"oolio-backend-challenge/internal/store"
)

// Handler bundles dependencies for HTTP handlers.
type Handler struct {
	Config     config.Config
	Products   *store.ProductStore
	Promo      *promo.Validator
	Idempotent *idempotency.Store
}
