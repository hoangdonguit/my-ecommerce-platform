import { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { listOrders } from "../api/gateway";
import ErrorBox from "../components/ErrorBox";
import Loading from "../components/Loading";
import StatusBadge from "../components/StatusBadge";

export default function ErrorOrders() {
  const [errorOrders, setErrorOrders] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  async function loadErrorOrders() {
    setLoading(true);
    setError(null);
    try {
      // Dùng keyword bí mật để lấy toàn bộ đơn lỗi từ Backend
      const res = await listOrders("ADMIN_ERROR_ALL", 1, 100); 
      setErrorOrders(res.data || []);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadErrorOrders();
  }, []);

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2>⚠ Giám sát Đơn hàng Lỗi (Admin)</h2>
          <p>Tự động bắt toàn bộ giao dịch thất bại trên hệ thống phân tán.</p>
        </div>
        <button className="btn" onClick={loadErrorOrders} style={{ backgroundColor: "#dc3545" }}>
          ↻ Làm mới danh sách
        </button>
      </div>

      <ErrorBox error={error} />

      {loading ? <Loading /> : (
        <div className="card">
          <div className="table-header">
            <h3 style={{ color: "#dc3545" }}>Giao dịch thất bại / Bị hủy</h3>
            <span>Total: {errorOrders.length}</span>
          </div>
          {!errorOrders.length ? (
            <div className="empty-state">Hệ thống đang hoạt động hoàn hảo! Không có đơn lỗi.</div>
          ) : (
            <div className="table-wrapper">
              <table>
                <thead>
                  <tr>
                    <th>Order ID</th><th>User ID</th><th>Status</th><th>Total</th><th>Created</th><th>Hành động</th>
                  </tr>
                </thead>
                <tbody>
                  {errorOrders.map((order) => (
                    <tr key={order.id}>
                      <td className="mono" style={{fontSize: "0.85rem"}}>{order.id}</td>
                      <td style={{fontWeight: "bold"}}>{order.user_id}</td>
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