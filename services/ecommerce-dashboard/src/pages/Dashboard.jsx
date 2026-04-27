import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getServicesHealth } from "../api/gateway";
import StatusBadge from "../components/StatusBadge";
import Loading from "../components/Loading";
import ErrorBox from "../components/ErrorBox";

export default function Dashboard() {
  const [health, setHealth] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  async function loadHealth() {
    try {
      setError(null);
      const res = await getServicesHealth();
      setHealth(res.data);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadHealth();
  }, []);

  const services = [
    ["Order Service", "order_service"],
    ["Inventory Service", "inventory_service"],
    ["Payment Service", "payment_service"],
    ["Notification Service", "notification_service"],
  ];

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2>Dashboard</h2>
          <p>Theo dõi luồng Saga của hệ thống xử lý đơn hàng phân tán.</p>
        </div>

        <button className="btn secondary" onClick={loadHealth}>
          Refresh
        </button>
      </div>

      <ErrorBox error={error} />

      {loading ? (
        <Loading />
      ) : (
        <div className="grid four">
          {services.map(([label, key]) => {
            const item = health?.[key];
            const ok = item?.ok === true;

            return (
              <div className="card stat-card" key={key}>
                <p>{label}</p>
                <h3>{ok ? "Online" : "Offline"}</h3>
                <StatusBadge status={ok ? "SUCCESS" : "FAILED"} />
                {!ok && item?.error ? <small>{item.error}</small> : null}
              </div>
            );
          })}
        </div>
      )}

      <div className="grid two mt">
        <div className="card">
          <h3>Demo thành công</h3>
          <p>
            Tạo order với payment method <b>COD</b>. Luồng kỳ vọng:
          </p>
          <div className="flow-line">
            Order Created → Inventory Reserved → Payment Completed → Notification Sent
          </div>
          <Link className="btn" to="/orders/new?demo=success">
            Tạo order success
          </Link>
        </div>

        <div className="card">
          <h3>Demo thất bại thanh toán</h3>
          <p>
            Tạo order với user <b>blocked-user-001</b> và payment method <b>CARD</b>.
          </p>
          <div className="flow-line">
            Order Created → Inventory Reserved → Payment Failed → Notification Sent
          </div>
          <Link className="btn danger" to="/orders/new?demo=failed">
            Tạo order failed
          </Link>
        </div>
      </div>
    </div>
  );
}