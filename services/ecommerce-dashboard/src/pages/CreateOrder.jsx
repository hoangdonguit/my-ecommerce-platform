import { useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { createOrder } from "../api/gateway";
import ErrorBox from "../components/ErrorBox";
import Loading from "../components/Loading";

function makeIdempotencyKey(prefix) {
  return `${prefix}-${Date.now()}`;
}

const baseForm = {
  user_id: "user-success-001",
  product_id: "prod-123",
  quantity: 1,
  currency: "VND",
  payment_method: "COD",
  shipping_address: "Thu Duc, Ho Chi Minh City",
  note: "created from ecommerce dashboard",
  idempotency_key: makeIdempotencyKey("web-success"),
};

export default function CreateOrder() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const demo = searchParams.get("demo");

  const initialForm = useMemo(() => {
    if (demo === "failed") {
      return {
        ...baseForm,
        user_id: "blocked-user-001",
        payment_method: "CARD",
        note: "payment failed demo from ecommerce dashboard",
        idempotency_key: makeIdempotencyKey("web-failed"),
      };
    }

    return {
      ...baseForm,
      idempotency_key: makeIdempotencyKey("web-success"),
    };
  }, [demo]);

  const [form, setForm] = useState(initialForm);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    setForm(initialForm);
  }, [initialForm]);

  function updateField(field, value) {
    setForm((prev) => ({
      ...prev,
      [field]: value,
    }));
  }

  function applySuccessDemo() {
    setForm({
      ...baseForm,
      idempotency_key: makeIdempotencyKey("web-success"),
    });
  }

  function applyFailedDemo() {
    setForm({
      ...baseForm,
      user_id: "blocked-user-001",
      payment_method: "CARD",
      note: "payment failed demo from ecommerce dashboard",
      idempotency_key: makeIdempotencyKey("web-failed"),
    });
  }

  async function handleSubmit(event) {
    event.preventDefault();

    setSubmitting(true);
    setError(null);

    try {
      const res = await createOrder(form);
      const orderId = res?.data?.order?.id;

      if (!orderId) {
        throw new Error("Gateway không trả về order.id");
      }

      navigate(`/orders/${orderId}`);
    } catch (err) {
      setError(err);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="page narrow">
      <div className="page-header">
        <div>
          <h2>Tạo đơn hàng</h2>
          <p>Tạo order qua Web Gateway, sau đó theo dõi Saga timeline.</p>
        </div>
      </div>

      <div className="demo-actions">
        <button className="btn secondary" onClick={applySuccessDemo} type="button">
          Fill demo success
        </button>
        <button className="btn danger-outline" onClick={applyFailedDemo} type="button">
          Fill demo failed
        </button>
      </div>

      <ErrorBox error={error} />

      <form className="card form" onSubmit={handleSubmit}>
        <label>
          User ID
          <input
            value={form.user_id}
            onChange={(e) => updateField("user_id", e.target.value)}
            required
          />
        </label>

        <div className="form-row">
          <label>
            Product ID
            <input
              value={form.product_id}
              onChange={(e) => updateField("product_id", e.target.value)}
              required
            />
          </label>

          <label>
            Quantity
            <input
              type="number"
              min="1"
              value={form.quantity}
              onChange={(e) => updateField("quantity", e.target.value)}
              required
            />
          </label>
        </div>

        <div className="form-row">
          <label>
            Currency
            <input
              value={form.currency}
              onChange={(e) => updateField("currency", e.target.value)}
              required
            />
          </label>

          <label>
            Payment Method
            <select
              value={form.payment_method}
              onChange={(e) => updateField("payment_method", e.target.value)}
              required
            >
              <option value="COD">COD</option>
              <option value="CARD">CARD</option>
              <option value="MOMO">MOMO</option>
              <option value="ZALOPAY">ZALOPAY</option>
              <option value="BANK_TRANSFER">BANK_TRANSFER</option>
            </select>
          </label>
        </div>

        <label>
          Shipping Address
          <input
            value={form.shipping_address}
            onChange={(e) => updateField("shipping_address", e.target.value)}
            required
          />
        </label>

        <label>
          Note
          <textarea
            value={form.note}
            onChange={(e) => updateField("note", e.target.value)}
            rows="3"
          />
        </label>

        <label>
          Idempotency Key
          <input
            value={form.idempotency_key}
            onChange={(e) => updateField("idempotency_key", e.target.value)}
            required
          />
          <small>
            Gửi lại cùng key sẽ không tạo order trùng.
          </small>
        </label>

        <button className="btn" disabled={submitting} type="submit">
          {submitting ? "Đang tạo..." : "Tạo đơn hàng"}
        </button>

        {submitting ? <Loading text="Đang gửi request đến Web Gateway..." /> : null}
      </form>

      <div className="note-box">
        <b>Lưu ý:</b> Product ID phải tồn tại trong Inventory DB và còn đủ hàng.
        Nếu order tạo xong nhưng inventory fail, kiểm tra lại dữ liệu bảng <code>inventories</code>.
      </div>
    </div>
  );
}