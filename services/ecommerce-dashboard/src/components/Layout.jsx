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
          <NavLink to="/" end>
            Dashboard
          </NavLink>
          <NavLink to="/orders/new">
            Tạo đơn hàng
          </NavLink>
          <NavLink to="/orders">
            Danh sách đơn
          </NavLink>
          {/* Thêm tab Đơn hàng lỗi ở đây */}
          <NavLink to="/orders/error" style={({ isActive }) => isActive ? { color: "#dc3545", fontWeight: "bold" } : { color: "#dc3545" }}>
            ⚠ Đơn hàng lỗi
          </NavLink>
        </nav>
      </aside>

      <main className="main">
        <Outlet />
      </main>
    </div>
  );
}