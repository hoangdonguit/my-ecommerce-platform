export default function Loading({ text = "Đang tải dữ liệu..." }) {
  return (
    <div className="loading">
      <div className="spinner" />
      <span>{text}</span>
    </div>
  );
}