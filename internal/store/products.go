package store

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"oolio-backend-challenge/internal/models"
)

// ProductStore is an in-memory catalog keyed by product id string.
type ProductStore struct {
	byID map[string]models.Product
	list []models.Product
}

var productImages = map[string]models.ProductImage{
	"1": {
		Thumbnail: "https://orderfoodonline.deno.dev/public/images/image-waffle-thumbnail.jpg",
		Mobile:    "https://orderfoodonline.deno.dev/public/images/image-waffle-mobile.jpg",
		Tablet:    "https://orderfoodonline.deno.dev/public/images/image-waffle-tablet.jpg",
		Desktop:   "https://orderfoodonline.deno.dev/public/images/image-waffle-desktop.jpg",
	},
	"2": {
		Thumbnail: "https://orderfoodonline.deno.dev/public/images/image-creme-brulee-thumbnail.jpg",
		Mobile:    "https://orderfoodonline.deno.dev/public/images/image-creme-brulee-mobile.jpg",
		Tablet:    "https://orderfoodonline.deno.dev/public/images/image-creme-brulee-tablet.jpg",
		Desktop:   "https://orderfoodonline.deno.dev/public/images/image-creme-brulee-desktop.jpg",
	},
	"3": {
		Thumbnail: "https://orderfoodonline.deno.dev/public/images/image-macaron-thumbnail.jpg",
		Mobile:    "https://orderfoodonline.deno.dev/public/images/image-macaron-mobile.jpg",
		Tablet:    "https://orderfoodonline.deno.dev/public/images/image-macaron-tablet.jpg",
		Desktop:   "https://orderfoodonline.deno.dev/public/images/image-macaron-desktop.jpg",
	},
	"4": {
		Thumbnail: "https://orderfoodonline.deno.dev/public/images/image-tiramisu-thumbnail.jpg",
		Mobile:    "https://orderfoodonline.deno.dev/public/images/image-tiramisu-mobile.jpg",
		Tablet:    "https://orderfoodonline.deno.dev/public/images/image-tiramisu-tablet.jpg",
		Desktop:   "https://orderfoodonline.deno.dev/public/images/image-tiramisu-desktop.jpg",
	},
	"5": {
		Thumbnail: "https://orderfoodonline.deno.dev/public/images/image-baklava-thumbnail.jpg",
		Mobile:    "https://orderfoodonline.deno.dev/public/images/image-baklava-mobile.jpg",
		Tablet:    "https://orderfoodonline.deno.dev/public/images/image-baklava-tablet.jpg",
		Desktop:   "https://orderfoodonline.deno.dev/public/images/image-baklava-desktop.jpg",
	},
	"6": {
		Thumbnail: "https://orderfoodonline.deno.dev/public/images/image-meringue-thumbnail.jpg",
		Mobile:    "https://orderfoodonline.deno.dev/public/images/image-meringue-mobile.jpg",
		Tablet:    "https://orderfoodonline.deno.dev/public/images/image-meringue-tablet.jpg",
		Desktop:   "https://orderfoodonline.deno.dev/public/images/image-meringue-desktop.jpg",
	},
	"7": {
		Thumbnail: "https://orderfoodonline.deno.dev/public/images/image-cake-thumbnail.jpg",
		Mobile:    "https://orderfoodonline.deno.dev/public/images/image-cake-mobile.jpg",
		Tablet:    "https://orderfoodonline.deno.dev/public/images/image-cake-tablet.jpg",
		Desktop:   "https://orderfoodonline.deno.dev/public/images/image-cake-desktop.jpg",
	},
	"8": {
		Thumbnail: "https://orderfoodonline.deno.dev/public/images/image-brownie-thumbnail.jpg",
		Mobile:    "https://orderfoodonline.deno.dev/public/images/image-brownie-mobile.jpg",
		Tablet:    "https://orderfoodonline.deno.dev/public/images/image-brownie-tablet.jpg",
		Desktop:   "https://orderfoodonline.deno.dev/public/images/image-brownie-desktop.jpg",
	},
	"9": {
		Thumbnail: "https://orderfoodonline.deno.dev/public/images/image-panna-cotta-thumbnail.jpg",
		Mobile:    "https://orderfoodonline.deno.dev/public/images/image-panna-cotta-mobile.jpg",
		Tablet:    "https://orderfoodonline.deno.dev/public/images/image-panna-cotta-tablet.jpg",
		Desktop:   "https://orderfoodonline.deno.dev/public/images/image-panna-cotta-desktop.jpg",
	},
}

// LoadProducts reads JSON array of products from path.
func LoadProducts(path string) (*ProductStore, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read products: %w", err)
	}
	var list []models.Product
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, fmt.Errorf("parse products: %w", err)
	}
	byID := make(map[string]models.Product, len(list))
	for _, p := range list {
		if img, ok := productImages[p.ID]; ok {
			p.Image = img
		}
		byID[p.ID] = p
	}
	for i := range list {
		if img, ok := productImages[list[i].ID]; ok {
			list[i].Image = img
		}
	}
	return &ProductStore{byID: byID, list: list}, nil
}

// List returns all products.
func (s *ProductStore) List() []models.Product {
	out := make([]models.Product, len(s.list))
	copy(out, s.list)
	return out
}

// GetByStringID returns a product by id string (order lines use string ids).
func (s *ProductStore) GetByStringID(id string) (models.Product, bool) {
	p, ok := s.byID[id]
	return p, ok
}

// GetByPathID parses int64 path parameter and looks up product.
func (s *ProductStore) GetByPathID(path string) (models.Product, error) {
	n, err := strconv.ParseInt(path, 10, 64)
	if err != nil || n < 0 {
		return models.Product{}, errPathID
	}
	id := strconv.FormatInt(n, 10)
	p, ok := s.byID[id]
	if !ok {
		return models.Product{}, errNotFound
	}
	return p, nil
}

type pathErr int

func (e pathErr) Error() string {
	switch e {
	case errPathID:
		return "invalid id"
	default:
		return "not found"
	}
}

var errPathID = pathErr(1)
var errNotFound = pathErr(2)

// IsInvalidID reports whether GetByPathID failed due to malformed id.
func IsInvalidID(err error) bool {
	return err == errPathID
}

// IsNotFound reports whether GetByPathID failed due to missing product.
func IsNotFound(err error) bool {
	return err == errNotFound
}
