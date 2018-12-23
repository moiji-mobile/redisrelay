load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@bazel_gazelle//:def.bzl", "gazelle")

# gazelle:prefix github.com/moiji-mobile/redisrelay
gazelle(name = "gazelle")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/moiji-mobile/redisrelay",
    visibility = ["//visibility:private"],
    deps = [
        "//relay:go_default_library",
        "@com_github_elastic_go_ucfg//:go_default_library",
        "@com_github_elastic_go_ucfg//yaml:go_default_library",
    ],
)

go_binary(
    name = "redisrelay",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
