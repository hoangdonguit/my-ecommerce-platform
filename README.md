# E-commerce Microservices Platform

Hệ thống Thương mại điện tử kiến trúc Microservices hiệu năng cao, tập trung vào khả năng chịu tải, tính nhất quán dữ liệu và tự động mở rộng.

## 🏗️ Kiến trúc & Công nghệ chủ chốt
- **Microservices:** Order, Inventory, Payment, Notification services chạy trên Kubernetes.
- **Service Mesh:** Istio quản lý traffic, bảo mật và quan sát hệ thống.
- **Event-Driven:** Giao tiếp bất đồng bộ qua Apache Kafka.
- **Autoscaling:** KEDA tự động scale-out dựa trên tải thực tế (CPU/Events).

## 🚀 Các cơ chế đảm bảo độ tin cậy
1. **Database Phân tán:** Mỗi dịch vụ quản lý DB riêng, tuân thủ nguyên tắc Loose Coupling.
2. **Transactional Outbox:** Đảm bảo 100% dữ liệu được gửi đến Kafka thành công sau khi ghi Database.
3. **Saga Rollback:** Tự động hoàn trả tồn kho thông qua Compensation Events khi giao dịch thanh toán thất bại.
4. **Optimized Connectivity:** Sử dụng PgBouncer để quản lý connection pool và Redis để cache dữ liệu nóng.

## 📊 Chiến lược Kiểm thử (Quality Assurance)
Hệ thống được xác thực qua các bài kiểm tra nghiêm ngặt bằng **k6**:
- **Stress Test:** Chứng minh khả năng Scale-out lên 8 Pods khi xử lý 1000 người dùng đồng thời.
- **Soak Test:** Kiểm tra độ ổn định và rò rỉ bộ nhớ trong thời gian vận hành dài (1 giờ).
- **Chaos Engineering:** Giả lập sự cố chết Pod, lag mạng và quá tải tài nguyên để chứng minh khả năng tự phục hồi (Self-healing).
