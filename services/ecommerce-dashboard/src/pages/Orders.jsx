import { useState, useEffect, useMemo } from "react";
import { Link } from "react-router-dom";
import { listOrders } from "../api/gateway";
import ErrorBox from "../components/ErrorBox";
import Loading from "../components/Loading";
import StatusBadge from "../components/StatusBadge";

export default function Orders() {
  const [searchQuery, setSearchQuery] = useState(() => sessionStorage.getItem("search_order_keyword") || "");
  const [allOrders, setAllOrders] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    setLoading(true);
    listOrders("ADMIN_FETCH_ALL", 1, 100)
      .then(res => setAllOrders(res.data || []))
      .catch(err => setError(err))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    sessionStorage.setItem("search_order_keyword", searchQuery);
  }, [searchQuery]);

  const displayedOrders = useMemo(() => {
    if (!searchQuery.trim()) return allOrders;
    return allOrders.filter(o => 
      o.user_id.toLowerCase().includes(searchQuery.toLowerCase()) || 
      o.id.toLowerCase().includes(searchQuery.toLowerCase())
    );
  }, [allOrders, searchQuery]);

  const uniqueUserIds = useMemo(() => {
    return [...new Set(allOrders.map(o => o.user_id))];
  }, [allOrders]);

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2>🔍 Tra cứu Đơn hàng Toàn Hệ Thống</h2>
          <p>Hiển thị tất cả. Gõ User ID hoặc Order ID để lọc nhanh.</p>
        </div>
      </div>

      <div className="toolbar" style={{ flexWrap: "wrap", flexDirection: "column", alignItems: "flex-start" }}>
        <input
          list="user-suggestions"
          style={{ width: "100%", padding: "10px", borderRadius: "4px", border: "1px solid #ccc", fontSize: "1rem" }}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="🔎 Nhập mã User ID hoặc Order ID để lọc tức thì..."
        />
        <datalist id="user-suggestions">
          {uniqueUserIds.map(id => <option key={id} value={id} />)}
        </datalist>
      </div>

      <ErrorBox error={error} />

      {loading ? <Loading /> : (
        <div className="card">
          <div className="table-header">
            <h3>Danh sách Giao dịch</h3>
            <span>Hiển thị: {displayedOrders.length} / {allOrders.length}</span>
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
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  {displayedOrders.map((order) => (
                    <tr key={order.id}>
                      <td className="mono" style={{fontSize: "0.85rem"}}>{order.id}</td>
                      <td style={{fontWeight: "bold", color: "#0056b3"}}>{order.user_id}</td>
                      <td><StatusBadge status={order.status} /></td>
                      <td>{Number(order.total_amount || 0).toLocaleString()} {order.currency}</td>
                      <td>{new Date(order.created_at).toLocaleString()}</td>
                      <td><Link className="link-btn" to={`/orders/${order.id}`}>Trace Saga</Link></td>
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