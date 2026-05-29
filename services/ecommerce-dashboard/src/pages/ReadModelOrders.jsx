import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { listReadModelOrders } from "../api/gateway";
import ErrorBox from "../components/ErrorBox";
import Loading from "../components/Loading";
import StatusBadge from "../components/StatusBadge";

const LIMIT_OPTIONS = [100, 500, 1000];

function getInitialLimit() {
  const value = Number(sessionStorage.getItem("read_model_orders_limit") || 1000);
  return LIMIT_OPTIONS.includes(value) ? value : 1000;
}

export default function ReadModelOrders() {
  const [searchQuery, setSearchQuery] = useState(
    () => sessionStorage.getItem("read_model_order_keyword") || ""
  );
  const [limit, setLimit] = useState(getInitialLimit);
  const [orders, setOrders] = useState([]);
  const [meta, setMeta] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [lastUpdatedAt, setLastUpdatedAt] = useState(null);

  async function loadOrders(nextLimit = limit) {
    setLoading(true);
    setError(null);

    try {
      const res = await listReadModelOrders(1, nextLimit);
      setOrders(res.data || []);
      setMeta(res.meta || null);
      setLastUpdatedAt(new Date());
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    sessionStorage.setItem("read_model_orders_limit", String(limit));
    loadOrders(limit);
  }, [limit]);

  useEffect(() => {
    sessionStorage.setItem("read_model_order_keyword", searchQuery);
  }, [searchQuery]);

  const displayedOrders = useMemo(() => {
    const keyword = searchQuery.trim().toLowerCase();
    if (!keyword) return orders;

    return orders.filter((order) => {
      const orderId = String(order?.order_id || "").toLowerCase();
      const userId = String(order?.user_id || "").toLowerCase();
      const transactionId = String(order?.transaction_id || "").toLowerCase();
      return orderId.includes(keyword) || userId.includes(keyword) || transactionId.includes(keyword);
    });
  }, [orders, searchQuery]);

  const uniqueUserIds = useMemo(() => {
    return [...new Set(orders.map((order) => order?.user_id).filter(Boolean))];
  }, [orders]);

  const totalDb = Number(meta?.total || 0);

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2>🍃 MongoDB Read Model Orders</h2>
          <p>
            Dữ liệu đọc nhanh được dựng từ Kafka event <b>payment.completed</b> và lưu trong MongoDB.
          </p>
          <div className="refresh-meta">
            {lastUpdatedAt
              ? `Cập nhật lần cuối: ${lastUpdatedAt.toLocaleTimeString()}`
              : "Chưa tải dữ liệu."}
          </div>
        </div>

        <div className="header-actions">
          <label className="limit-control">
            Hiển thị
            <select
              value={limit}
              onChange={(event) => setLimit(Number(event.target.value))}
              disabled={loading}
            >
              {LIMIT_OPTIONS.map((item) => (
                <option key={item} value={item}>
                  {item}
                </option>
              ))}
            </select>
          </label>

          <button className="btn" onClick={() => loadOrders()} disabled={loading}>
            {loading ? "Đang làm mới..." : "↻ Làm mới danh sách"}
          </button>
        </div>
      </div>

      <div className="toolbar-row">
        <input
          list="read-model-user-suggestions"
          value={searchQuery}
          onChange={(event) => setSearchQuery(event.target.value)}
          placeholder="🔎 Lọc theo User ID, Order ID hoặc Transaction ID..."
        />
        <datalist id="read-model-user-suggestions">
          {uniqueUserIds.map((id) => (
            <option key={id} value={id} />
          ))}
        </datalist>
      </div>

      <div className="small-muted read-model-note">
        Lưu ý: MongoDB Read Model là dữ liệu đọc bất đồng bộ. Trace Saga sẽ đối chiếu lại với PostgreSQL source of truth.
      </div>

      <ErrorBox error={error} />

      {loading ? (
        <Loading />
      ) : (
        <div className="card">
          <div className="table-header">
            <h3>Read Model từ MongoDB</h3>
            <span>
              Hiển thị: {displayedOrders.length} / {orders.length} đã tải
              {meta?.limit ? ` | limit=${meta.limit}` : ""}
              {totalDb > 0 ? ` | Tổng DB: ${totalDb}` : ""}
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
                    <th>Hành động</th>
                  </tr>
                </thead>
                <tbody>
                  {displayedOrders.map((order) => (
                    <tr key={order.order_id}>
                      <td className="mono">{order.order_id}</td>
                      <td className="strong-blue">{order.user_id}</td>
                      <td>
                        <StatusBadge status={order.saga_status} />
                      </td>
                      <td>
                        <StatusBadge status={order.payment_status} />
                      </td>
                      <td>
                        {Number(order.amount || 0).toLocaleString()} {order.currency}
                      </td>
                      <td>{order.payment_method}</td>
                      <td>{order.updated_at ? new Date(order.updated_at).toLocaleString() : "-"}</td>
                      <td>
                        <Link className="link-btn" to={`/orders/${order.order_id}`}>
                          Trace Saga
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
