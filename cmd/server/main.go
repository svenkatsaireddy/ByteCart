package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"oolio-backend-challenge/internal/config"
	"oolio-backend-challenge/internal/handlers"
	"oolio-backend-challenge/internal/idempotency"
	"oolio-backend-challenge/internal/middleware"
	"oolio-backend-challenge/internal/promo"
	"oolio-backend-challenge/internal/store"
)

func main() {
	cfg := config.Load()
	gin.SetMode(gin.ReleaseMode)
	log.Printf("products file: %s", cfg.ProductsPath)
	log.Printf("coupon files: %s", strings.Join(cfg.CouponPaths, ", "))

	ps, err := store.LoadProducts(cfg.ProductsPath)
	if err != nil {
		log.Fatalf("load products: %v", err)
	}

	pv, err := promo.NewValidatorFromGzipFiles(cfg.CouponPaths)
	if err != nil {
		if cfg.RequireCouponFiles {
			log.Fatalf("load promo corpora: %v", err)
		}
		log.Printf("warning: promo validator disabled (coupons will fail): %v", err)
		pv = nil
	} else {
		log.Printf("promo: streaming validation enabled (fast startup; first use of each new code scans the gzip files)")
	}

	auth := middleware.NewAPIAuth(cfg.APIKeys)
	h := &handlers.Handler{
		Config:     cfg,
		Products:   ps,
		Promo:      pv,
		Idempotent: idempotency.NewStore(),
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	api.GET("/product", h.ListProducts)
	api.GET("/product/:productId", h.GetProduct)
	api.POST("/order", auth.RequireScope("create_order"), h.PlaceOrder)

	log.Printf("listening on %s", cfg.Addr)
	if err := r.Run(cfg.Addr); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}
