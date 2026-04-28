# Oolio Food Ordering App

This repository contains a mini food ordering app with a React shopping-cart frontend and a Go implementation of the food ordering API described by the Oolio OpenAPI specification.

The app focuses on:

- product listing with responsive images
- cart add, remove, increase, and decrease behavior
- INR totals, discount totals, and checkout confirmation
- explicit promo-code validation before discounts are applied
- correctness against the published contract
- robust order validation
- scalable promo-code validation for large gzip corpora
- simple extension points for future production work

Reference material:

- OpenAPI HTML: [https://orderfoodonline.deno.dev/public/openapi.html](https://orderfoodonline.deno.dev/public/openapi.html)
- OpenAPI YAML: [https://orderfoodonline.deno.dev/public/openapi.yaml](https://orderfoodonline.deno.dev/public/openapi.yaml)
- Local spec copy: [api/openapi.yaml](api/openapi.yaml)

## Implemented API

All endpoints from the checked-in OpenAPI spec are implemented:

- `GET /api/product`
- `GET /api/product/{productId}`
- `POST /api/order`
- `POST /api/coupon/validate`

The order and coupon-validation endpoints require an API key with the `create_order` scope.

## Tech Stack

- Go 1.22
- Gin for HTTP routing
- UUIDs for order IDs
- React 19 + Vite for the shopping-cart UI
- vendored Red Hat Text `.woff2` files for offline-safe typography

## Project Structure

- [cmd/server/main.go](cmd/server/main.go): application bootstrap and route registration
- [internal/config/config.go](internal/config/config.go): environment-driven configuration
- [internal/handlers](internal/handlers): HTTP handlers and request validation
- [internal/middleware/auth.go](internal/middleware/auth.go): API key auth and scope checks
- [internal/promo/promo.go](internal/promo/promo.go): streaming coupon validation
- [internal/store/products.go](internal/store/products.go): product catalog loading
- [internal/idempotency/store.go](internal/idempotency/store.go): in-memory idempotency replay store
- [frontend](frontend): React product listing, cart, coupons, and checkout confirmation

## Running Locally

### Prerequisites

- Go 1.22+
- Node.js 20+
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

### Start the web app in development

In a second terminal:

```bash
cd frontend
npm install
npm run dev
```

Vite proxies `/api` requests to `http://127.0.0.1:8080`, so keep the Go server running while using the dev UI.

### Build and serve as one app

```bash
cd frontend
npm run build
cd ..
go run ./cmd/server
```

When `frontend/dist/index.html` exists, the Go server serves the built frontend and still exposes the API under `/api`.

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

The Docker image builds the React frontend, copies the local `data/` directory into the container, and serves the web app and API from the same `:8080` process.

## Configuration

Environment variables supported by the server:

- `ADDR`: listen address, default `:8080`
- `PRODUCTS_PATH`: path to `products.json`, default `data/products.json`
- `COUPON_FILE_PATHS`: comma-separated gzip file paths
- `REQUIRE_COUPON_FILES`: default `true`; set to `false` to start even when coupon files are missing
- `API_KEYS`: comma-separated key-to-scope mapping, default `apitest:create_order,readonly:`
- `PROMO_DISCOUNT_PERCENT`: discount percent applied when a promo code is valid, default `10`
- `FRONTEND_DIST`: optional path to a built frontend, default `frontend/dist`

## Authentication

Authentication is implemented in [internal/middleware/auth.go](internal/middleware/auth.go).

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

Built-in challenge coupons:

- `HAPPYHOURS`: applies an 18% discount.
- `BUYGETONE`: makes the lowest priced item free.

Frontend behavior:

- The cart does not display promo-code suggestions.
- Typing a promo code does not change totals by itself.
- Clicking Apply calls `POST /api/coupon/validate` and only shows a discount after the backend accepts the code.
- Clicking Confirm Order with a typed but unapplied code validates the code first, then places the order only if validation succeeds.
- Applied discounts are cleared when cart quantities change, so totals cannot become stale.

Additional generated coupon codes can still be validated with the gzip corpus rules below.

Validation rules for generated coupons:

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

- [internal/promo/promo.go](internal/promo/promo.go)
- [internal/promo/promo_test.go](internal/promo/promo_test.go)

## Order Handling Notes

`POST /api/order` performs:

- JSON body parsing
- item presence validation
- product existence checks
- quantity validation
- coupon validation
- subtotal, discount, and total calculation
- idempotency replay when `Idempotency-Key` is provided

Idempotency keys are bound to a canonical request fingerprint. Reusing a key with the same request replays the original response; reusing it with a different cart or coupon returns `409 Conflict`.

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
- Product responses include the demo image object so the frontend can render responsive product artwork.
- Order responses include helpful `subtotal`, `discount`, and `total` fields in addition to the base OpenAPI response shape.

## Example Requests

List products:

```bash
curl http://127.0.0.1:8080/api/product
```

Get a product:

```bash
curl http://127.0.0.1:8080/api/product/1
```

Validate a promo code without creating an order:

```bash
curl -X POST http://127.0.0.1:8080/api/coupon/validate \
  -H "Content-Type: application/json" \
  -H "api_key: apitest" \
  -d '{"items":[{"productId":"1","quantity":1}],"couponCode":"HAPPYHOURS"}'
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
  -d '{"items":[{"productId":"1","quantity":1}],"couponCode":"HAPPYHOURS"}'
```

## Test Status

The test suite passes with:

```bash
go test ./...
cd frontend && npm run lint && npm run build
```

On Windows/OneDrive, if a running server has files open in `frontend/dist`, stop the server before rebuilding or run:

```bash
cd frontend
npm run build -- --emptyOutDir=false
```
