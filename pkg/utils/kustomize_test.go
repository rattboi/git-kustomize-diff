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
		"refs2",
		"refs2/components",
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
		"refs2",
		"refs2/components",
	}, dirs)
}

func sameStringSlice(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y] -= 1
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	return len(diff) == 0
}

func TestBuildReferences(t *testing.T) {
	wd, _ := os.Getwd()

	refsMap, err := BuildRefs(filepath.Join(wd, "fixtures", "kustomize"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	expected := map[string][]string{
		"a/kustomization.yaml": []string{
			"refs/kustomization.yaml",
			"refs2/kustomization.yaml",
		},
		"a/pod.yaml": []string{
			"a/kustomization.yaml",
		},
		"b/pod.yaml": []string{
			"b/kustomization.yaml",
		},
		"refs/components/kustomization.yaml": []string{
			"refs/kustomization.yaml",
		},
		"refs/deployment.yaml": []string{
			"refs/kustomization.yaml",
		},
		"refs/pod.yaml": []string{
			"refs/kustomization.yaml",
		},
		"refs/release-patch.yaml": []string{
			"refs/kustomization.yaml",
		},
		"refs2/components/kustomization.yaml": []string{
			"refs2/kustomization.yaml",
		},
		"refs2/deployment.yaml": []string{
			"refs2/kustomization.yaml",
		},
		"refs2/pod.yaml": []string{
			"refs2/kustomization.yaml",
		},
		"refs2/release-patch.yaml": []string{
			"refs2/kustomization.yaml",
		},
	}
	for k, v := range refsMap {
		assert.True(t, sameStringSlice(v, expected[k]))
	}
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

func TestInvertRefs(t *testing.T) {
	refs := InvertRefs(make(map[string][]string))

	assert.Equal(t, map[string][]string{}, refs)

	refs = InvertRefs(map[string][]string{
		"a": []string{"elem1", "elem2"},
		"b": []string{"elem1", "elem2"},
		"c": []string{"elem3", "elem4"},
		"d": []string{"elem3", "elem4"},
		"e": []string{"elem1", "elem2", "elem3", "elem4"},
		"f": []string{"elem1", "elem3"},
		"g": []string{},
	})

	assert.True(t, sameStringSlice(refs["elem1"], []string{"a", "b", "e", "f"}))
	assert.True(t, sameStringSlice(refs["elem2"], []string{"a", "b", "e"}))
	assert.True(t, sameStringSlice(refs["elem3"], []string{"c", "d", "e", "f"}))
	assert.True(t, sameStringSlice(refs["elem4"], []string{"c", "d", "e"}))
}

func TestFindParents(t *testing.T) {
	wd, _ := os.Getwd()

	basePath := filepath.Join(wd, "fixtures", "kustomize")

	// test 1
	changedFile := filepath.Join("a", "pod.yaml")

	parents, err := FindParents(changedFile, basePath)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.True(t, sameStringSlice([]string{
		"refs/kustomization.yaml",
		"refs2/kustomization.yaml",
	}, parents))

	// test 2
	changedFile = filepath.Join("refs", "pod.yaml")

	parents, err = FindParents(changedFile, basePath)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.True(t, sameStringSlice([]string{
		"refs/kustomization.yaml",
	}, parents))

	// test 3
	changedFile = filepath.Join("b", "pod.yaml")

	parents, err = FindParents(changedFile, basePath)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.True(t, sameStringSlice([]string{
		"b/kustomization.yaml",
	}, parents))
}
