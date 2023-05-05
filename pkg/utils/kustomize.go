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
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type ListKustomizeDirsOpts struct {
	IncludeRegexp *regexp.Regexp
	ExcludeRegexp *regexp.Regexp
}

func ListKustomizeDirs(dirPath string, opts ListKustomizeDirsOpts) ([]string, error) {
	targetFiles := make([]string, 0)
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.WithStack(err)
		}
		if !d.IsDir() {
			return nil
		}
		exists, _ := KustomizationExists(path)
		if !exists {
			return nil
		}
		included := true
		if opts.IncludeRegexp != nil {
			m := opts.IncludeRegexp.Match([]byte(path))
			if !m {
				included = false
			}
		}
		if included {
			if opts.ExcludeRegexp != nil {
				m := opts.ExcludeRegexp.Match([]byte(path))
				if m {
					included = false
				}
			}
		}
		if included {
			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return errors.WithStack(err)
			}
			targetFiles = append(targetFiles, relPath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return targetFiles, nil
}

func KustomizationExists(path string) (bool, string) {
	exists1 := Exists(filepath.Join(path, "kustomization.yaml"))
	if exists1 {
		return true, "kustomization.yaml"
	}
	exists2 := Exists(filepath.Join(path, "kustomization.yml"))
	if exists2 {
		return true, "kustomization.yml"
	}
	return false, ""
}

func MakeKustomizeDir(dirPath string) error {
	err := os.MkdirAll(dirPath, 0700)
	if err != nil {
		return errors.WithStack(err)
	}
	kustomizationFilePath := filepath.Join(dirPath, "kustomization.yaml")
	if Exists(kustomizationFilePath) {
		return fmt.Errorf("File already exists: %s", kustomizationFilePath)
	}
	f, err := os.Create(kustomizationFilePath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	return nil
}

type Kustomization struct {
	Resources             []string `yaml:"resources"`
	Components            []string `yaml:"components"`
	PatchesStrategicMerge []string `yaml:"patchesStrategicMerge"`
}

func GetKustomizationRefs(basePath, path string) ([]string, error) {
	exists, f := KustomizationExists(path)
	if !exists {
		return nil, fmt.Errorf("no kustomization file found: %v", path)
	}

	filename, _ := filepath.Abs(filepath.Join(path, f))
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	var kustomization Kustomization

	err = yaml.Unmarshal(yamlFile, &kustomization)
	if err != nil {
		return nil, err
	}

	refs := make([]string, 0)

	// get paths for simple resources
	simpleResources := make([]string, 0)
	simpleResources = append(simpleResources, kustomization.Resources...)
	simpleResources = append(simpleResources, kustomization.Components...)
	simpleResources = append(simpleResources, kustomization.PatchesStrategicMerge...)

	for _, r := range simpleResources {
		rel, err := filepath.Rel(basePath, filepath.Join(path, r))
		if err != nil {
			return nil, err
		}
		candidatePath := filepath.Join(path, r)
		fileInfo, err := os.Stat(candidatePath)
		if err != nil {
			// file not found, just add as relative link anyways
			// maybe remote resource..
			refs = append(refs, rel)
		} else {
			if fileInfo.IsDir() {
				exists, kustomizationFilename := KustomizationExists(candidatePath)
				if !exists {
					return nil, errors.New("No Kustomization found in dir")
				}
				kustomizationPath, err := filepath.Rel(basePath, filepath.Join(candidatePath, kustomizationFilename))
				if err != nil {
					return nil, err
				}
				refs = append(refs, kustomizationPath)
			} else {
				refs = append(refs, rel)
			}
		}
	}

	return refs, nil
}

func BuildRefs(dirPath string) (map[string][]string, error) {
	refMap := make(map[string][]string)

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.WithStack(err)
		}
		if !d.IsDir() {
			return nil
		}
		exists, _ := KustomizationExists(path)
		if !exists {
			return nil
		}
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return errors.WithStack(err)
		}

		absPath := filepath.Join(dirPath, relPath)
		refs, err := GetKustomizationRefs(dirPath, absPath)
		if err != nil {
			return errors.WithStack(err)
		}
		refMap[relPath] = refs
		return nil
	})
	if err != nil {
		return nil, err
	}
	return refMap, nil
}
