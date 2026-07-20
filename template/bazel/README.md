# metrics_tools

Offline tooling for verifying export archive integrity. Built with Bazel
because it shares (planned) codegen with the data-plane team's monorepo.

```sh
bazel build //:checksum_tool
bazel test //:checksum_test
```
