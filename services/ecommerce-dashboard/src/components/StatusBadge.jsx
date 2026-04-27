function normalizeStatus(status) {
  return String(status || "UNKNOWN").toUpperCase();
}

export default function StatusBadge({ status }) {
  const normalized = normalizeStatus(status);

  let className = "badge badge-neutral";

  if (["SUCCESS", "SENT", "COMPLETED", "RESERVED", "FAILED_NOTIFIED", "PAYMENT_COMPLETED"].includes(normalized)) {
    className = "badge badge-success";
  }

  if (["FAILED", "PAYMENT_FAILED", "INVENTORY_FAILED", "CANCELLED"].includes(normalized)) {
    className = "badge badge-danger";
  }

  if (["PENDING", "PROCESSING"].includes(normalized)) {
    className = "badge badge-warning";
  }

  if (["WAITING", "UNKNOWN"].includes(normalized)) {
    className = "badge badge-muted";
  }

  return <span className={className}>{normalized}</span>;
}