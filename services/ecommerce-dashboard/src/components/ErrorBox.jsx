export default function ErrorBox({ error }) {
  if (!error) return null;

  return (
    <div className="error-box">
      <strong>Có lỗi xảy ra</strong>
      <p>{error.message || String(error)}</p>
    </div>
  );
}