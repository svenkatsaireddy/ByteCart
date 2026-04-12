# Oolio Backend Challenge

This repository contains a Go implementation of the food ordering API described by the Oolio OpenAPI specification.

The server exposes the API under the `/api` prefix and focuses on:

- correctness against the published contract
- robust order validation
- scalable promo-code validation for large gzip corpora
- simple extension points for future production work

Reference material:

- OpenAPI HTML: [https://orderfoodonline.deno.dev/public/openapi.html](https://orderfoodonline.deno.dev/public/openapi.html)
- OpenAPI YAML: [https://orderfoodonline.deno.dev/public/openapi.yaml](https://orderfoodonline.deno.dev/public/openapi.yaml)
- Local spec copy: [api/openapi.yaml](/C:/Users/dines/OneDrive/Desktop/coding/backendchallenge/oolio-backend-challenge#/api/openapi.yaml)

## Implemented API

All endpoints from the checked-in OpenAPI spec are implemented:

- `GET /api/product`
- `GET /api/product/{productId}`
- `POST /api/order`

The order endpoint requires an API key with the `create_order` scope.

## Tech Stack

- Go 1.22
- Gin for HTTP routing
- UUIDs for order IDs

## Project Structure

- [cmd/server/main.go](/C:/Users/dines/OneDrive/Desktop/coding/backendchallenge/oolio-backend-challenge#/cmd/server/main.go): application bootstrap and route registration
- [internal/config/config.go](/C:/Users/dines/OneDrive/Desktop/coding/backendchallenge/oolio-backend-challenge#/internal/config/config.go): environment-driven configuration
- [internal/handlers](/C:/Users/dines/OneDrive/Desktop/coding/backendchallenge/oolio-backend-challenge#/internal/handlers): HTTP handlers and request validation
- [internal/middleware/auth.go](/C:/Users/dines/OneDrive/Desktop/coding/backendchallenge/oolio-backend-challenge#/internal/middleware/auth.go): API key auth and scope checks
- [internal/promo/promo.go](/C:/Users/dines/OneDrive/Desktop/coding/backendchallenge/oolio-backend-challenge#/internal/promo/promo.go): streaming coupon validation
- [internal/store/products.go](/C:/Users/dines/OneDrive/Desktop/coding/backendchallenge/oolio-backend-challenge#/internal/store/products.go): product catalog loading
- [internal/idempotency/store.go](/C:/Users/dines/OneDrive/Desktop/coding/backendchallenge/oolio-backend-challenge#/internal/idempotency/store.go): in-memory idempotency replay store

## Running Locally

### Prerequisites

- Go 1.22+
- The product catalog at `data/products.json`
- The three coupon gzip files in `data/coupons/`

Coupon files:

- [couponbase1.gz](https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz)
- [couponbase2.gz](https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz)
- [couponbase3.gz](https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz)

### Start the server

```bash
go run ./cmd/server
```

By default the service listens on `:8080`.

### Run tests

```bash
go test ./...
```

### Smoke test with curl

```bash
curl http://127.0.0.1:8080/api/product
curl http://127.0.0.1:8080/api/product/1
curl -X POST http://127.0.0.1:8080/api/order \
  -H "Content-Type: application/json" \
  -H "api_key: apitest" \
  --data-binary @examples/curl/order.json
```

On Windows PowerShell, use:

```powershell
.\scripts\smoke-curl.ps1
```

## Running With Docker

Build the image:

```bash
docker build -t oolio-backend-challenge .
```

Run the container:

```bash
docker run --rm -p 8080:8080 \
  -e ADDR=:8080 \
  -e API_KEYS=apitest:create_order,readonly: \
  -e PRODUCTS_PATH=/app/data/products.json \
  -e COUPON_FILE_PATHS=/app/data/coupons/couponbase1.gz,/app/data/coupons/couponbase2.gz,/app/data/coupons/couponbase3.gz \
  oolio-backend-challenge
```

If you want to start the container without coupon files being enforced, set:

```bash
-e REQUIRE_COUPON_FILES=false
```

The Docker image copies the local `data/` directory into the container, so the checked-in product catalog and coupon files are available at runtime.

## Configuration

Environment variables supported by the server:

- `ADDR`: listen address, default `:8080`
- `PRODUCTS_PATH`: path to `products.json`, default `data/products.json`
- `COUPON_FILE_PATHS`: comma-separated gzip file paths
- `REQUIRE_COUPON_FILES`: default `true`; set to `false` to start even when coupon files are missing
- `API_KEYS`: comma-separated key-to-scope mapping, default `apitest:create_order,readonly:`
- `PROMO_DISCOUNT_PERCENT`: discount percent applied when a promo code is valid, default `10`

## Authentication

Authentication is implemented in [internal/middleware/auth.go](/C:/Users/dines/OneDrive/Desktop/coding/backendchallenge/oolio-backend-challenge#/internal/middleware/auth.go).

The server expects an `api_key` header.

Default configured keys:

- `apitest`: allowed to create orders
- `readonly`: authenticated but cannot create orders

Behavior:

- missing key: `401`
- invalid key: `401`
- valid key without required scope: `403`

## Promo Code Validation

Coupon validation is the most performance-sensitive part of the challenge, so the design intentionally avoids building a huge in-memory index at startup.

Validation rules:

1. Promo code length must be between 8 and 10 characters.
2. Promo code must be alphanumeric.
3. Promo code must appear as a substring in at least 2 of the 3 gzip corpora.

Implementation details:

- The validator streams each gzip file instead of decompressing the full corpus into memory.
- It uses chunked reads with overlap handling so matches across chunk boundaries are still found.
- Results are cached in memory per code, which makes repeated validation much faster.
- Startup stays fast even with large corpora because no global coupon index is precomputed.

Tradeoff:

- The first request for a new valid or invalid code can be slower because it scans the files.
- Repeated requests for the same code are fast because of the in-memory cache.

Relevant code:

- [internal/promo/promo.go](/C:/Users/dines/OneDrive/Desktop/coding/backendchallenge/oolio-backend-challenge#/internal/promo/promo.go)
- [internal/promo/promo_test.go](/C:/Users/dines/OneDrive/Desktop/coding/backendchallenge/oolio-backend-challenge#/internal/promo/promo_test.go)

## Order Handling Notes

`POST /api/order` performs:

- JSON body parsing
- item presence validation
- product existence checks
- quantity validation
- coupon validation
- subtotal, discount, and total calculation
- idempotency replay when `Idempotency-Key` is provided

The service returns extra helpful fields in the order response:

- `couponCode`
- `subtotal`
- `discount`
- `total`

These fields are useful for clients, even though the published spec documents a smaller response shape.

## Scale and Extensibility

This solution is intentionally small, but it is structured so it can evolve into a production service.

### What already scales reasonably well

- Product reads are in-memory after startup.
- Promo validation is designed for large coupon corpora.
- Request handling is stateless except for in-memory caches.

### Current limits

- Idempotency storage is in-memory only, so it does not survive process restarts and is not shared across replicas.
- Orders are not persisted.
- Product data is loaded from a local JSON file instead of a database or service.
- Coupon validation cache is local to one process.

### Natural next steps for production

1. Move idempotency storage to Redis or a database.
2. Persist orders in a durable datastore.
3. Move products to a repository abstraction backed by a DB or another service.
4. Introduce metrics, tracing, structured logging, and health endpoints.
5. Consider a coupon index or background preprocessing if coupon throughput becomes a bottleneck.

## Conformance Notes

- All paths in the checked-in OpenAPI spec are implemented.
- The request parser now tolerates unknown JSON fields to stay closer to OpenAPI 3.1 defaults.
- One known gap versus the live demo API is that product responses do not yet include the demo's `image` object.

## Example Requests

List products:

```bash
curl http://127.0.0.1:8080/api/product
```

Get a product:

```bash
curl http://127.0.0.1:8080/api/product/1
```

Create an order:

```bash
curl -X POST http://127.0.0.1:8080/api/order \
  -H "Content-Type: application/json" \
  -H "api_key: apitest" \
  -d '{"items":[{"productId":"1","quantity":2}]}'
```

Create an order with promo code:

```bash
curl -X POST http://127.0.0.1:8080/api/order \
  -H "Content-Type: application/json" \
  -H "api_key: apitest" \
  -d '{"items":[{"productId":"1","quantity":1}],"couponCode":"HAPPYHRS"}'
```

## Test Status

The test suite passes with:

```bash
go test ./...
```
