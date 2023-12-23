# Domain Exporter
Exporter for prometheus.

## Domain metrics

### Request
```http
GET /probe?target=ya.ru
```
### Response
```prometheus
# HELP domain_free_date_seconds Domain free date
# TYPE domain_free_date_seconds gauge
domain_free_date_seconds{domain="ya.ru"} 2.1801465790705413e+07
# HELP domain_paid_till_seconds Domain paid till
# TYPE domain_paid_till_seconds gauge
domain_paid_till_seconds{domain="ya.ru"} 1.9112265790707015e+07
# HELP domain_success Domain check was successful
# TYPE domain_success gauge
domain_success{domain="ya.ru"} 1
```

## Exporter metrics
### Request
```http
GET /metrics

```
### Response
Default golang metrics from `promhttp`.
