import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { listOrders } from "../api/gateway";
import ErrorBox from "../components/ErrorBox";
import Loading from "../components/Loading";
import StatusBadge from "../components/StatusBadge";

const LIMIT_OPTIONS = [100, 500, 1000];

function getInitialLimit() {
  const value = Number(sessionStorage.getItem("orders_limit") || 1000);
  return LIMIT_OPTIONS.includes(value) ? value : 1000;
}

export default function Orders() {
  const [searchQuery, setSearchQuery] = useState(
    () => sessionStorage.getItem("search_order_keyword") || ""
  );
  const [limit, setLimit] = useState(getInitialLimit);
  const [allOrders, setAllOrders] = useState([]);
  const [meta, setMeta] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [lastUpdatedAt, setLastUpdatedAt] = useState(null);

  async function loadOrders(nextLimit = limit) {
    setLoading(true);
    setError(null);

    try {
      const res = await listOrders("ADMIN_FETCH_ALL", 1, nextLimit);
      setAllOrders(res.data || []);
      setMeta(res.meta || null);
      setLastUpdatedAt(new Date());
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    sessionStorage.setItem("orders_limit", String(limit));
    loadOrders(limit);
  }, [limit]);

  useEffect(() => {
    sessionStorage.setItem("search_order_keyword", searchQuery);
  }, [searchQuery]);

  const displayedOrders = useMemo(() => {
    const keyword = searchQuery.trim().toLowerCase();
    if (!keyword) return allOrders;

    return allOrders.filter((order) => {
      const userId = String(order?.user_id || "").toLowerCase();
      const orderId = String(order?.id || "").toLowerCase();
      return userId.includes(keyword) || orderId.includes(keyword);
    });
  }, [allOrders, searchQuery]);

  const uniqueUserIds = useMemo(() => {
    return [...new Set(allOrders.map((order) => order?.user_id).filter(Boolean))];
  }, [allOrders]);

  const totalDb = Number(meta?.total || 0);

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2>🔍 Tra cứu Đơn hàng Toàn Hệ Thống</h2>
          <p>Hiển thị tất cả đơn hàng. Gõ User ID hoặc Order ID để lọc nhanh.</p>
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
          list="user-suggestions"
          value={searchQuery}
          onChange={(event) => setSearchQuery(event.target.value)}
          placeholder="🔎 Nhập mã User ID hoặc Order ID để lọc tức thì..."
        />
        <datalist id="user-suggestions">
          {uniqueUserIds.map((id) => (
            <option key={id} value={id} />
          ))}
        </datalist>
      </div>

      <ErrorBox error={error} />

      {loading ? (
        <Loading />
      ) : (
        <div className="card">
          <div className="table-header">
            <h3>Danh sách Giao dịch</h3>
            <span>
              Hiển thị: {displayedOrders.length} / {allOrders.length} đã tải
              {totalDb > 0 ? ` | Tổng DB: ${totalDb}` : ""}
            </span>
          </div>

          {!displayedOrders.length ? (
            <div className="empty-state">Không tìm thấy đơn hàng nào khớp với tìm kiếm.</div>
          ) : (
            <div className="table-wrapper">
              <table>
                <thead>
                  <tr>
                    <th>Order ID</th>
                    <th>User ID</th>
                    <th>Status</th>
                    <th>Total</th>
                    <th>Created</th>
                    <th>Hành động</th>
                  </tr>
                </thead>
                <tbody>
                  {displayedOrders.map((order) => (
                    <tr key={order.id}>
                      <td className="mono">{order.id}</td>
                      <td className="strong-blue">{order.user_id}</td>
                      <td>
                        <StatusBadge status={order.status} />
                      </td>
                      <td>
                        {Number(order.total_amount || 0).toLocaleString()} {order.currency}
                      </td>
                      <td>{order.created_at ? new Date(order.created_at).toLocaleString() : "-"}</td>
                      <td>
                        <Link className="link-btn" to={`/orders/${order.id}`}>
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
