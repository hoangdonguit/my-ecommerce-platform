import { useState, useEffect } from "react";
import { createOrder, getInventories } from "../api/gateway";
import ErrorBox from "../components/ErrorBox";
import { useSaga } from "../context/SagaContext"; // MÓC CONTEXT VÀO ĐÂY ĐỂ GIỮ DATA

export default function Storefront() {
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState(null);
  const [error, setError] = useState(null);
  const [checkoutItem, setCheckoutItem] = useState(null);
  const [demoType, setDemoType] = useState("none");

  // DÙNG DATA TỪ CONTEXT THAY VÌ STATE CỤC BỘ (Chuyển tab không bị mất)
  const { products, setProducts } = useSaga();

  // Nếu kho rỗng (lần đầu vào web), set mặc định
  useEffect(() => {
    if (!products || products.length === 0) {
      setProducts([
        { id: "prod-123", name: "Laptop ASUS TUF Gaming F15", price: 24000000, icon: "💻", stock: 0 },
        { id: "prod-456", name: "Bàn phím cơ Keychron", price: 2500000, icon: "⌨️", stock: 0 },
        { id: "prod-789", name: "Chuột Gaming Logitech", price: 1200000, icon: "🖱️", stock: 0 },
      ]);
    }
  }, []);

  async function fetchRealInventory() {
    try {
      const res = await getInventories();
      if (res && res.data) {
        setProducts(prevProducts => {
          if (!prevProducts || prevProducts.length === 0) return prevProducts;
          return prevProducts.map(p => {
            const realData = res.data.find(item => (item.product_id || item.productId) === p.id);
            return realData ? { ...p, stock: realData.available_quantity ?? realData.availableQuantity ?? 0 } : p;
          });
        });
      }
    } catch (err) {
      console.error("Lỗi đồng bộ kho hàng:", err);
    }
  }

  // Tự động đồng bộ kho mỗi 10 giây
  useEffect(() => {
    fetchRealInventory();
    const interval = setInterval(fetchRealInventory, 10000);
    return () => clearInterval(interval);
  }, []);

  const [formData, setFormData] = useState({ shipping_address: "", payment_method: "COD", note: "", quantity: 1 });
  const [currentUserId] = useState(() => {
    let id = sessionStorage.getItem("storefront_user_id") || "customer-" + Math.floor(Math.random() * 9000 + 1000);
    sessionStorage.setItem("storefront_user_id", id);
    return id;
  });

  function handleOpenCheckout(product) {
    setCheckoutItem(product);
    setFormData({ shipping_address: "", payment_method: "COD", note: "", quantity: 1 });
    setDemoType("none");
    setMessage(null);
    setError(null);
  }

  async function submitOrder(e) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setMessage(null);
    
    try {
      let finalUserId = currentUserId;
      let finalQuantity = formData.quantity;
      let finalPaymentMethod = formData.payment_method;
      let finalProductId = checkoutItem.id;

      if (demoType === "inventory") { 
          finalProductId = "sanphamsieucapvippro-789"; 
          finalQuantity = 1; 
      } else if (demoType === "payment") { 
          finalUserId = "blocked-user-001"; 
          finalPaymentMethod = "CARD"; 
      }

      const payload = {
        user_id: finalUserId,
        product_id: finalProductId, 
        quantity: finalQuantity, 
        currency: "VND",
        payment_method: finalPaymentMethod,
        shipping_address: formData.shipping_address,
        note: demoType !== "none" ? `Demo Error: ${demoType}` : formData.note
      };
      
      const res = await createOrder(payload);
      
      if (demoType === "none") {
          setMessage(`🎉 Mua thành công ${formData.quantity} [${checkoutItem.name}]. Mã đơn: ${res.data.order.id}`);
          setProducts(prev => prev.map(p => p.id === checkoutItem.id ? { ...p, stock: p.stock - formData.quantity } : p));
      } else {
          setMessage(`🚧 Request Demo [${demoType.toUpperCase()}] đã gửi. Mã đơn: ${res.data.order.id}.`);
          setTimeout(fetchRealInventory, 3000);
      }

      setCheckoutItem(null);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="page" style={{ padding: "20px" }}>
      <div className="page-header" style={{ marginBottom: "20px", borderBottom: "1px solid #eee", paddingBottom: "10px" }}>
        <h2>🛒 Cửa hàng trực tuyến</h2>
        <p>Phiên làm việc của: <strong style={{color: "#28a745"}}>{currentUserId}</strong></p>
      </div>
      
      <ErrorBox error={error} />
      
      {message && (
        <div style={{ 
          padding: "15px", 
          backgroundColor: message.includes("🚧") ? "#fff3cd" : "#d4edda", 
          color: message.includes("🚧") ? "#856404" : "#155724",
          borderRadius: "8px", 
          marginBottom: "20px",
          border: `1px solid ${message.includes("🚧") ? "#ffeeba" : "#c3e6cb"}`
        }}>
          {message}
        </div>
      )}

      {/* CHỈNH LẠI FLEXBOX CHỖ NÀY ĐỂ CÁC NÚT LUÔN THẲNG HÀNG */}
      <div style={{ display: "flex", gap: "20px", flexWrap: "wrap", alignItems: "stretch" }}>
        {(products || []).map(p => (
          <div key={p.id} className="card" style={{ 
            width: "300px", 
            padding: "25px", 
            display: "flex", 
            flexDirection: "column", 
            boxShadow: "0 4px 6px rgba(0,0,0,0.1)", 
            borderRadius: "12px" 
          }}>
            <div style={{ textAlign: "center", flexGrow: 1 }}>
              <div style={{ fontSize: "60px", marginBottom: "15px" }}>{p.icon}</div>
              <h3 style={{ margin: "10px 0", minHeight: "56px" }}>{p.name}</h3>
              <div style={{ margin: "15px 0" }}>
                <span style={{ 
                  backgroundColor: p.stock > 0 ? "#e8f5e9" : "#ffe3e3", 
                  color: p.stock > 0 ? "#2e7d32" : "#c62828",
                  padding: "5px 12px",
                  borderRadius: "20px",
                  fontSize: "0.9rem",
                  fontWeight: "bold"
                }}>
                  {p.stock === 0 && !message ? "Đang tải..." : `Còn lại: ${p.stock}`}
                </span>
              </div>
            </div>

            {/* PHẦN GIÁ VÀ NÚT BẤM BỊ ÉP XUỐNG ĐÁY */}
            <div style={{ marginTop: "auto", textAlign: "center" }}>
              <h3 style={{ color: "#007bff", marginBottom: "20px" }}>{p.price.toLocaleString()} VND</h3>
              <button 
                className="btn" 
                onClick={() => handleOpenCheckout(p)} 
                disabled={p.stock <= 0}
                style={{ 
                  width: "100%", 
                  padding: "12px", 
                  backgroundColor: p.stock <= 0 ? "#ccc" : "#007bff",
                  color: "white",
                  border: "none",
                  borderRadius: "8px",
                  cursor: p.stock <= 0 ? "not-allowed" : "pointer"
                }}
              >
                {p.stock <= 0 ? "Hết hàng" : "Mua Ngay"}
              </button>
            </div>
          </div>
        ))}
      </div>

      {checkoutItem && (
        <div style={{
          position: "fixed", top: 0, left: 0, right: 0, bottom: 0,
          backgroundColor: "rgba(0,0,0,0.5)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1000
        }}>
          <div className="card" style={{ width: "450px", padding: "30px", backgroundColor: "white", borderRadius: "15px", maxHeight: "90vh", overflowY: "auto" }}>
            <h3 style={{ marginBottom: "20px", textAlign: "center" }}>Xác nhận mua hàng</h3>
            <form onSubmit={submitOrder}>
               <div style={{ marginBottom: "15px" }}>
                  <label style={{ display: "block", marginBottom: "5px", fontWeight: "bold" }}>Số lượng:</label>
                  <input 
                    type="number" min="1" 
                    style={{ width: "100%", padding: "10px", borderRadius: "5px", border: "1px solid #ddd" }}
                    value={formData.quantity} 
                    onChange={e => setFormData({...formData, quantity: parseInt(e.target.value) || 1})}
                    disabled={demoType !== "none"}
                  />
               </div>
               <div style={{ marginBottom: "15px" }}>
                  <label style={{ display: "block", marginBottom: "5px", fontWeight: "bold" }}>Địa chỉ giao hàng:</label>
                  <input 
                    type="text" required 
                    style={{ width: "100%", padding: "10px", borderRadius: "5px", border: "1px solid #ddd" }}
                    value={formData.shipping_address} 
                    onChange={e => setFormData({...formData, shipping_address: e.target.value})}
                  />
               </div>
               <div style={{ marginBottom: "15px" }}>
                  <label style={{ display: "block", marginBottom: "5px", fontWeight: "bold" }}>Thanh toán:</label>
                  <select 
                    style={{ width: "100%", padding: "10px", borderRadius: "5px", border: "1px solid #ddd" }}
                    value={formData.payment_method} 
                    onChange={e => setFormData({...formData, payment_method: e.target.value})}
                    disabled={demoType === "payment"}
                  >
                    <option value="COD">Thanh toán khi nhận hàng (COD)</option>
                    <option value="CARD">Thẻ Tín dụng / Ghi nợ (CARD)</option>
                  </select>
               </div>

               <div style={{ padding: "15px", backgroundColor: "#fff8e1", borderRadius: "8px", border: "1px solid #ffe082", marginBottom: "20px" }}>
                  <strong style={{ display: "block", marginBottom: "12px", color: "#f57c00" }}>🛠️ Tùy chọn Demo Saga:</strong>
                  <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
                    <label style={{ display: "flex", alignItems: "center", cursor: "pointer", gap: "10px", margin: 0 }}>
                      <input type="radio" name="demo" style={{ margin: 0, width: "18px", height: "18px", cursor: "pointer" }} checked={demoType==="none"} onChange={()=>setDemoType("none")}/> 
                      <span style={{ fontSize: "1rem", lineHeight: "1" }}>🟢 Bình thường</span>
                    </label>
                    <label style={{ display: "flex", alignItems: "center", cursor: "pointer", gap: "10px", margin: 0 }}>
                      <input type="radio" name="demo" style={{ margin: 0, width: "18px", height: "18px", cursor: "pointer" }} checked={demoType==="inventory"} onChange={()=>setDemoType("inventory")}/> 
                      <span style={{ fontSize: "1rem", lineHeight: "1" }}>🔴 Ép lỗi Kho (Trừ kho rồi Rollback)</span>
                    </label>
                    <label style={{ display: "flex", alignItems: "center", cursor: "pointer", gap: "10px", margin: 0 }}>
                      <input type="radio" name="demo" style={{ margin: 0, width: "18px", height: "18px", cursor: "pointer" }} checked={demoType==="payment"} onChange={()=>setDemoType("payment")}/> 
                      <span style={{ fontSize: "1rem", lineHeight: "1" }}>🔴 Ép lỗi Thanh toán (Trừ kho rồi Rollback)</span>
                    </label>
                  </div>
               </div>
               
               <div style={{ display: "flex", gap: "15px" }}>
                  <button type="submit" className="btn" disabled={loading} style={{ flex: 1, padding: "12px", backgroundColor: "#28a745", color: "white", border: "none", borderRadius: "8px" }}>
                    {loading ? "Đang xử lý..." : "Xác nhận"}
                  </button>
                  <button type="button" className="btn" onClick={()=>setCheckoutItem(null)} style={{ flex: 1, padding: "12px", backgroundColor: "#6c757d", color: "white", border: "none", borderRadius: "8px" }}>
                    Hủy
                  </button>
               </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
