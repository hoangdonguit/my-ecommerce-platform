import { useState } from "react";
import { Link } from "react-router-dom";
import { listOrders } from "../api/gateway";
import ErrorBox from "../components/ErrorBox";
import Loading from "../components/Loading";
import StatusBadge from "../components/StatusBadge";

export default function ErrorOrders() {
  const [userId, setUserId] = useState("user-success-001");
  const [errorOrders, setErrorOrders] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  async function loadErrorOrders(event) {
    if (event) event.preventDefault();

    setLoading(true);
    setError(null);

    try {
      // Tải tối đa 100 đơn gần nhất của user
      const res = await listOrders(userId, 1, 100); 
      const allOrders = res.data || [];
      // Chỉ lọc ra những đơn bị lỗi hoặc hủy
      const failed = allOrders.filter(o => o.status === "failed" || o.status === "cancelled");
      setErrorOrders(failed);
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
          <h2>Đơn hàng Lỗi (Saga Failed)</h2>
          <p>Lọc các giao dịch gặp sự cố trong luồng phân tán.</p>
        </div>
      </div>

      <form className="toolbar" onSubmit={loadErrorOrders}>
        <input
          value={userId}
          onChange={(e) => setUserId(e.target.value)}
          placeholder="Nhập user_id"
          required
        />
        <button className="btn" style={{ backgroundColor: "#dc3545" }} type="submit">
          Lọc đơn lỗi
        </button>
      </form>

      <ErrorBox error={error} />

      {loading ? (
        <Loading />
      ) : (
        <div className="card">
          <div className="table-header">
            <h3 style={{ color: "#dc3545" }}>Danh sách giao dịch thất bại</h3>
            <span>Total: {errorOrders.length}</span>
          </div>

          {!errorOrders.length ? (
            <div className="empty-state">
              Chưa có đơn hàng lỗi nào cho User này. Hệ thống hoạt động tốt!
            </div>
          ) : (
            <div className="table-wrapper">
              <table>
                <thead>
                  <tr>
                    <th>Order ID</th>
                    <th>Status</th>
                    <th>Total</th>
                    <th>Created</th>
                    <th>Hành động</th>
                  </tr>
                </thead>
                <tbody>
                  {errorOrders.map((order) => (
                    <tr key={order.id}>
                      <td className="mono">{order.id}</td>
                      <td><StatusBadge status={order.status} /></td>
                      <td>{Number(order.total_amount || 0).toLocaleString()} {order.currency}</td>
                      <td>{new Date(order.created_at).toLocaleString()}</td>
                      <td>
                        <Link className="link-btn" to={`/orders/${order.id}`}>
                          Xem chi tiết lỗi Saga
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