# CFFI Go Proxy

Based on https://github.com/elazarl/goproxy which was vendored in to replace the `net/http` dependency to `github.com/bogdanfinn/fhttp`, making the requests with `github.com/bogdanfinn/tls-client`

## Running the code

```bash
go run main.go
```

## Running the benchmark

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

```bash
python benchmark/test.py
```

Looking at the results, we validate that the proxy was able to add the `x-returned-by` header to the response.
We can also validate that the proxy was able to intercept the request, and add the `X-Hello-World` header before sending it to HTTPBIN

Currently the benchmark only works for `verify=False`, however the proxy service could be configured to trust a certain CA certificate.
If the client also trusts that CA certificate, `verify=False` is no longer needed.
