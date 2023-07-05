# h2client
http2 client used to test Kong Gateway

## Semantics

The tools is used internally by Kong and Kong Gateway through `spec.helpers` tool, in the
[http2_client function](https://github.com/Kong/kong/blob/99e33e39cf3a5c097c768cab6eb96fcb16678639/spec/helpers.lua#L1023).

It's possible to set `KONG_TEST_DEBUG_HTTP2=1` (which sets `GODEBUG=http2debug=1` to the go program) to expose useful
HTTP/2 transport information.

The response is encoded in JSON and parsed by the helper tool. 


## Release

Developer with write access to this repo is able to create a `tag` locally, and push the `tag` to Github.

The tag should start with `v` and follow semantic versioning, like `v0.1.0`.

After the tag is created, the [release workflow](https://github.com/Kong/h2client/actions/workflows/release.yml) will automatically
build, create a release and upload artifacts.
