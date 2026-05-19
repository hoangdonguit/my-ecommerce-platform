# Payment Consumer Throughput Tuning

## Bối cảnh

Trong load test 50 VUs, API Gateway và Order Service tiếp nhận đơn ổn định, error rate 0%. Tuy nhiên, sau khi request tạo đơn thành công, luồng xử lý bất đồng bộ qua Kafka Saga xuất hiện backlog ở tầng Payment Consumer.

## Kết quả trước tối ưu

- K6 load test tạo khoảng 20.694 đơn.
- HTTP error rate: 0%.
- p95 latency: khoảng 125.98ms.
- Inventory xử lý đủ reservation.
- Payment Consumer chỉ scale tối đa 8 replicas.
- Topic `inventory.reserved` có 8 partitions.
- Sau khoảng 10 phút, vẫn còn nhiều đơn ở trạng thái PENDING.
- Bottleneck nằm ở consumer group `payment-service-group`.

## Nguyên nhân

Payment Consumer xử lý message theo mô hình Kafka consumer group. Với 8 partitions, số consumer active thực tế bị giới hạn bởi 8 partition. Khi tốc độ tạo đơn lớn hơn tốc độ xử lý payment, backlog hình thành ở topic `inventory.reserved`.

## Tối ưu đã thực hiện

- Tăng `maxReplicaCount` của Kafka consumers từ 8 lên 16.
- Reset benchmark script hỗ trợ `KAFKA_PARTITIONS=16`.
- Đảm bảo script reset có bước `--alter --partitions` để topic luôn đạt số partition mong muốn.
- Chạy lại load test với cùng kịch bản 50 VUs.

## Kết quả sau tối ưu

- K6 load test tạo khoảng 20.700 đơn.
- HTTP error rate: 0%.
- p95 latency: khoảng 140.61ms.
- Inventory xử lý đủ 20.700 reservation.
- Payment Consumer scale lên 16 replicas.
- Topic `inventory.reserved` có 16 partitions.
- Sau khoảng 10 phút chỉ còn khoảng 118 đơn PENDING.
- Sau thêm khoảng 1 phút, toàn bộ Saga hoàn tất.

## Kết luận

Việc tăng số partition Kafka và giới hạn scale của KEDA giúp cải thiện rõ rệt tốc độ drain backlog ở Payment Consumer. Kết quả cho thấy bottleneck không nằm ở API Gateway hay PostgreSQL, mà nằm ở độ song song của tầng xử lý bất đồng bộ Payment Consumer.
