function getStepClass(status) {
  const normalized = String(status || "").toLowerCase();

  if (normalized === "success") return "timeline-step success";
  if (normalized === "failed") return "timeline-step failed";
  if (normalized === "pending") return "timeline-step pending";
  if (normalized === "waiting") return "timeline-step waiting";

  return "timeline-step";
}

export default function SagaTimeline({ steps = [] }) {
  if (!steps.length) {
    return (
      <div className="empty-state">
        Chưa có timeline.
      </div>
    );
  }

  return (
    <div className="timeline">
      {steps.map((step, index) => (
        <div className={getStepClass(step.status)} key={`${step.key}-${index}`}>
          <div className="timeline-dot">
            {step.status === "success" ? "✓" : step.status === "failed" ? "!" : index + 1}
          </div>

          <div className="timeline-content">
            <div className="timeline-title-row">
              <h3>{step.label}</h3>
              <span>{step.status}</span>
            </div>

            {step.description ? (
              <p>{step.description}</p>
            ) : null}
          </div>
        </div>
      ))}
    </div>
  );
}