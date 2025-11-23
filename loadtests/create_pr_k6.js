import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
    stages: [
        { duration: '10s', target: 10 },
        { duration: '20s', target: 50 },
        { duration: '20s', target: 100 },
        { duration: '10s', target: 0 },
    ],
};

let prCounter = 0;

export default function () {
    const id = `pr-load-${__VU}-${prCounter++}`; // уникальный id из VU + счётчика
    const payload = JSON.stringify({
        pull_request_id: id,
        pull_request_name: 'Load test PR',
        author_id: 'u1',
    });

    const headers = { 'Content-Type': 'application/json' };

    const res = http.post(`${BASE_URL}/pullRequest/create`, payload, { headers });

    check(res, {
        'status is 201': (r) => r.status === 201,
    });

    sleep(0.1);
}