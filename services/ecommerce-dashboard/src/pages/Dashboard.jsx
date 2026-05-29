import { useEffect, useState } from "react";
import { getServicesHealth } from "../api/gateway";
import ErrorBox from "../components/ErrorBox";
import Loading from "../components/Loading";

const SERVICES = [
  ["order_service", "ORDER SERVICE"],
  ["inventory_service", "INVENTORY SERVICE"],
  ["payment_service", "PAYMENT SERVICE"],
  ["notification_service", "NOTIFICATION SERVICE"],
];

export default function Dashboard() {
  const [health, setHealth] = useState({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [lastUpdatedAt, setLastUpdatedAt] = useState(null);

  async function loadHealth() {
    setLoading(true);
    setError(null);

    try {
      const res = await getServicesHealth();
      setHealth(res.data || {});
      setLastUpdatedAt(new Date());
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadHealth();
  }, []);

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2>Dashboard Tổng Quan</h2>
          <p>Trạng thái kết nối của các microservices trong cụm Kubernetes.</p>
          <div className="refresh-meta">
            {lastUpdatedAt
              ? `Cập nhật lần cuối: ${lastUpdatedAt.toLocaleTimeString()}`
              : "Chưa có dữ liệu cập nhật."}
          </div>
        </div>

        <button className="btn" onClick={loadHealth} disabled={loading}>
          {loading ? "Đang làm mới..." : "↻ Làm mới trạng thái"}
        </button>
      </div>

      <ErrorBox error={error} />

      {loading && !lastUpdatedAt ? (
        <Loading />
      ) : (
        <div className="grid">
          {SERVICES.map(([key, label]) => {
            const status = health?.[key];
            const isOk = Boolean(status?.ok);

            return (
              <div key={key} className="card">
                <p className="service-name">{label}</p>
                <h3 className="service-status">{isOk ? "Online" : "Offline"}</h3>
                <span className={`badge ${isOk ? "badge-success" : "badge-danger"}`}>
                  {isOk ? "SUCCESS" : "FAILED"}
                </span>
                {!isOk && status?.error ? (
                  <p className="service-error">{status.error}</p>
                ) : null}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
