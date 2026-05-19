import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { listReadModelOrders } from "../api/gateway";
import ErrorBox from "../components/ErrorBox";
import Loading from "../components/Loading";
import StatusBadge from "../components/StatusBadge";

export default function ReadModelOrders() {
  const [searchQuery, setSearchQuery] = useState(() => sessionStorage.getItem("read_model_order_keyword") || "");
  const [orders, setOrders] = useState([]);
  const [meta, setMeta] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  function loadOrders() {
    setLoading(true);
    setError(null);

    listReadModelOrders(1, 200)
      .then((res) => {
        setOrders(res.data || []);
        setMeta(res.meta || null);
      })
      .catch((err) => setError(err))
      .finally(() => setLoading(false));
  }

  useEffect(() => {
    loadOrders();
  }, []);

  useEffect(() => {
    sessionStorage.setItem("read_model_order_keyword", searchQuery);
  }, [searchQuery]);

  const displayedOrders = useMemo(() => {
    const keyword = searchQuery.trim().toLowerCase();
    if (!keyword) return orders;

    return orders.filter((order) =>
      String(order.order_id || "").toLowerCase().includes(keyword) ||
      String(order.user_id || "").toLowerCase().includes(keyword) ||
      String(order.transaction_id || "").toLowerCase().includes(keyword)
    );
  }, [orders, searchQuery]);

  const uniqueUserIds = useMemo(() => {
    return [...new Set(orders.map((order) => order.user_id).filter(Boolean))];
  }, [orders]);

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2>🍃 MongoDB Read Model Orders</h2>
          <p>
            Dữ liệu đọc nhanh được dựng từ Kafka event <b>payment.completed</b> và lưu trong MongoDB.
          </p>
        </div>

        <button className="btn" onClick={loadOrders} disabled={loading}>
          Làm mới
        </button>
      </div>

      <div className="toolbar" style={{ flexWrap: "wrap", flexDirection: "column", alignItems: "flex-start" }}>
        <input
          list="read-model-user-suggestions"
          style={{ width: "100%", padding: "10px", borderRadius: "4px", border: "1px solid #ccc", fontSize: "1rem" }}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="🔎 Lọc theo User ID, Order ID hoặc Transaction ID..."
        />
        <datalist id="read-model-user-suggestions">
          {uniqueUserIds.map((id) => <option key={id} value={id} />)}
        </datalist>
      </div>

      <ErrorBox error={error} />

      {loading ? <Loading /> : (
        <div className="card">
          <div className="table-header">
            <h3>Read Model từ MongoDB</h3>
            <span>
              Hiển thị: {displayedOrders.length} / {orders.length}
              {meta?.limit ? ` | limit=${meta.limit}` : ""}
            </span>
          </div>

          {!displayedOrders.length ? (
            <div className="empty-state">Chưa có dữ liệu read model hoặc không khớp từ khóa.</div>
          ) : (
            <div className="table-wrapper">
              <table>
                <thead>
                  <tr>
                    <th>Order ID</th>
                    <th>User ID</th>
                    <th>Saga</th>
                    <th>Payment</th>
                    <th>Amount</th>
                    <th>Method</th>
                    <th>Updated</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  {displayedOrders.map((order) => (
                    <tr key={order.order_id}>
                      <td className="mono" style={{ fontSize: "0.85rem" }}>{order.order_id}</td>
                      <td style={{ fontWeight: "bold", color: "#0056b3" }}>{order.user_id}</td>
                      <td><StatusBadge status={order.saga_status} /></td>
                      <td><StatusBadge status={order.payment_status} /></td>
                      <td>{Number(order.amount || 0).toLocaleString()} {order.currency}</td>
                      <td>{order.payment_method}</td>
                      <td>{order.updated_at ? new Date(order.updated_at).toLocaleString() : "-"}</td>
                      <td><Link className="link-btn" to={`/orders/${order.order_id}`}>Trace Saga</Link></td>
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
