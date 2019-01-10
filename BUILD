load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@bazel_gazelle//:def.bzl", "gazelle")
load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar", "pkg_deb")

# gazelle:prefix github.com/moiji-mobile/redisrelay
gazelle(name = "gazelle")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/moiji-mobile/redisrelay",
    visibility = ["//visibility:private"],
    deps = [
        "//relay/proto:go_default_library",
        "//relay:go_default_library",
        "@com_github_golang_protobuf//jsonpb:go_default_library_gen",
    ],
)

go_binary(
    name = "redisrelay",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

pkg_tar(
    name = "bazel-bin",
    package_dir = "/usr/bin",
    srcs = [":redisrelay"],
    mode = "0755",
)

pkg_tar(
    name = "debian-data",
    extension = "tar.gz",
    deps = [
       ":bazel-bin",
    ],
)

pkg_deb(
    name = "redisrelay-debian",
    package = "redisrelay",
    data = ":debian-data",
    maintainer = "you?",
    version = "1",
    description = "Redis Relay for weak consistency fan-in/fan-out",
    architecture = "amd64",
)
