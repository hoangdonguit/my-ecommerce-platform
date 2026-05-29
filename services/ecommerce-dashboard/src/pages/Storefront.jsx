import { useState, useEffect } from "react";
import { createOrder, getInventories } from "../api/gateway";
import ErrorBox from "../components/ErrorBox";
import Loading from "../components/Loading";
import { useSaga } from "../context/SagaContext";

const DEFAULT_PRODUCTS = [
  { id: "prod-123", name: "Laptop ASUS TUF Gaming F15", price: 24000000, icon: "💻", stock: 0 },
  { id: "prod-456", name: "Bàn phím cơ Keychron", price: 2500000, icon: "⌨️", stock: 0 },
  { id: "prod-789", name: "Chuột Gaming Logitech", price: 1200000, icon: "🖱️", stock: 0 },
];

const DEMO_CHOICES = [
  {
    value: "none",
    title: "Bình thường",
    description: "Tạo đơn thành công và hoàn tất toàn bộ Saga.",
    tone: "success",
  },
  {
    value: "inventory",
    title: "Ép lỗi Kho",
    description: "Dùng product id không tồn tại để kiểm tra inventory.failed.",
    tone: "danger",
  },
  {
    value: "payment",
    title: "Ép lỗi Thanh toán",
    description: "Dùng blocked-user-001 và CARD để kiểm tra rollback.",
    tone: "danger",
  },
];

export default function Storefront() {
  const [loading, setLoading] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [message, setMessage] = useState(null);
  const [error, setError] = useState(null);
  const [checkoutItem, setCheckoutItem] = useState(null);
  const [demoType, setDemoType] = useState("none");
  const [formData, setFormData] = useState({
    shipping_address: "",
    payment_method: "COD",
    note: "",
    quantity: 1,
  });

  const { products, setProducts } = useSaga();

  const [currentUserId] = useState(() => {
    const existing = sessionStorage.getItem("storefront_user_id");
    const id = existing || `customer-${Math.floor(Math.random() * 9000 + 1000)}`;
    sessionStorage.setItem("storefront_user_id", id);
    return id;
  });

  useEffect(() => {
    if (!products || products.length === 0) {
      setProducts(DEFAULT_PRODUCTS);
    }
  }, []);

  async function fetchRealInventory(options = {}) {
    const silent = Boolean(options.silent);

    if (!silent) {
      setSyncing(true);
      setError(null);
    }

    try {
      const res = await getInventories();

      if (res?.data) {
        setProducts((prevProducts) => {
          const currentProducts = prevProducts?.length ? prevProducts : DEFAULT_PRODUCTS;

          return currentProducts.map((product) => {
            const realData = res.data.find((item) => {
              return (item.product_id || item.productId) === product.id;
            });

            return realData
              ? {
                  ...product,
                  stock: realData.available_quantity ?? realData.availableQuantity ?? 0,
                }
              : product;
          });
        });
      }
    } catch (err) {
      if (!silent) {
        setError(err);
      } else {
        console.error("Lỗi đồng bộ kho hàng:", err);
      }
    } finally {
      if (!silent) {
        setSyncing(false);
      }
    }
  }

  useEffect(() => {
    fetchRealInventory({ silent: true });
    const interval = window.setInterval(() => fetchRealInventory({ silent: true }), 10000);
    return () => window.clearInterval(interval);
  }, []);

  function handleOpenCheckout(product) {
    setCheckoutItem(product);
    setFormData({
      shipping_address: "",
      payment_method: "COD",
      note: "",
      quantity: 1,
    });
    setDemoType("none");
    setMessage(null);
    setError(null);
  }

  function closeCheckout() {
    if (!loading) {
      setCheckoutItem(null);
    }
  }

  async function submitOrder(event) {
    event.preventDefault();

    if (!checkoutItem) return;

    setLoading(true);
    setError(null);
    setMessage(null);

    try {
      const quantity = Math.max(1, Number(formData.quantity) || 1);

      let finalUserId = currentUserId;
      let finalQuantity = quantity;
      let finalPaymentMethod = formData.payment_method;
      let finalProductId = checkoutItem.id;

      if (demoType === "inventory") {
        finalProductId = "sanphamsieucapvippro-789";
        finalQuantity = 1;
      }

      if (demoType === "payment") {
        finalUserId = "blocked-user-001";
        finalPaymentMethod = "CARD";
      }

      const payload = {
        user_id: finalUserId,
        product_id: finalProductId,
        quantity: finalQuantity,
        currency: "VND",
        payment_method: finalPaymentMethod,
        shipping_address: formData.shipping_address,
        note: demoType !== "none" ? `Demo Error: ${demoType}` : formData.note,
      };

      const res = await createOrder(payload);
      const orderId = res?.data?.order?.id;

      if (demoType === "none") {
        setMessage({
          type: "success",
          text: `Mua thành công ${quantity} ${checkoutItem.name}. Mã đơn: ${orderId}`,
        });

        setProducts((prev) =>
          prev.map((product) =>
            product.id === checkoutItem.id
              ? { ...product, stock: Math.max(0, Number(product.stock || 0) - quantity) }
              : product
          )
        );
      } else {
        setMessage({
          type: "warning",
          text: `Đã gửi kịch bản ${demoType.toUpperCase()}. Mã đơn: ${orderId}. Kiểm tra tab đơn lỗi hoặc Trace Saga.`,
        });

        window.setTimeout(() => fetchRealInventory({ silent: true }), 3000);
      }

      setCheckoutItem(null);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="page store-page">
      <div className="store-hero">
        <div>
          <p className="product-meta">Khu vực khách hàng</p>
          <h2>🛒 Cửa hàng Demo Saga</h2>
          <p>
            Tạo đơn hàng thật qua Web Gateway để kiểm thử luồng Order, Inventory,
            Payment, Read Model và Notification.
          </p>
        </div>

        <div className="store-hero-actions">
          <div className="store-session">
            Phiên: <strong>{currentUserId}</strong>
          </div>
          <button className="btn secondary" onClick={() => fetchRealInventory()} disabled={syncing}>
            {syncing ? "Đang đồng bộ..." : "↻ Đồng bộ kho"}
          </button>
        </div>
      </div>

      <ErrorBox error={error} />

      {message ? <div className={`message-box ${message.type}`}>{message.text}</div> : null}

      <div className="store-grid">
        {(products || []).map((product) => {
          const stock = Number(product.stock || 0);
          const isAvailable = stock > 0;

          return (
            <article key={product.id} className="product-card">
              <div className="product-card-header">
                <div className="product-icon">{product.icon}</div>
                <div>
                  <div className="product-meta">{product.id}</div>
                  <h3 className="product-title">{product.name}</h3>
                </div>
              </div>

              <div className="product-body">
                <span className={`stock-pill ${isAvailable ? "" : "low"}`}>
                  {isAvailable ? `Còn lại: ${stock.toLocaleString()}` : "Hết hàng"}
                </span>
                <div className="price">{product.price.toLocaleString()} VND</div>
              </div>

              <button
                className="btn product-buy-btn"
                onClick={() => handleOpenCheckout(product)}
                disabled={!isAvailable}
              >
                {isAvailable ? "Mua ngay" : "Hết hàng"}
              </button>
            </article>
          );
        })}
      </div>

      {checkoutItem ? (
        <div className="checkout-backdrop" role="dialog" aria-modal="true">
          <div className="checkout-modal">
            <div className="checkout-head">
              <div>
                <p className="product-meta">Xác nhận mua hàng</p>
                <h3>{checkoutItem.name}</h3>
              </div>
              <button type="button" className="btn secondary" onClick={closeCheckout} disabled={loading}>
                Đóng
              </button>
            </div>

            <div className="checkout-summary">
              <div className="product-icon">{checkoutItem.icon}</div>
              <div>
                <strong>{checkoutItem.price.toLocaleString()} VND</strong>
                <p>Tồn kho hiện tại: {Number(checkoutItem.stock || 0).toLocaleString()}</p>
              </div>
            </div>

            <form className="checkout-form" onSubmit={submitOrder}>
              <label>
                Số lượng
                <input
                  type="number"
                  min="1"
                  value={formData.quantity}
                  onChange={(event) => {
                    setFormData({ ...formData, quantity: Number(event.target.value) || 1 });
                  }}
                  disabled={demoType !== "none" || loading}
                  required
                />
              </label>

              <label>
                Địa chỉ giao hàng
                <input
                  type="text"
                  value={formData.shipping_address}
                  onChange={(event) => {
                    setFormData({ ...formData, shipping_address: event.target.value });
                  }}
                  placeholder="Ví dụ: UIT, Thủ Đức, TP.HCM"
                  disabled={loading}
                  required
                />
              </label>

              <label>
                Thanh toán
                <select
                  value={formData.payment_method}
                  onChange={(event) => {
                    setFormData({ ...formData, payment_method: event.target.value });
                  }}
                  disabled={demoType === "payment" || loading}
                  required
                >
                  <option value="COD">Thanh toán khi nhận hàng (COD)</option>
                  <option value="CARD">Thẻ tín dụng / ghi nợ (CARD)</option>
                </select>
              </label>

              <label>
                Ghi chú
                <textarea
                  rows="2"
                  value={formData.note}
                  onChange={(event) => {
                    setFormData({ ...formData, note: event.target.value });
                  }}
                  placeholder="Ghi chú đơn hàng..."
                  disabled={loading}
                />
              </label>

              <div>
                <strong>Kịch bản Saga</strong>
                <div className="demo-options">
                  {DEMO_CHOICES.map((choice) => (
                    <label
                      key={choice.value}
                      className={`demo-choice ${demoType === choice.value ? "active" : ""} ${choice.tone}`}
                    >
                      <input
                        type="radio"
                        name="demo"
                        checked={demoType === choice.value}
                        onChange={() => setDemoType(choice.value)}
                        disabled={loading}
                      />
                      <span>
                        <b>{choice.title}</b>
                        <small>{choice.description}</small>
                      </span>
                    </label>
                  ))}
                </div>
              </div>

              {loading ? <Loading text="Đang gửi request tạo đơn..." /> : null}

              <div className="checkout-actions">
                <button type="submit" className="btn" disabled={loading}>
                  {loading ? "Đang xử lý..." : "Xác nhận mua hàng"}
                </button>
                <button type="button" className="btn secondary" onClick={closeCheckout} disabled={loading}>
                  Hủy
                </button>
              </div>
            </form>
          </div>
        </div>
      ) : null}
    </div>
  );
}
