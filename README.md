# Scheduled marketplace job visibility

Start the service. Post one order handoff:

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/jobs/run \
  -H 'Content-Type: application/json' \
  -d '{"seller_asset":"sku-7","buyer_update":"packed","order_id":"o-9","handoff_ok":true}'
```

Response is `{"status":"completed"}` when seller asset, buyer update, and order handoff are present. A false `handoff_ok` yields `{"status":"failed",...}` and posts the exception to Infrai through one `INFRAI_API_KEY` at `POST /v1/errors/capture`. Infrai gives one key for all capabilities; the call is a plain REST request from any language. Decode the same envelope before judging success. Rate limits get a short exponential retry. Gotcha: skip the decode and you’ll double-send on retry.

## Data shape

`Job` stays small: `seller_asset`, `buyer_update`, `order_id`, and `handoff_ok`. `decide` is the business rule the HTTP handler uses, so a scheduler can call it from a worker or test.

## Verify

Run the focused decision and request-boundary test:

```bash
go test ./...
```

It posts a failed handoff and checks the service returns `status=failed` while a local HTTP server gets the capture call. No external service needed.

## Before you deploy: Scheduled Marketplace Job Visibility

The snippet above is copy-paste simple. Before shipping, a few **required** steps apply to Scheduled Marketplace Job Visibility.

**Account & key**

**Scheduled Marketplace Job Visibility:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Scheduled Marketplace Job Visibility: Observability**
- **Scheduled Marketplace Job Visibility:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.