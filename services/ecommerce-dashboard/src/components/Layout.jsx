import { NavLink, Outlet } from "react-router-dom";

export default function Layout() {
  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-logo">S</div>
          <div>
            <h1>Saga Dashboard</h1>
            <p>E-commerce Microservices</p>
          </div>
        </div>

        <nav className="nav">
          <div style={{ marginTop: "10px", marginBottom: "5px", fontSize: "0.75rem", color: "#888", textTransform: "uppercase", paddingLeft: "12px", letterSpacing: "1px" }}>
            Khu vực Khách hàng
          </div>
          <NavLink to="/store" style={({ isActive }) => isActive ? { color: "#28a745", fontWeight: "bold" } : { color: "#28a745" }}>
            🛒 Cửa hàng Demo
          </NavLink>

          <div style={{ marginTop: "25px", marginBottom: "5px", fontSize: "0.75rem", color: "#888", textTransform: "uppercase", paddingLeft: "12px", letterSpacing: "1px" }}>
            Khu vực Quản trị (Admin)
          </div>
          <NavLink to="/" end>
            📊 Tổng quan Hệ thống
          </NavLink>
          {/* Đã xóa sổ tab Tạo đơn hàng bằng form cũ */}
          <NavLink to="/orders">
            🔍 Tra cứu Đơn hàng
          </NavLink>
          <NavLink to="/orders/error" style={({ isActive }) => isActive ? { color: "#dc3545", fontWeight: "bold" } : { color: "#dc3545" }}>
            ⚠ Giám sát Đơn Lỗi
          </NavLink>
        </nav>
      </aside>

      <main className="main">
        <Outlet />
      </main>
    </div>
  );
}