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

## Failing the build on a policy

Define a policy in the console, then add `--gate`:

```shell
trivy repo --format json -o scan.json .
rscli trivy upload-trivy-container-image-scan -f scan.json --gate
```

```
Uploaded . (repository) to https://api.reysys.com

Policy: FAILED
  target   https://github.com/org/repo
  59 finding(s) at HIGH or above with a fix available; the budget is 0

  FAIL  severity_budget  59 finding(s) at HIGH or above with a fix available; the budget is 0
  ok    kev              0 finding(s) on the CISA Known Exploited Vulnerabilities catalogue
  ok    secret           0 secret(s) committed to the repository
```

Every rule is reported, not just the one that failed, so fixing the first does
not surface a second you were never told about.

**Without `--gate` the verdict is printed and the command still succeeds.** Run
it that way first. A team that meets a blocking gate for the first time as an
unexplained outage turns it off; one that has watched it for a week turns it on.

## Large reports

The ingest is synchronous: an upload waits for the server to store the whole
report, not for the bytes to travel. A container image with a thousand packages
can take minutes.

`RS_HTTP_TIMEOUT` sets the ceiling, default `10m`:

```shell
RS_HTTP_TIMEOUT=20m rscli trivy upload-trivy-container-image-scan -f scan.json --gate
```

If an upload does time out, **the server may still have finished storing it** —
check the console before assuming otherwise. rscli reports it as a tool error
(exit 1), never as a policy failure, so a gated pipeline is not failed by it.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Uploaded, and passed — or no policy is enforced for this target |
| 1 | The tool could not do its job: bad config, upload failed, or Reysys could not evaluate the scan |
| 2 | The artifact did not meet the policy |
| 3 | The scan could not be judged, and the policy says not to allow that |

Code 1 is never a statement about your code. If our side cannot evaluate a scan
we allow the build and report it as a tool error, because a pipeline that reds
during our outage — under a message saying the code failed a security policy —
is a pipeline whose owner stops gating.

A failed upload also fails the step: a build whose scan never reached Reysys
should not look the same as one that did.

## Which CI systems this works with

Any of them. The gate is an HTTP call and an exit code, so it runs wherever a
binary runs — GitHub Actions, GitLab CI, Azure DevOps, Jenkins, Bitbucket,
CircleCI. What differs between them is only the install step and how the secret
is supplied. `setup-rscli-action` exists for GitHub Actions; elsewhere,
`go install` and the provider's own secret store.

## Development

```shell
go mod tidy && gofmt -s -w . && go vet ./... && go build ./...
```

Set `RS_BASE_URL` and `RS_TOKEN_URL` to point at a local stack. `RS_INSECURE_SKIP_VERIFY=true`
skips TLS verification for a self-signed local certificate — never set it in CI.
