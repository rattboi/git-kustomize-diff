/*
Copyright 2021 Daisuke Taniwaki.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListKustomizeDirs(t *testing.T) {
	wd, _ := os.Getwd()

	dirs, err := ListKustomizeDirs(filepath.Join(wd, "fixtures", "kustomize"), ListKustomizeDirsOpts{})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, []string{
		"a",
		"b",
		"refs",
		"refs/components",
	}, dirs)

	includeRegexp, _ := regexp.Compile(".*/a$")
	dirs, err = ListKustomizeDirs(filepath.Join(wd, "fixtures", "kustomize"), ListKustomizeDirsOpts{IncludeRegexp: includeRegexp})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, []string{
		"a",
	}, dirs)

	excludeRegexp, _ := regexp.Compile(".*/a$")
	dirs, err = ListKustomizeDirs(filepath.Join(wd, "fixtures", "kustomize"), ListKustomizeDirsOpts{ExcludeRegexp: excludeRegexp})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, []string{
		"b",
		"refs",
		"refs/components",
	}, dirs)
}

func TestBuildReferences(t *testing.T) {
	wd, _ := os.Getwd()

	refsMap, err := BuildRefs(filepath.Join(wd, "fixtures", "kustomize"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	expected := map[string][]string{
		"a": []string{
			"a/pod.yaml",
		},
		"b": []string{
			"b/pod.yaml",
		},
		"refs": []string{
			"refs/pod.yaml",
			"refs/deployment.yaml",
			"a/kustomization.yaml",
			"refs/components/kustomization.yaml",
			"refs/release-patch.yaml",
		},
		"refs/components": []string{},
	}
	assert.Equal(t, expected, refsMap)
}

func TestGetKustomizationRefs(t *testing.T) {
	wd, _ := os.Getwd()

	k, err := GetKustomizationRefs(wd, filepath.Join(wd, "fixtures", "kustomize", "a"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, []string{"fixtures/kustomize/a/pod.yaml"}, k)

	k2, err := GetKustomizationRefs(wd, filepath.Join(wd, "fixtures", "kustomize", "b"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, []string{"fixtures/kustomize/b/pod.yaml"}, k2)

	k3, err := GetKustomizationRefs(wd, filepath.Join(wd, "fixtures", "kustomize", "refs"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, []string{
		"fixtures/kustomize/refs/pod.yaml",
		"fixtures/kustomize/refs/deployment.yaml",
		"fixtures/kustomize/a/kustomization.yaml",
		"fixtures/kustomize/refs/components/kustomization.yaml",
		"fixtures/kustomize/refs/release-patch.yaml",
	}, k3)
}
