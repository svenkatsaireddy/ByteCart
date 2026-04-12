package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"oolio-backend-challenge/internal/errs"
	"oolio-backend-challenge/internal/models"
)

const headerIdempotency = "Idempotency-Key"

// PlaceOrder is POST /order.
func (h *Handler) PlaceOrder(c *gin.Context) {
	if key := strings.TrimSpace(c.GetHeader(headerIdempotency)); key != "" && h.Idempotent != nil {
		if body, status, ok := h.Idempotent.Get(key); ok {
			c.Data(status, "application/json", body)
			return
		}
	}

	var req models.OrderReq
	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(&req); err != nil {
		msg := "invalid JSON body"
		var se *json.SyntaxError
		if errors.As(err, &se) {
			msg = fmt.Sprintf("invalid JSON at byte offset %d", se.Offset)
		}
		errs.JSON(c, http.StatusBadRequest, errs.CodeBadRequest, msg)
		return
	}

	if len(req.Items) == 0 {
		errs.JSON(c, http.StatusUnprocessableEntity, errs.CodeValidation, "items must not be empty")
		return
	}

	var subtotal float64
	var outItems []models.OrderItemOut
	var products []models.Product
	seen := make(map[string]struct{})

	for _, line := range req.Items {
		pid := strings.TrimSpace(line.ProductID)
		if pid == "" {
			errs.JSON(c, http.StatusUnprocessableEntity, errs.CodeValidation, "productId is required for each item")
			return
		}
		if line.Quantity <= 0 {
			errs.JSON(c, http.StatusUnprocessableEntity, errs.CodeValidation, "quantity must be a positive integer")
			return
		}
		p, ok := h.Products.GetByStringID(pid)
		if !ok {
			errs.JSON(c, http.StatusUnprocessableEntity, errs.CodeValidation, "unknown product id: "+pid)
			return
		}
		subtotal += p.Price * float64(line.Quantity)
		outItems = append(outItems, models.OrderItemOut{ProductID: pid, Quantity: line.Quantity})
		if _, ok := seen[p.ID]; !ok {
			seen[p.ID] = struct{}{}
			products = append(products, p)
		}
	}

	discount := 0.0
	var couponPtr *string
	if req.CouponCode != nil {
		code := strings.TrimSpace(*req.CouponCode)
		if code != "" {
			if h.Promo == nil || !h.Promo.Valid(code) {
				errs.JSON(c, http.StatusUnprocessableEntity, errs.CodeValidation, "invalid or ineligible coupon code")
				return
			}
			discount = subtotal * (h.Config.PromoDiscountPercent / 100.0)
			couponPtr = &code
		}
	}

	total := subtotal - discount
	if total < 0 {
		total = 0
	}

	order := models.Order{
		ID:         uuid.NewString(),
		Items:      outItems,
		Products:   products,
		CouponCode: couponPtr,
		Subtotal:   round2(subtotal),
		Discount:   round2(discount),
		Total:      round2(total),
	}

	raw, err := json.Marshal(order)
	if err != nil {
		errs.JSON(c, http.StatusInternalServerError, errs.CodeInternal, "failed to build response")
		return
	}

	if key := strings.TrimSpace(c.GetHeader(headerIdempotency)); key != "" && h.Idempotent != nil {
		h.Idempotent.Put(key, http.StatusOK, raw)
	}

	c.Data(http.StatusOK, "application/json", raw)
}

func round2(x float64) float64 {
	return float64(int64(x*100+0.5)) / 100
}
