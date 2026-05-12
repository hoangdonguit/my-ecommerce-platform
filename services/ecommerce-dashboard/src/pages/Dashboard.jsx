import { useEffect, useState } from "react";
import { getServicesHealth } from "../api/gateway";
import Loading from "../components/Loading";

export default function Dashboard() {
  const [health, setHealth] = useState(null);
  const [loading, setLoading] = useState(true);

  async function loadHealth() {
    setLoading(true);
    try {
      const res = await getServicesHealth();
      setHealth(res.data);
    } catch (error) {
      console.error(error);
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
        </div>
        <button className="btn" onClick={loadHealth}>Làm mới</button>
      </div>

      {loading && !health ? <Loading /> : (
        <div className="grid">
          {["order_service", "inventory_service", "payment_service", "notification_service"].map(srv => {
            const status = health?.[srv];
            const isOk = status?.ok;

            return (
              <div key={srv} className="card">
                <p style={{ color: "gray", fontSize: "0.9rem", marginBottom: "10px" }}>
                  {srv.replace("_", " ").toUpperCase()}
                </p>
                <h3 style={{ marginBottom: "10px" }}>{isOk ? "Online" : "Offline"}</h3>
                <span className={`badge ${isOk ? "badge-success" : "badge-danger"}`}>
                  {isOk ? "SUCCESS" : "FAILED"}
                </span>
                {!isOk && status?.error && (
                  <p style={{ marginTop: "15px", fontSize: "0.8rem", color: "red", wordBreak: "break-all" }}>
                    {status.error}
                  </p>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}