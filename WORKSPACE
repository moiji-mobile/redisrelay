load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")
http_archive(
    name = "io_bazel_rules_go",
    urls = ["https://github.com/bazelbuild/rules_go/releases/download/0.16.5/rules_go-0.16.5.tar.gz"],
    sha256 = "7be7dc01f1e0afdba6c8eb2b43d2fa01c743be1b9273ab1eaf6c233df078d705",
)
http_archive(
    name = "bazel_gazelle",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/releases/download/0.16.0/bazel-gazelle-0.16.0.tar.gz"],
    sha256 = "7949fc6cc17b5b191103e97481cf8889217263acf52e00b560683413af204fcb",
)
load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains")
go_rules_dependencies()
go_register_toolchains()
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
gazelle_dependencies()

# dependencies

# Config handling
go_repository(
    name = "com_github_elastic_go_ucfg",
    commit = "92d43887f91851c9936621665af7f796f4d03412",  # Version as of 2018-12-23
    importpath = "github.com/elastic/go-ucfg",
)

# Logging and dependencies
go_repository(
    name = "org_uber_go_zap",
    commit = "ff33455a0e382e8a81d14dd7c922020b6b5e7982",  # Version 1.9.1 as of 2018-12-23
    importpath = "go.uber.org/zap",
)

go_repository(
    name = "org_uber_go_atomic",
    commit = "8dc6146f7569370a472715e178d8ae31172ee6da",  # Version as of 2018-12-23
    importpath = "go.uber.org/atomic",
)

go_repository(
    name = "org_uber_go_multierr",
    commit = "ddea229ff1dff9e6fe8a6c0344ac73b09e81fce5",  # Version as of 2018-12-23
    importpath = "go.uber.org/multierr",
)
