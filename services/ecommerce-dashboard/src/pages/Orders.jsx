import { useState } from "react";
import { Link } from "react-router-dom";
import { listOrders } from "../api/gateway";
import ErrorBox from "../components/ErrorBox";
import Loading from "../components/Loading";
import StatusBadge from "../components/StatusBadge";

export default function Orders() {
  const [userId, setUserId] = useState("user-success-001");
  const [orders, setOrders] = useState([]);
  const [meta, setMeta] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  async function loadOrders(event) {
    if (event) event.preventDefault();

    setLoading(true);
    setError(null);

    try {
      const res = await listOrders(userId, 1, 20);
      setOrders(res.data || []);
      setMeta(res.meta || null);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2>Danh sách đơn hàng</h2>
          <p>Tìm order theo user_id rồi mở timeline Saga.</p>
        </div>
      </div>

      <form className="toolbar" onSubmit={loadOrders}>
        <input
          value={userId}
          onChange={(e) => setUserId(e.target.value)}
          placeholder="Nhập user_id"
          required
        />
        <button className="btn" type="submit">
          Tìm kiếm
        </button>
      </form>

      <ErrorBox error={error} />

      {loading ? (
        <Loading />
      ) : (
        <div className="card">
          <div className="table-header">
            <h3>Orders</h3>
            {meta ? <span>Total: {meta.total}</span> : null}
          </div>

          {!orders.length ? (
            <div className="empty-state">
              Chưa có dữ liệu. Nhập user_id và bấm Tìm kiếm.
            </div>
          ) : (
            <div className="table-wrapper">
              <table>
                <thead>
                  <tr>
                    <th>Order ID</th>
                    <th>Status</th>
                    <th>Total</th>
                    <th>Payment</th>
                    <th>Created</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  {orders.map((order) => (
                    <tr key={order.id}>
                      <td className="mono">{order.id}</td>
                      <td>
                        <StatusBadge status={order.status} />
                      </td>
                      <td>{Number(order.total_amount || 0).toLocaleString()} {order.currency}</td>
                      <td>{order.payment_method}</td>
                      <td>{formatDate(order.created_at)}</td>
                      <td>
                        <Link className="link-btn" to={`/orders/${order.id}`}>
                          Xem Saga
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function formatDate(value) {
  if (!value) return "-";

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;

  return date.toLocaleString();
}