package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"oolio-backend-challenge/internal/config"
	"oolio-backend-challenge/internal/idempotency"
	"oolio-backend-challenge/internal/middleware"
	"oolio-backend-challenge/internal/models"
	"oolio-backend-challenge/internal/promo"
	"oolio-backend-challenge/internal/store"
)

func testHandler(t *testing.T) *Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)
	root := findRepoRoot(t)
	productsPath := filepath.Join(root, "data", "products.json")
	ps, err := store.LoadProducts(productsPath)
	if err != nil {
		t.Fatalf("load products: %v", err)
	}
	cfg := config.Config{
		PromoDiscountPercent: 10,
	}
	f0 := "x TWOFILES1 y" // 9 chars
	f1 := "TWOFILES1"
	f2 := "nope"
	pv, err := promo.NewValidatorFromStringContents([]string{f0, f1, f2})
	if err != nil {
		t.Fatal(err)
	}
	return &Handler{
		Config:     cfg,
		Products:   ps,
		Promo:      pv,
		Idempotent: idempotency.NewStore(),
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Join(dir, "..")
	}
	t.Fatal("go.mod not found")
	return ""
}

func TestListAndGetProduct(t *testing.T) {
	h := testHandler(t)
	r := gin.New()
	r.GET("/api/product", h.ListProducts)
	r.GET("/api/product/:productId", h.GetProduct)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/product", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("list: %d", w.Code)
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/product/1", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != 200 {
		t.Fatalf("get 1: %d", w2.Code)
	}

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/api/product/999", nil)
	r.ServeHTTP(w3, req3)
	if w3.Code != 404 {
		t.Fatalf("get 999: %d", w3.Code)
	}

	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodGet, "/api/product/abc", nil)
	r.ServeHTTP(w4, req4)
	if w4.Code != 400 {
		t.Fatalf("get abc: %d", w4.Code)
	}
}

func TestPlaceOrderAuth(t *testing.T) {
	h := testHandler(t)
	auth := middleware.NewAPIAuth(map[string][]string{
		"apitest": {"create_order"},
		"readonly": {},
	})
	r := gin.New()
	r.POST("/api/order", auth.RequireScope("create_order"), h.PlaceOrder)

	body := `{"items":[{"productId":"1","quantity":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/order", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Fatalf("missing key: %d %s", w.Code, w.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/order", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("api_key", "readonly")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != 403 {
		t.Fatalf("readonly: %d", w2.Code)
	}
}

func TestPlaceOrderHappyPath(t *testing.T) {
	h := testHandler(t)
	auth := middleware.NewAPIAuth(map[string][]string{"apitest": {"create_order"}})
	r := gin.New()
	r.POST("/api/order", auth.RequireScope("create_order"), h.PlaceOrder)

	body := `{"items":[{"productId":"1","quantity":2}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/order", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_key", "apitest")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("order: %d %s", w.Code, w.Body.String())
	}
	var o models.Order
	if err := json.Unmarshal(w.Body.Bytes(), &o); err != nil {
		t.Fatal(err)
	}
	if o.Total != 13 {
		t.Fatalf("total want 13 got %v", o.Total)
	}
}

func TestPlaceOrderIgnoresUnknownFields(t *testing.T) {
	h := testHandler(t)
	auth := middleware.NewAPIAuth(map[string][]string{"apitest": {"create_order"}})
	r := gin.New()
	r.POST("/api/order", auth.RequireScope("create_order"), h.PlaceOrder)

	body := `{"items":[{"productId":"1","quantity":2,"note":"extra"}],"clientTag":"ignored"}`
	req := httptest.NewRequest(http.MethodPost, "/api/order", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_key", "apitest")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("order with unknown fields: %d %s", w.Code, w.Body.String())
	}
}

func TestPlaceOrderCoupon(t *testing.T) {
	h := testHandler(t)
	auth := middleware.NewAPIAuth(map[string][]string{"apitest": {"create_order"}})
	r := gin.New()
	r.POST("/api/order", auth.RequireScope("create_order"), h.PlaceOrder)

	body := `{"items":[{"productId":"1","quantity":1}],"couponCode":"TWOFILES1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/order", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_key", "apitest")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("order: %d %s", w.Code, w.Body.String())
	}
	var o models.Order
	_ = json.Unmarshal(w.Body.Bytes(), &o)
	if o.Discount != 0.65 || o.Total != 5.85 { // 10% of 6.5
		t.Fatalf("discount/total got %+v", o)
	}
}

func TestIdempotency(t *testing.T) {
	h := testHandler(t)
	auth := middleware.NewAPIAuth(map[string][]string{"apitest": {"create_order"}})
	r := gin.New()
	r.POST("/api/order", auth.RequireScope("create_order"), h.PlaceOrder)

	body := `{"items":[{"productId":"1","quantity":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/order", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_key", "apitest")
	req.Header.Set("Idempotency-Key", "idem-1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	first := w.Body.String()

	req2 := httptest.NewRequest(http.MethodPost, "/api/order", strings.NewReader(`{"items":[{"productId":"2","quantity":5}]}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("api_key", "apitest")
	req2.Header.Set("Idempotency-Key", "idem-1")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Body.String() != first {
		t.Fatalf("idempotency replay mismatch")
	}
}
