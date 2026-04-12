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
		byID[p.ID] = p
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
