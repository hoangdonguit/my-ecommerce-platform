import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getOrderSaga } from "../api/gateway";
import ErrorBox from "../components/ErrorBox";
import Loading from "../components/Loading";
import SagaTimeline from "../components/SagaTimeline";
import StatusBadge from "../components/StatusBadge";

const FINAL_STATUSES = new Set([
  "COMPLETED",
  "FAILED_NOTIFIED",
  "INVENTORY_FAILED",
]);

export default function OrderSaga() {
  const { id } = useParams();

  const [saga, setSaga] = useState(null);
  const [loading, setLoading] = useState(true);
  const [polling, setPolling] = useState(true);
  const [error, setError] = useState(null);
  const [lastUpdatedAt, setLastUpdatedAt] = useState(null);

  async function loadSaga() {
    try {
      setError(null);

      const res = await getOrderSaga(id);
      const detail = res.data;

      setSaga(detail);
      setLastUpdatedAt(new Date());

      if (FINAL_STATUSES.has(String(detail?.saga_status || "").toUpperCase())) {
        setPolling(false);
      }
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    let alive = true;
    let timer = null;

    async function tick() {
      if (!alive) return;

      try {
        await loadSaga();
      } finally {
        if (alive && polling) {
          timer = window.setTimeout(tick, 1500);
        }
      }
    }

    tick();

    return () => {
      alive = false;
      if (timer) window.clearTimeout(timer);
    };
  }, [id, polling]);

  function handleManualRefresh() {
    setPolling(true);
    loadSaga();
  }

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2>Order Saga Detail</h2>
          <p className="mono">{id}</p>
        </div>

        <div className="header-actions">
          <button className="btn secondary" onClick={handleManualRefresh}>
            Refresh
          </button>
          <Link className="btn secondary" to="/orders">
            Quay lại
          </Link>
        </div>
      </div>

      <ErrorBox error={error} />

      {loading && !saga ? (
        <Loading />
      ) : saga ? (
        <>
          <div className="grid four">
            <SummaryCard title="Saga Status" value={saga.saga_status} />
            <SummaryCard title="Inventory" value={saga.inventory?.status} />
            <SummaryCard title="Payment" value={saga.payment?.status} />
            <SummaryCard title="Notification" value={saga.notifications?.status} />
          </div>

          {saga.warnings?.length ? (
            <div className="warning-box">
              <b>Warnings</b>
              {saga.warnings.map((item, index) => (
                <p key={index}>{item}</p>
              ))}
            </div>
          ) : null}

          <div className="grid two mt">
            <div className="card">
              <h3>Saga Timeline</h3>
              <SagaTimeline steps={saga.timeline || []} />
            </div>

            <div className="card">
              <h3>Order Summary</h3>
              <KeyValue label="Order ID" value={saga.order?.id} mono />
              <KeyValue label="User ID" value={saga.order?.user_id} />
              <KeyValue label="Order Status" value={<StatusBadge status={saga.order?.status} />} />
              <KeyValue label="Payment Method" value={saga.order?.payment_method} />
              <KeyValue
                label="Total"
                value={`${Number(saga.order?.total_amount || 0).toLocaleString()} ${saga.order?.currency || ""}`}
              />
              <KeyValue label="Shipping" value={saga.order?.shipping_address} />
            </div>
          </div>

          <div className="grid three mt">
            <DataCard title="Inventory Data" state={saga.inventory} />
            <DataCard title="Payment Data" state={saga.payment} />
            <DataCard title="Notification Data" state={saga.notifications} />
          </div>

          <div className="small-muted mt">
            {polling ? "Đang tự động cập nhật mỗi 1.5 giây..." : "Đã dừng polling vì Saga đã đến trạng thái cuối."}
            {lastUpdatedAt ? ` Lần cập nhật cuối: ${lastUpdatedAt.toLocaleTimeString()}` : ""}
          </div>
        </>
      ) : (
        <div className="empty-state">
          Không có dữ liệu Saga.
        </div>
      )}
    </div>
  );
}

function SummaryCard({ title, value }) {
  return (
    <div className="card stat-card">
      <p>{title}</p>
      <h3>
        <StatusBadge status={value || "UNKNOWN"} />
      </h3>
    </div>
  );
}

function KeyValue({ label, value, mono }) {
  return (
    <div className="kv">
      <span>{label}</span>
      <strong className={mono ? "mono" : ""}>{value || "-"}</strong>
    </div>
  );
}

function DataCard({ title, state }) {
  return (
    <div className="card data-card">
      <div className="table-header">
        <h3>{title}</h3>
        <StatusBadge status={state?.status || "UNKNOWN"} />
      </div>

      {state?.reason ? <p className="small-muted">{state.reason}</p> : null}

      <pre>{JSON.stringify(state?.data || {}, null, 2)}</pre>
    </div>
  );
}