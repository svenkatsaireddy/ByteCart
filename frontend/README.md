# Food Ordering Frontend

React + Vite shopping-cart UI for the Oolio food ordering challenge.

## Features

- Product listing with responsive API-provided images
- Add, remove, increase, and decrease cart items
- INR subtotal, discount, and grand total display
- Explicit Apply button for promo-code validation
- Confirm-time validation when a user typed a promo code but did not click Apply
- Order submission to the Go API and confirmation modal
- Responsive layout and keyboard-visible focus states
- Vendored Red Hat Text `.woff2` files for offline-safe typography

## Promo Codes

The cart does not show promo-code suggestions. Users can enter a code and click Apply; the frontend calls `POST /api/coupon/validate` and only updates the discount total after the backend accepts the code.

Supported challenge coupons:

- `HAPPYHOURS`: 18% off the order total
- `BUYGETONE`: lowest priced item free

If a user changes cart quantities after applying a code, the applied discount is cleared and must be validated again.

## Development

Start the Go API from the repository root first:

```bash
go run ./cmd/server
```

Then run the frontend:

```bash
cd frontend
npm install
npm run dev
```

The Vite dev server proxies `/api` requests to `http://127.0.0.1:8080`.

## Build

```bash
npm run lint
npm run build
```

If a local Go server is serving `frontend/dist` and Windows reports a locked `dist/assets` directory, stop the server before rebuilding or use:

```bash
npm run build -- --emptyOutDir=false
```
