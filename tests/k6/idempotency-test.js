import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const idempotencyCorrect = new Rate('idempotency_correct');

export const options = {
  vus: 20,
  iterations: 100,
  thresholds: { idempotency_correct: ['rate>0.99'] },
};

const BASE_URL = 'http://100.65.255.2';
const HEADERS = { 'Content-Type': 'application/json', 'X-API-Key': 'UIT-DOAN-2026-SECRET' };

export default function () {
  const idempotencyKey = `idem-key-${__VU}-${__ITER}`;
  
  // Bổ sung shipping_address để không bị API đá văng
  const payload = JSON.stringify({
    user_id: `user-${__VU}`,
    items: [{ product_id: 'prod-123', quantity: 1 }],
    currency: 'VND',
    payment_method: 'CARD',
    shipping_address: '123 Idempotency Street'
  });

  // Request 1: Gửi lần đầu
  const res1 = http.post(`${BASE_URL}/api/orders`, payload, { 
    headers: { ...HEADERS, 'X-Idempotency-Key': idempotencyKey } 
  });
  
  let orderId1 = null;
  if (res1.status === 200 || res1.status === 201) {
    orderId1 = JSON.parse(res1.body)?.data?.order?.id;
  } else {
    console.log(`Lỗi Req 1: ${res1.status} - ${res1.body}`);
  }

  sleep(0.1);

  // Request 2: Cố tình spam lại cùng 1 Idempotency Key
  const res2 = http.post(`${BASE_URL}/api/orders`, payload, { 
    headers: { ...HEADERS, 'X-Idempotency-Key': idempotencyKey } 
  });
  
  let orderId2 = null;
  if (res2.status === 200 || res2.status === 201) {
    orderId2 = JSON.parse(res2.body)?.data?.order?.id;
  }

  // Nếu cả 2 đều trả về cùng 1 mã OrderID (hoặc Redis chặn trả về đúng data cũ) -> Chính xác 100%
  const isCorrect = (orderId1 !== null) && (orderId2 !== null) && (orderId1 === orderId2);
  
  // Ép kiểu boolean tường minh để K6 không bị lỗi Undefined
  idempotencyCorrect.add(isCorrect ? 1 : 0);
  
  check(res2, { 'idem status 2xx': (r) => r.status >= 200 && r.status < 300 });
}