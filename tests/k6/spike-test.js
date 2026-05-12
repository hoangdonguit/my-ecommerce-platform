import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 50 },
    { duration: '30s', target: 1000 }, // FLASH SALE: 1000 user ập vào trong 30s
    { duration: '1m', target: 1000 },
    { duration: '20s', target: 0 },
  ],
};

const BASE_URL = 'http://100.65.255.2'; 
const HEADERS = { 
  'Content-Type': 'application/json', 
  'X-API-Key': 'UIT-DOAN-2026-SECRET' // Cần API Key để qua Gateway
};

export default function () {
  const payload = JSON.stringify({
    user_id: `flash-sale-${__VU}`,
    items: [{ product_id: 'prod-123', quantity: 1 }],
    currency: 'VND',
    payment_method: 'COD',
    shipping_address: 'Flash Sale St',
    note: 'K6 Spike Test'
  });

  const res = http.post(`${BASE_URL}/api/orders`, payload, { headers: HEADERS });
  check(res, { 'status is 2xx': (r) => r.status >= 200 && r.status < 300 });
}