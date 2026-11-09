package apptest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// manifest is the part of AppLoad's external.manifest.json that decides whether a program opens.
type manifest struct {
	Application string `json:"application"`
	Framebuffer bool   `json:"qtfb"`
}

func TestEveryProgramsManifestNamesItsOwnBinaryAndAsksForAFramebuffer(test *testing.T) {
	paths, err := filepath.Glob("../cmd/*/external.manifest.json")
	if err != nil || len(paths) == 0 {
		test.Fatalf("no manifests found: %v", err)
	}
	for _, path := range paths {
		contents, err := os.ReadFile(path)
		if err != nil {
			test.Fatal(err)
		}
		var read manifest
		if err := json.Unmarshal(contents, &read); err != nil {
			test.Fatalf("%s: %v", path, err)
		}
		if read.Application != filepath.Base(filepath.Dir(path)) || !read.Framebuffer {
			test.Errorf("%s: %+v", path, read)
		}
	}
}
