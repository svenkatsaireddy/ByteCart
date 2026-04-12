package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"oolio-backend-challenge/internal/errs"
	"oolio-backend-challenge/internal/store"
)

// ListProducts is GET /product.
func (h *Handler) ListProducts(c *gin.Context) {
	c.JSON(http.StatusOK, h.Products.List())
}

// GetProduct is GET /product/:productId.
func (h *Handler) GetProduct(c *gin.Context) {
	pid := c.Param("productId")
	p, err := h.Products.GetByPathID(pid)
	if store.IsInvalidID(err) {
		errs.JSON(c, http.StatusBadRequest, errs.CodeBadRequest, "invalid product id")
		return
	}
	if store.IsNotFound(err) {
		errs.JSON(c, http.StatusNotFound, errs.CodeNotFound, "product not found")
		return
	}
	c.JSON(http.StatusOK, p)
}
