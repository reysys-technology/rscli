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

rscli authenticates with an OAuth2 client-credentials pair. Create one in the
console under **Account Info → API clients**, giving it a name that says what
will hold it (`frontend-ci`, `nightly-scan`). The secret is shown once, so copy
it then; if you lose it, rotate that client rather than creating another.

> **A client secret is an administrator credential for your account.**
> It is not scoped to uploading scans: anything the API can do for your account,
> it can do — read every finding, change integrations, delete data. Treat it the
> way you would treat an admin password.
>
> What follows from that:
> - Keep it in the CI provider's secret store, never in the repository, and
>   never in a file the pipeline checks out.
> - Give each pipeline its own named client. When one leaks you rotate that one;
>   the others keep running.
> - On a pull-request pipeline, remember that the branch being built is
>   attacker-controlled. Do not expose the secret to jobs that run code from a
>   fork.
> - Delete clients you no longer use. A forgotten credential is the one nobody
>   notices being used.
>
> rscli refuses to send either your credentials or the resulting access token
> over plaintext http: `RS_TOKEN_URL` and `RS_BASE_URL` must both be https
> (loopback http is allowed, for local development). Redirecting either one was
> a way to walk off with a full-access token from a single CI variable.

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

Keep the credentials in the CI provider's secret store, never in the repository —
and see the warning under [Credentials](#credentials) about what one grants.

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
