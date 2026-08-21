# rscli

Command-line client for [Reysys](https://reysys.com). Uploads a Trivy scan report
from a pipeline (or a laptop) to your Reysys account, where it is scored and
prioritised alongside everything else you scan.

## Install

```shell
go install github.com/reysys-technology/rscli/cmd/rscli@latest
```

This is how `setup-rscli-action` installs it, so it is the path that gets the most
use. Signed binaries and a container image at `ghcr.io/reysys-technology/rscli`
are published for some releases; check the releases page for the version you want.

## Credentials

rscli authenticates with an OAuth2 client-credentials pair. Provision one in the
console under **Account Info → API client**. The secret is shown once, so copy it
then; if you lose it, rotate rather than provision again.

```shell
export RS_CLIENT_ID=...
export RS_CLIENT_SECRET=...
```

Check it works — this uploads nothing:

```shell
rscli account get-account-information
```

`rscli configure` prints the full reference: every variable, the config-file form,
and the precedence between them.

## Upload a scan

```shell
trivy image --format json -o scan.json ghcr.io/acme/api:1.4.2
rscli trivy upload-trivy-container-image-scan -f scan.json
```

Source repositories work the same way — the report says which it is, so the same
command handles both:

```shell
trivy repo --format json -o scan.json .
rscli trivy upload-trivy-container-image-scan -f scan.json
```

## In a pipeline

Keep the credentials in the CI provider's secret store, never in the repository.

**GitHub Actions**

```yaml
- uses: reysys-technology/setup-rscli-action@v1
- run: trivy image --format json -o scan.json "$IMAGE"
- run: rscli trivy upload-trivy-container-image-scan -f scan.json
  env:
    RS_CLIENT_ID: ${{ secrets.RS_CLIENT_ID }}
    RS_CLIENT_SECRET: ${{ secrets.RS_CLIENT_SECRET }}
```

**GitLab CI**

```yaml
scan:
  script:
    - trivy image --format json -o scan.json "$IMAGE"
    - rscli trivy upload-trivy-container-image-scan -f scan.json
  variables:
    RS_CLIENT_ID: $RS_CLIENT_ID
    RS_CLIENT_SECRET: $RS_CLIENT_SECRET
```

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Uploaded |
| 1 | Failed — the message on stderr says why |

A failed upload fails the pipeline step, which is usually what you want: a build
whose scan never reached Reysys should not look the same as one that did.

## Development

```shell
go mod tidy && gofmt -s -w . && go vet ./... && go build ./...
```

Set `RS_BASE_URL` and `RS_TOKEN_URL` to point at a local stack. `RS_INSECURE_SKIP_VERIFY=true`
skips TLS verification for a self-signed local certificate — never set it in CI.
