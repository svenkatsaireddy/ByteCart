import { startTransition, useEffect, useState } from 'react'
import './App.css'

const currencyFormatter = new Intl.NumberFormat('en-IN', {
  style: 'currency',
  currency: 'INR',
})

function formatCurrency(value) {
  return currencyFormatter.format(value || 0)
}

function roundToCents(value) {
  return Math.round(value * 100) / 100
}

function normalizeCoupon(code) {
  return code.trim().toUpperCase()
}

function getErrorMessage(error, fallback) {
  if (error instanceof Error && error.message) {
    return error.message
  }
  return fallback
}

function getApiErrorMessage(data, fallback) {
  if (typeof data?.error === 'string') {
    return data.error
  }
  if (typeof data?.error?.message === 'string') {
    return data.error.message
  }
  if (typeof data?.message === 'string') {
    return data.message
  }
  if (typeof data?.error?.code === 'string') {
    return data.error.code
  }
  return fallback
}

function CartIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path
        d="M3 4h2.4l1.3 7.03A2 2 0 0 0 8.67 13H17a2 2 0 0 0 1.94-1.53L20 7H7.1"
        fill="none"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.8"
      />
      <circle cx="9.2" cy="19.2" r="1.3" fill="currentColor" />
      <circle cx="17.2" cy="19.2" r="1.3" fill="currentColor" />
    </svg>
  )
}

function CheckIcon() {
  return (
    <svg viewBox="0 0 48 48" aria-hidden="true">
      <circle cx="24" cy="24" r="22" fill="currentColor" opacity="0.14" />
      <path
        d="M16 24.5l5.2 5.2L32.5 18.4"
        fill="none"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="3.2"
      />
    </svg>
  )
}

function EmptyCartIllustration() {
  return (
    <svg viewBox="0 0 140 110" aria-hidden="true">
      <rect x="20" y="28" width="100" height="62" rx="18" fill="#fff3ee" />
      <path
        d="M44 38h12l7 27h34l7-20H56"
        fill="none"
        stroke="#c73b0f"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="4"
      />
      <circle cx="68" cy="79" r="4.5" fill="#87635a" />
      <circle cx="96" cy="79" r="4.5" fill="#87635a" />
      <path
        d="M32 18c6 0 10 4 13 10m49-9c8 1 13 6 14 13"
        fill="none"
        stroke="#f4aa8b"
        strokeLinecap="round"
        strokeWidth="3.5"
      />
    </svg>
  )
}

function LoadingCard() {
  return (
    <article className="product-card skeleton-card" aria-hidden="true">
      <div className="product-media skeleton-block"></div>
      <div className="product-copy">
        <div className="skeleton-line short"></div>
        <div className="skeleton-line"></div>
        <div className="skeleton-line tiny"></div>
      </div>
    </article>
  )
}

function App() {
  const [products, setProducts] = useState([])
  const [cart, setCart] = useState({})
  const [couponCode, setCouponCode] = useState('')
  const [appliedCoupon, setAppliedCoupon] = useState(null)
  const [couponStatus, setCouponStatus] = useState('idle')
  const [couponMessage, setCouponMessage] = useState('')
  const [status, setStatus] = useState('loading')
  const [loadError, setLoadError] = useState('')
  const [orderError, setOrderError] = useState('')
  const [confirmation, setConfirmation] = useState(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [reloadSeed, setReloadSeed] = useState(0)

  useEffect(() => {
    let ignore = false

    async function loadProducts() {
      setStatus('loading')
      setLoadError('')

      try {
        const response = await fetch('/api/product')
        if (!response.ok) {
          throw new Error(`Unable to load products (${response.status})`)
        }
        const data = await response.json()
        if (ignore) {
          return
        }

        setProducts(Array.isArray(data) ? data : [])
        setStatus('ready')
      } catch (error) {
        if (ignore) {
          return
        }
        setLoadError(getErrorMessage(error, 'Unable to load the menu right now.'))
        setStatus('error')
      }
    }

    loadProducts()

    return () => {
      ignore = true
    }
  }, [reloadSeed])

  const cartItems = []
  let cartCount = 0
  let subtotal = 0

  for (const product of products) {
    const quantity = cart[product.id] ?? 0
    if (quantity > 0) {
      cartItems.push({
        product,
        quantity,
        lineTotal: roundToCents(product.price * quantity),
      })
      cartCount += quantity
      subtotal += product.price * quantity
    }
  }

  subtotal = roundToCents(subtotal)

  const discount = appliedCoupon ? roundToCents(appliedCoupon.discount) : 0
  const total = appliedCoupon
    ? roundToCents(appliedCoupon.total)
    : roundToCents(subtotal)

  function clearAppliedCoupon() {
    setAppliedCoupon(null)
    setCouponStatus('idle')
    setCouponMessage('')
  }

  function buildOrderPayload(coupon = '') {
    const payload = {
      items: cartItems.map(({ product, quantity }) => ({
        productId: product.id,
        quantity,
      })),
    }

    if (coupon) {
      payload.couponCode = coupon
    }

    return payload
  }

  async function validateCouponCode(code) {
    const response = await fetch('/api/coupon/validate', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        api_key: 'apitest',
      },
      body: JSON.stringify(buildOrderPayload(code)),
    })

    const data = await response.json().catch(() => null)
    if (!response.ok) {
      throw new Error(
        getApiErrorMessage(data, `Unable to apply coupon (${response.status})`),
      )
    }

    return {
      couponCode: data.couponCode,
      subtotal: data.subtotal,
      discount: data.discount,
      total: data.total,
    }
  }

  function updateQuantity(productId, nextQuantity) {
    clearAppliedCoupon()
    setOrderError('')
    setCart((currentCart) => {
      if (nextQuantity <= 0) {
        const nextCart = { ...currentCart }
        delete nextCart[productId]
        return nextCart
      }

      return {
        ...currentCart,
        [productId]: nextQuantity,
      }
    })
  }

  function addItem(productId) {
    clearAppliedCoupon()
    setOrderError('')
    setCart((currentCart) => ({
      ...currentCart,
      [productId]: (currentCart[productId] ?? 0) + 1,
    }))
  }

  function removeItem(productId) {
    updateQuantity(productId, 0)
  }

  function handleCouponChange(event) {
    setCouponCode(event.target.value)
    setOrderError('')
    if (appliedCoupon) {
      clearAppliedCoupon()
    }
  }

  async function handleApplyCoupon() {
    const normalizedCoupon = normalizeCoupon(couponCode)
    if (cartItems.length === 0) {
      setCouponStatus('error')
      setCouponMessage('Add at least one item before applying a coupon.')
      setAppliedCoupon(null)
      return
    }
    if (!normalizedCoupon) {
      setCouponStatus('error')
      setCouponMessage('Enter a discount code first.')
      setAppliedCoupon(null)
      return
    }

    setCouponStatus('applying')
    setCouponMessage('')
    setOrderError('')

    try {
      const validatedCoupon = await validateCouponCode(normalizedCoupon)
      setAppliedCoupon(validatedCoupon)
      setCouponStatus('applied')
      setCouponMessage(`${validatedCoupon.couponCode} applied successfully.`)
    } catch (error) {
      setAppliedCoupon(null)
      setCouponStatus('error')
      setCouponMessage(getErrorMessage(error, 'Unable to apply coupon.'))
    }
  }

  async function handlePlaceOrder() {
    if (cartItems.length === 0 || isSubmitting) {
      return
    }

    setIsSubmitting(true)
    setOrderError('')

    const normalizedCoupon = normalizeCoupon(couponCode)
    let couponForOrder = appliedCoupon?.couponCode ?? ''

    if (normalizedCoupon && normalizedCoupon !== appliedCoupon?.couponCode) {
      setCouponStatus('applying')
      setCouponMessage('')
      try {
        const validatedCoupon = await validateCouponCode(normalizedCoupon)
        setAppliedCoupon(validatedCoupon)
        setCouponStatus('applied')
        setCouponMessage(`${validatedCoupon.couponCode} applied successfully.`)
        couponForOrder = validatedCoupon.couponCode
      } catch (error) {
        setAppliedCoupon(null)
        setCouponStatus('error')
        setCouponMessage(getErrorMessage(error, 'Unable to apply coupon.'))
        setIsSubmitting(false)
        return
      }
    }

    const payload = buildOrderPayload(couponForOrder)

    try {
      const response = await fetch('/api/order', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          api_key: 'apitest',
          'Idempotency-Key':
            globalThis.crypto?.randomUUID?.() ?? `order-${Date.now()}`,
        },
        body: JSON.stringify(payload),
      })

      const data = await response.json().catch(() => null)
      if (!response.ok) {
        throw new Error(
          getApiErrorMessage(data, `Unable to place order (${response.status})`),
        )
      }

      startTransition(() => {
        setConfirmation(data)
        setCart({})
        setCouponCode('')
        setAppliedCoupon(null)
        setCouponStatus('idle')
        setCouponMessage('')
      })
    } catch (error) {
      setOrderError(
        getErrorMessage(error, 'Something went wrong while placing the order.'),
      )
    } finally {
      setIsSubmitting(false)
    }
  }

  const confirmationItems =
    confirmation?.items
      ?.map((item) => {
        const product = confirmation.products?.find(
          (entry) => entry.id === item.productId,
        )
        if (!product) {
          return null
        }
        return {
          product,
          quantity: item.quantity,
          lineTotal: roundToCents(product.price * item.quantity),
        }
      })
      .filter(Boolean) ?? []

  return (
    <>
      <main className="app-shell">
        <section className="catalog-panel">
          <header className="hero-panel">
            <p className="eyebrow">Dessert delivery</p>
            <div className="hero-copy">
              <div>
                <h1>Pick your favorites and confirm in one smooth flow.</h1>
                <p className="hero-text">
                  Live menu data comes from the Go backend, including pricing,
                  product images, and checkout totals.
                </p>
              </div>
              <div className="hero-note">
                <span className="hero-note-label">Discounts</span>
                <strong>Validated at checkout</strong>
                <span>Enter a code in the cart to preview savings.</span>
              </div>
            </div>
          </header>

          <section className="catalog-section" aria-labelledby="dessert-list">
            <div className="section-heading">
              <div>
                <p className="section-kicker">Menu</p>
                <h2 id="dessert-list">Desserts</h2>
              </div>
              <p className="section-meta">
                {status === 'ready'
                  ? `${products.length} products ready to order`
                  : 'Fetching menu'}
              </p>
            </div>

            {status === 'error' ? (
              <div className="status-card" role="alert">
                <h3>Couldn&apos;t load the dessert menu.</h3>
                <p>{loadError}</p>
                <button
                  type="button"
                  className="secondary-button"
                  onClick={() => setReloadSeed((seed) => seed + 1)}
                >
                  Try again
                </button>
              </div>
            ) : (
              <div className="product-grid">
                {status === 'loading'
                  ? Array.from({ length: 9 }, (_, index) => (
                      <LoadingCard key={index} />
                    ))
                  : products.map((product) => {
                      const quantity = cart[product.id] ?? 0
                      const image = product.image ?? {}

                      return (
                        <article className="product-card" key={product.id}>
                          <div
                            className={`product-media ${
                              quantity > 0 ? 'selected' : ''
                            }`}
                          >
                            <picture>
                              <source
                                media="(min-width: 1024px)"
                                srcSet={image.desktop || image.tablet}
                              />
                              <source
                                media="(min-width: 640px)"
                                srcSet={image.tablet || image.mobile}
                              />
                              <img
                                src={image.mobile || image.thumbnail}
                                alt={product.name}
                                loading="lazy"
                              />
                            </picture>

                            {quantity === 0 ? (
                              <button
                                type="button"
                                className="cart-action add-button"
                                onClick={() => addItem(product.id)}
                              >
                                <CartIcon />
                                Add to Cart
                              </button>
                            ) : (
                              <div
                                className="cart-action quantity-control"
                                aria-label={`${product.name} quantity`}
                              >
                                <button
                                  type="button"
                                  className="quantity-button"
                                  aria-label={`Decrease ${product.name}`}
                                  onClick={() =>
                                    updateQuantity(product.id, quantity - 1)
                                  }
                                >
                                  -
                                </button>
                                <span>{quantity}</span>
                                <button
                                  type="button"
                                  className="quantity-button"
                                  aria-label={`Increase ${product.name}`}
                                  onClick={() =>
                                    updateQuantity(product.id, quantity + 1)
                                  }
                                >
                                  +
                                </button>
                              </div>
                            )}
                          </div>

                          <div className="product-copy">
                            <p className="product-category">{product.category}</p>
                            <h3>{product.name}</h3>
                            <p className="product-price">
                              {formatCurrency(product.price)}
                            </p>
                          </div>
                        </article>
                      )
                    })}
              </div>
            )}
          </section>
        </section>

        <aside className="cart-panel" aria-labelledby="cart-heading">
          <div className="cart-card">
            <div className="cart-header">
              <h2 id="cart-heading">Your Cart ({cartCount})</h2>
              <p>Adjust quantities before you confirm.</p>
            </div>

            {cartItems.length === 0 ? (
              <div className="empty-cart">
                <EmptyCartIllustration />
                <p>Your added items will appear here.</p>
              </div>
            ) : (
              <>
                <ul className="cart-list">
                  {cartItems.map(({ product, quantity, lineTotal }) => (
                    <li className="cart-row" key={product.id}>
                      <div className="cart-row-copy">
                        <h3>{product.name}</h3>
                        <p>
                          <span className="quantity-text">{quantity}x</span>
                          <span>@ {formatCurrency(product.price)}</span>
                        </p>
                      </div>
                      <div className="cart-row-actions">
                        <strong>{formatCurrency(lineTotal)}</strong>
                        <button
                          type="button"
                          className="remove-button"
                          aria-label={`Remove ${product.name}`}
                          onClick={() => removeItem(product.id)}
                        >
                          <span aria-hidden="true">&times;</span>
                        </button>
                      </div>
                    </li>
                  ))}
                </ul>

                <div className="coupon-group">
                  <label className="coupon-field" htmlFor="couponCode">
                    <span>Discount code</span>
                    <div className="coupon-control">
                      <input
                        id="couponCode"
                        name="couponCode"
                        type="text"
                        placeholder="Enter discount code"
                        value={couponCode}
                        onChange={handleCouponChange}
                      />
                      <button
                        type="button"
                        className="apply-button"
                        disabled={couponStatus === 'applying'}
                        onClick={handleApplyCoupon}
                      >
                        {couponStatus === 'applying' ? 'Applying...' : 'Apply'}
                      </button>
                    </div>
                  </label>

                  {couponMessage ? (
                    <p
                      className={`coupon-hint ${
                        couponStatus === 'applied' ? 'success' : 'error'
                      }`}
                      role={couponStatus === 'error' ? 'alert' : undefined}
                    >
                      {couponMessage}
                    </p>
                  ) : null}
                </div>

                <div className="totals-panel">
                  <div className="total-row">
                    <span>Order total</span>
                    <strong>{formatCurrency(subtotal)}</strong>
                  </div>
                  <div className="total-row">
                    <span>Discount</span>
                    <strong>
                      {discount > 0
                        ? `- ${formatCurrency(discount)}`
                        : formatCurrency(0)}
                    </strong>
                  </div>
                  <div className="total-row grand-total">
                    <span>Grand total</span>
                    <strong>{formatCurrency(total)}</strong>
                  </div>
                </div>

                <div className="delivery-note">
                  <span className="delivery-dot"></span>
                  <p>
                    This order is carbon-neutral and confirmed through the live
                    backend API.
                  </p>
                </div>

                {orderError ? (
                  <p className="form-error" role="alert">
                    {orderError}
                  </p>
                ) : null}

                <button
                  type="button"
                  className="confirm-button"
                  disabled={isSubmitting || couponStatus === 'applying'}
                  onClick={handlePlaceOrder}
                >
                  {isSubmitting || couponStatus === 'applying'
                    ? 'Confirming Order...'
                    : 'Confirm Order'}
                </button>
              </>
            )}
          </div>
        </aside>
      </main>

      {confirmation ? (
        <div className="modal-backdrop" role="presentation">
          <section
            className="confirmation-modal"
            role="dialog"
            aria-modal="true"
            aria-labelledby="confirmation-title"
          >
            <div className="confirmation-icon">
              <CheckIcon />
            </div>
            <p className="section-kicker">Order confirmed</p>
            <h2 id="confirmation-title">We&apos;re getting your order ready.</h2>
            <p className="confirmation-copy">
              Your order ID is <strong>{confirmation.id}</strong>.
            </p>

            <div className="confirmation-summary">
              <ul className="confirmation-list">
                {confirmationItems.map(({ product, quantity, lineTotal }) => (
                  <li key={product.id} className="confirmation-row">
                    <img
                      src={product.image?.thumbnail || product.image?.mobile}
                      alt=""
                    />
                    <div className="confirmation-row-copy">
                      <strong>{product.name}</strong>
                      <p>
                        <span>{quantity}x</span>
                        <span>@ {formatCurrency(product.price)}</span>
                      </p>
                    </div>
                    <strong>{formatCurrency(lineTotal)}</strong>
                  </li>
                ))}
              </ul>

              <div className="confirmation-totals">
                <div className="total-row">
                  <span>Subtotal</span>
                  <strong>{formatCurrency(confirmation.subtotal)}</strong>
                </div>
                <div className="total-row">
                  <span>Discount</span>
                  <strong>{formatCurrency(confirmation.discount)}</strong>
                </div>
                <div className="total-row grand-total">
                  <span>Order Total</span>
                  <strong>{formatCurrency(confirmation.total)}</strong>
                </div>
              </div>
            </div>

            <button
              type="button"
              className="confirm-button"
              onClick={() => {
                setConfirmation(null)
                setOrderError('')
              }}
            >
              Start New Order
            </button>
          </section>
        </div>
      ) : null}
    </>
  )
}

export default App
