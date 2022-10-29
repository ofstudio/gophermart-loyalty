#!/usr/bin/env sh

BASE_URL=http://accrual:8080

curl -vs --location "${BASE_URL}/api/goods" \
    -H "Content-Type: application/json" \
    -d "{\"match\": \"Pizza\",\"reward\": 10,\"reward_type\": \"%\"}"

curl -vs --location "${BASE_URL}/api/orders" \
    -H "Content-Type: application/json" \
    -d "{\"order\": \"01008\",\"goods\": [{ \"description\": \"Pizza Margarita\",\"price\": 100}]}"

curl -vs --location "${BASE_URL}/api/orders" \
    -H "Content-Type: application/json" \
    -d "{\"order\": \"01016\",\"goods\": [{ \"description\": \"Pizza Pepperoni\",\"price\": 110}]}"

curl -vs --location "${BASE_URL}/api/orders" \
    -H "Content-Type: application/json" \
    -d "{\"order\": \"01024\",\"goods\": [{ \"description\": \"Pizza BBQ\",\"price\": 120}]}"

curl -vs --location "${BASE_URL}/api/orders" \
    -H "Content-Type: application/json" \
    -d "{\"order\": \"01032\",\"goods\": [{ \"description\": \"Pizza New York\",\"price\": 130}]}"

curl -vs --location "${BASE_URL}/api/orders" \
    -H "Content-Type: application/json" \
    -d "{\"order\": \"01040\",\"goods\": [{ \"description\": \"Pizza Chicken\",\"price\": 140}]}"

curl -vs --location "${BASE_URL}/api/orders" \
    -H "Content-Type: application/json" \
    -d "{\"order\": \"01057\",\"goods\": [{ \"description\": \"Pizza Hawaii\",\"price\": 150}]}"

curl -vs --location "${BASE_URL}/api/orders" \
    -H "Content-Type: application/json" \
    -d "{\"order\": \"01065\",\"goods\": [{ \"description\": \"Pizza Mexico\",\"price\": 160}]}"

curl -vs --location "${BASE_URL}/api/orders" \
    -H "Content-Type: application/json" \
    -d "{\"order\": \"01073\",\"goods\": [{ \"description\": \"Pizza Roma\",\"price\": 170}]}"

curl -vs --location "${BASE_URL}/api/orders" \
    -H "Content-Type: application/json" \
    -d "{\"order\": \"01081\",\"goods\": [{ \"description\": \"Pizza Supreme\",\"price\": 180}]}"

curl -vs --location "${BASE_URL}/api/orders" \
    -H "Content-Type: application/json" \
    -d "{\"order\": \"01099\",\"goods\": [{ \"description\": \"Pizza Deluxe\",\"price\": 190}]}"
