import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

const errorRate = new Rate('error_rate');
const orderDuration = new Trend('order_duration_ms');
const failedOrders = new Counter('failed_orders');
const successOrders = new Counter('success_orders');

export const options = {
  stages: [
    { duration: '1m', target: 100 },   // Khởi động
    { duration: '2m', target: 400 },   // Tải nặng
    { duration: '2m', target: 800 },   // Stress Test (Gần ngưỡng chết)
    { duration: '2m', target: 1000 },  // Đạp ga tối đa (Stress to Death)
    { duration: '1m', target: 0 },     // Hạ nhiệt
  ]
};

const BASE_URL = 'http://100.65.255.2'; // IP Ingress Gateway của ông
const HEADERS = { 'Content-Type': 'application/json', 'X-API-Key': 'UIT-DOAN-2026-SECRET' };

export default function () {
  const userId = `user-stress-${__VU}`;
  const start = Date.now();

  const res = http.post(`${BASE_URL}/api/orders`, JSON.stringify({
    user_id: userId,
    items: [{ product_id: 'prod-456', quantity: 1 }],
    currency: 'VND',
    payment_method: 'CARD', 
    shipping_address: 'Stress Test Address',
    note: `k6 stress test VU:${__VU}`
  }), {
    headers: { ...HEADERS, 'X-Idempotency-Key': `stress-${__VU}-${__ITER}-${Date.now()}` },
    tags: { name: 'stress_order' }
  });

  orderDuration.add(Date.now() - start);
  const success = check(res, { 'status is 2xx': (r) => r.status >= 200 && r.status < 300 });
  errorRate.add(!success);
  if (success) successOrders.add(1); else failedOrders.add(1);
  sleep(1); // Giảm sleep xuống để dồn tải mạnh hơn
}