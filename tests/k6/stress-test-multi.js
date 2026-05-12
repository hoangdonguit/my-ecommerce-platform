import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

const errorRate = new Rate('error_rate');
const orderDuration = new Trend('order_duration_ms');
const failedOrders = new Counter('failed_orders');
const successOrders = new Counter('success_orders');

export const options = {
  stages: [
    { duration: '2m', target: 200 },
    { duration: '2m', target: 500 },
    { duration: '3m', target: 1000 },
    { duration: '2m', target: 0 },
  ],
  thresholds: {
    'http_req_failed': ['rate<0.1'],
    'http_req_duration': ['p(95)<5000'],
    'error_rate': ['rate<0.1'],
  }
};

const BASE_URL = 'http://100.65.255.2';
const HEADERS = {
  'Content-Type': 'application/json',
  'X-API-Key': 'UIT-DOAN-2026-SECRET'
};

const PRODUCT_CATALOG = ['prod-123', 'prod-456', 'prod-789'];

function getRandomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function shuffle(arr) {
  let a = [...arr];
  for (let i = a.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [a[i], a[j]] = [a[j], a[i]];
  }
  return a;
}

export default function () {
  const userId = `user-stress-${__VU}`;
  const idempotencyKey = `${userId}-${__ITER}-${Date.now()}`;

  const numberOfItemsToBuy = getRandomInt(1, 3);
  const shuffledCatalog = shuffle(PRODUCT_CATALOG);

  let cartItems = [];
  for (let i = 0; i < numberOfItemsToBuy; i++) {
    cartItems.push({
      product_id: shuffledCatalog[i],
      quantity: getRandomInt(1, 2)
    });
  }

  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/orders`,
    JSON.stringify({
      user_id: userId,
      items: cartItems,
      currency: 'VND',
      payment_method: 'CARD',
      shipping_address: 'Stress Test Address',
      note: `k6 stress VU:${__VU} ITER:${__ITER}`
    }),
    {
      headers: {
        ...HEADERS,
        'X-Idempotency-Key': idempotencyKey
      },
      timeout: '10s',
    }
  );

  const duration = Date.now() - start;
  orderDuration.add(duration);

  const isSuccess = check(res, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
  });

  if (!isSuccess) {
    errorRate.add(1);
    failedOrders.add(1);
  } else {
    errorRate.add(0);
    successOrders.add(1);
  }

  sleep(1);
}
