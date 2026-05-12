import { BrowserRouter, Route, Routes } from "react-router-dom";
import Layout from "./components/Layout";
import Dashboard from "./pages/Dashboard";
import Orders from "./pages/Orders";
import OrderSaga from "./pages/OrderSaga";
import ErrorOrders from "./pages/ErrorOrders";
import Storefront from "./pages/Storefront";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/orders" element={<Orders />} />
          <Route path="/orders/error" element={<ErrorOrders />} />
          <Route path="/orders/:id" element={<OrderSaga />} />
          {/* Storefront đã được đưa vào trong Layout để có Sidebar */}
          <Route path="/store" element={<Storefront />} /> 
        </Route>
      </Routes>
    </BrowserRouter>
  );
}