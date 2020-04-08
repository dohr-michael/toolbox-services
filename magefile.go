// +build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func init() {}

func downloadDeps() error {
	return sh.RunV("go", "mod", "download")
}

func Protoc() error {
	return sh.RunV("protoc", "-I", "proto", "proto/*.proto", "--go_out=plugins=grpc:proto")
}

func Generate() error {
	return sh.RunV("go", "generate", "./...")
}

func Build() error {
	mg.Deps(downloadDeps, Protoc, Generate)
	return sh.RunV("go", "build", "./...")
}
