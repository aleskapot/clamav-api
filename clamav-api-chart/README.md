# clamav-api Helm Chart

A Helm chart for deploying the ClamAV antivirus API service.

## Installation

```bash
helm install clamav-api ./clamav-api-chart
```

## Example values.yaml

```yaml
replicaCount: 2

image:
  repository: aleskapot/clamav-api
  tag: "latest"
  pullPolicy: Always

imagePullSecrets:
  - name: private-registry-secret

service:
  type: ClusterIP
  port: 80

env:
  CLAMAV_HOST: "clamav.host"
  CLAMAV_PORT: "30331"
  WEBHOOK_URL: "https://webhook.url:8080/"

secret:
  existingSecret: "clamav-api-auth"
  create: false
  authApiKey: ""
```

## Notes

- The `AUTH_API_KEY` is expected to be stored in a Kubernetes Secret (by default `clamav-api-auth`) under the key `AUTH_API_KEY`.
- If you wish the chart to create the secret, set `secret.create: true` and provide `secret.authApiKey`.