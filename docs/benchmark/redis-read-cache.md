# Redis Read Cache Benchmark

## Bối cảnh

Sau khi bổ sung MongoDB CQRS Read Model, dashboard có các API đọc dữ liệu thường được gọi lặp lại:

- GET /api/read-model/orders
- GET /api/read-model/orders/:id
- GET /api/inventories

Đây là các API đọc, không thay đổi dữ liệu, phù hợp để áp dụng Redis Cache-Aside với TTL ngắn.

## Trước khi cache ở Web Gateway

Kết quả đo nhanh bằng curl:

### Read model orders

- Cold request: khoảng 0.414s
- Các request sau: khoảng 0.031s - 0.035s

### Inventories

- Cold request: khoảng 0.863s
- Các request sau: khoảng 0.053s - 0.064s

## Sau khi cache ở Web Gateway

Web Gateway bổ sung Redis cache cho:

- GET /api/read-model/orders
- GET /api/read-model/orders/:id
- GET /api/inventories

TTL: 5 giây.

Kết quả kiểm tra header:

GET /api/read-model/orders?limit=100

- Lần 1: X-Cache: MISS
- Lần 2: X-Cache: HIT

GET /api/inventories

- Lần 1: X-Cache: MISS
- Lần 2: X-Cache: HIT

## Kết quả quan sát từ log Web Gateway

Read model orders:

- Request đầu đi thật: khoảng 373ms
- Các request sau cache hit: khoảng 3ms - 5ms

Inventories:

- Request đầu đi thật: khoảng 814ms
- Các request sau cache hit: khoảng 7ms - 9ms

## Kết luận

Redis Cache-Aside giúp giảm thời gian phản hồi cho các API đọc lặp lại trên dashboard, đồng thời giảm số lần gọi xuống read-model-service, MongoDB hoặc inventory-service.

TTL ngắn 5 giây giúp cân bằng giữa hiệu năng và độ mới của dữ liệu. Redis không thay thế database chính, mà chỉ đóng vai trò cache tạm thời cho các truy vấn đọc có thể chấp nhận trễ ngắn.
