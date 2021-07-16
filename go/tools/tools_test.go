// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package tools_test tests common functions in package tools.
package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func MockModuleVersioning(modSetMap ModuleSetMap, modPathMap ModulePathMap) ModuleVersioning {
	vCfg := versionConfig{
		ModuleSets: modSetMap,
		ExcludedModules: []ModulePath{},
	}

	modInfoMap, _ := vCfg.buildModuleMap()

	return ModuleVersioning{
		ModSetMap: modSetMap,
		ModPathMap: modPathMap,
		ModInfoMap: modInfoMap,
	}
}

func TestMockModuleVersioning(t *testing.T) {
	// prepare
	modSetMap := ModuleSetMap{
		"mod-set-1" : ModuleSet{
			Version: "v1.2.3-RC1+meta",
			Modules: []ModulePath{
				"go.opentelemetry.io/test/test1",
				"go.opentelemetry.io/test/test2",
			},
		},
		"mod-set-2" : ModuleSet{
			Version: "v0.1.0",
			Modules: []ModulePath{
				"go.opentelemetry.io/test3",
			},
		},
	}

	modPathMap := ModulePathMap{
		"go.opentelemetry.io/test/test1" : "root/path/to/mod/test/test1/go.mod",
		"go.opentelemetry.io/test/test2" : "root/path/to/mod/test/test2/go.mod",
		"go.opentelemetry.io/test3" : "root/test3/go.mod",
	}

	expected := ModuleVersioning{
		ModSetMap: modSetMap,
		ModPathMap: modPathMap,
		ModInfoMap: ModuleInfoMap{
			"go.opentelemetry.io/test/test1" : ModuleInfo{
				ModuleSetName: "mod-set-1",
				Version: "v1.2.3-RC1+meta",
			},
			"go.opentelemetry.io/test/test2" : ModuleInfo{
				ModuleSetName: "mod-set-1",
				Version: "v1.2.3-RC1+meta",
			},
			"go.opentelemetry.io/test3" : ModuleInfo{
				ModuleSetName: "mod-set-2",
				Version: "v0.1.0",
			},
		},
	}

	// test
	actual := MockModuleVersioning(modSetMap, modPathMap)

	// verify
	assert.Equal(t, expected, actual)
}

func TestReadVersioningFile(t *testing.T) {
	// prepare
	for _, tt := range []struct{
		VersioningFileName 		string
		ShouldError		  		bool
		ExpectedModuleSets 		ModuleSetMap
		ExpectedExcludedModules []ModulePath
	}{
		{
			VersioningFileName: "./internal/test_data/read_versioning_filename/versions_valid.yaml",
			ShouldError:        false,
			ExpectedModuleSets: ModuleSetMap{
				"mod-set-1": ModuleSet{
					Version: "v1.2.3-RC1+meta",
					Modules: []ModulePath{
						"go.opentelemetry.io/test/test1",
						"go.opentelemetry.io/test/test2",
					},
				},
				"mod-set-2": ModuleSet{
					Version: "v0.1.0",
					Modules: []ModulePath{
						"go.opentelemetry.io/test3",
					},
				},
			},
			ExpectedExcludedModules: []ModulePath{
				"go.opentelemetry.io/excluded1",
			},
		},
		{
			VersioningFileName: "./internal/test_data/read_versioning_filename/versions_invalid_syntax.yaml",
			ShouldError: true,
			ExpectedModuleSets: nil,
			ExpectedExcludedModules: nil,
		},

	}{

		// test
		actual, err := readVersioningFile(tt.VersioningFileName)

		// verify
		if tt.ShouldError {
			assert.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		assert.IsType(t, versionConfig{}, actual)
		assert.Equal(t, tt.ExpectedModuleSets, actual.ModuleSets)
		assert.Equal(t, tt.ExpectedExcludedModules, actual.ExcludedModules)
	}

}

func TestBuildModuleSetsMap(t *testing.T) {
	// prepare
	vCfg := versionConfig{
		ModuleSets: ModuleSetMap{
			"mod-set-1": ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []ModulePath{
					"go.opentelemetry.io/test/test1",
					"go.opentelemetry.io/test/test2",
				},
			},
			"mod-set-2": ModuleSet{
				Version: "v0.1.0",
				Modules: []ModulePath{
					"go.opentelemetry.io/test3",
				},
			},
		},
		ExcludedModules: []ModulePath{
			"go.opentelemetry.io/excluded1",
		},
	}

	expected := ModuleSetMap{
		"mod-set-1": ModuleSet{
			Version: "v1.2.3-RC1+meta",
			Modules: []ModulePath{
				"go.opentelemetry.io/test/test1",
				"go.opentelemetry.io/test/test2",
			},
		},
		"mod-set-2": ModuleSet{
			Version: "v0.1.0",
			Modules: []ModulePath{
				"go.opentelemetry.io/test3",
			},
		},
	}

	// test
	actual, err := vCfg.buildModuleSetsMap()
	require.NoError(t, err)

	// verify
	assert.Equal(t, expected, actual)
}

func TestBuildModuleMap(t *testing.T) {
	// prepare
	vCfg := versionConfig{
		ModuleSets: ModuleSetMap{
			"mod-set-1" : ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []ModulePath{
					"go.opentelemetry.io/test/test1",
					"go.opentelemetry.io/test/test2",
				},
			},
			"mod-set-2" : ModuleSet{
				Version: "v0.1.0",
				Modules: []ModulePath{
					"go.opentelemetry.io/test3",
				},
			},
		},
		ExcludedModules: []ModulePath{
			"go.opentelemetry.io/excluded1",
		},
	}

	expected := ModuleInfoMap{
		"go.opentelemetry.io/test/test1" : ModuleInfo{
			ModuleSetName: "mod-set-1",
			Version: "v1.2.3-RC1+meta",
		},
		"go.opentelemetry.io/test/test2" : ModuleInfo{
			ModuleSetName: "mod-set-1",
			Version: "v1.2.3-RC1+meta",
		},
		"go.opentelemetry.io/test3" : ModuleInfo{
			ModuleSetName: "mod-set-2",
			Version: "v0.1.0",
		},
	}

	// test
	actual, err := vCfg.buildModuleMap()
	require.NoError(t, err)

	// verify
	assert.Equal(t, expected, actual)
}

func TestBuildModuleMap_ModuleDuplicated(t *testing.T) {
	// prepare
	vCfg := versionConfig{
		ModuleSets: ModuleSetMap{
			"mod-set-1" : ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []ModulePath{
					"go.opentelemetry.io/duplicate",
				},
			},
			"mod-set-2" : ModuleSet{
				Version: "v0.1.0",
				Modules: []ModulePath{
					"go.opentelemetry.io/duplicate",
				},
			},
		},
		ExcludedModules: []ModulePath{
			"go.opentelemetry.io/excluded1",
		},
	}

	// test
	m, err := vCfg.buildModuleMap()

	// verify
	assert.Error(t, err)
	assert.Nil(t, m)
}

func TestBuildModuleMap_ModuleExcluded(t *testing.T) {
	// prepare
	vCfg := versionConfig{
		ModuleSets: ModuleSetMap{
			"mod-set-1" : ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []ModulePath{
					"go.opentelemetry.io/excluded",
				},
			},
		},
		ExcludedModules: []ModulePath{
			"go.opentelemetry.io/excluded",
		},
	}

	// test
	m, err := vCfg.buildModuleMap()

	// verify
	assert.Error(t, err)
	assert.Nil(t, m)
}

func TestShouldExcludeModule(t *testing.T) {
	// prepare
	vCfg := versionConfig{
		ModuleSets: ModuleSetMap{
			"mod-set-1" : ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []ModulePath{
					"go.opentelemetry.io/test/test1",
					"go.opentelemetry.io/test/test2",
				},
			},
			"mod-set-2" : ModuleSet{
				Version: "v0.1.0",
				Modules: []ModulePath{
					"go.opentelemetry.io/test3",
				},
			},
		},
		ExcludedModules: []ModulePath{
			"go.opentelemetry.io/excluded1",
		},
	}

	for _, tt := range []struct {
		ModPath ModulePath
		Expected bool
	}{
		{ModPath: "go.opentelemetry.io/test/test1", Expected: false},
		{ModPath: "go.opentelemetry.io/test/test2", Expected: false},
		{ModPath: "go.opentelemetry.io/test3", Expected: false},
		{ModPath: "go.opentelemetry.io/excluded1", Expected: true},
		{ModPath: "go.opentelemetry.io/doesnotexist", Expected: false},
	}{
		// test
		actual := vCfg.shouldExcludeModule(tt.ModPath)

		// verify
		assert.Equal(t, actual, tt.Expected)
	}
}

func TestGetExcludedModules(t *testing.T) {
	// prepare
	vCfg := versionConfig{
		ModuleSets: ModuleSetMap{
			"mod-set-1" : ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []ModulePath{
					"go.opentelemetry.io/test/test1",
					"go.opentelemetry.io/test/test2",
				},
			},
			"mod-set-2" : ModuleSet{
				Version: "v0.1.0",
				Modules: []ModulePath{
					"go.opentelemetry.io/test3",
				},
			},
		},
		ExcludedModules: []ModulePath{
			"go.opentelemetry.io/excluded1",
		},
	}

	expected := excludedModulesSet{
		"go.opentelemetry.io/excluded1": struct{}{},
	}

	// test
	actual := vCfg.getExcludedModules()

	// verify
	assert.Equal(t, expected, actual)
}

func TestBuildModulePathMap(t *testing.T) {
	// prepare
	vCfg := versionConfig{
		ModuleSets: ModuleSetMap{
			"mod-set-1": ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []ModulePath{
					"go.opentelemetry.io/test/test1",
				},
			},
			"mod-set-2": ModuleSet{
				Version: "v0.1.0",
				Modules: []ModulePath{
					"go.opentelemetry.io/test3",
					"go.opentelemetry.io/testroot",
				},
			},
		},
		ExcludedModules: []ModulePath{
			"go.opentelemetry.io/test/testexcluded",
		},
	}

	repoRoot := "./internal/test_data/build_module_path_map"

	expected := ModulePathMap{
		"go.opentelemetry.io/test/test1": ModuleFilePath(filepath.Join(repoRoot, "test", "test1", "go.mod")),
		"go.opentelemetry.io/test3": ModuleFilePath(filepath.Join(repoRoot, "test", "go.mod")),
		"go.opentelemetry.io/testroot": ModuleFilePath(filepath.Join(repoRoot, "go.mod")),
	}

	// test
	actual, err := vCfg.BuildModulePathMap(repoRoot)

	// verify
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestGetModuleSet(t *testing.T) {
	// prepare
	modVersioning := MockModuleVersioning(
		ModuleSetMap{
			"mod-set-1": ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []ModulePath{
					"go.opentelemetry.io/test/test1",
					"go.opentelemetry.io/test/test2",
				},
			},
			"mod-set-2": ModuleSet{
				Version: "v0.1.0",
				Modules: []ModulePath{
					"go.opentelemetry.io/test3",
				},
			},
		},
		ModulePathMap{
			"go.opentelemetry.io/test/test1" : "root/path/to/mod/test/test1/go.mod",
			"go.opentelemetry.io/test/test2" : "root/path/to/mod/test/test2/go.mod",
			"go.opentelemetry.io/test3" : "root/test3/go.mod",
		},
	)

	for _, tt := range []struct{
		ModSetName 	string
		ShouldError bool
		Expected 	ModuleSet
	}{
		{
			ModSetName: "mod-set-1",
			ShouldError: false,
			Expected: ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []ModulePath{
					"go.opentelemetry.io/test/test1",
					"go.opentelemetry.io/test/test2",
				},
			},
		},
		{
			ModSetName: "mod-set-2",
			ShouldError: false,
			Expected: ModuleSet{
				Version: "v0.1.0",
				Modules: []ModulePath{
					"go.opentelemetry.io/test3",
				},
			},
		},
		{
			ModSetName: "mod-set-3-does-not-exist",
			ShouldError: true,
			Expected: ModuleSet{},
		},
	}{
		// test
		m, err := modVersioning.GetModuleSet(tt.ModSetName)
		if tt.ShouldError {
			assert.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		// verify
		assert.Equal(t, tt.Expected, m)
	}
}

func TestCombineModuleTagNamesAndVersion(t *testing.T) {
	// prepare
	modTagNames := []ModuleTagName{
		"tag1",
		"tag2",
		"another/tag3",
		repoRootTag,
	}

	version := "v1.2.3-RC1+meta-RC1"

	expected := []string{
		"tag1/v1.2.3-RC1+meta-RC1",
		"tag2/v1.2.3-RC1+meta-RC1",
		"another/tag3/v1.2.3-RC1+meta-RC1",
		"v1.2.3-RC1+meta-RC1",
	}

	// test
	actual := CombineModuleTagNamesAndVersion(modTagNames, version)

	// verify
	assert.Equal(t, expected, actual)
}

func TestModulePathsToTagNames(t *testing.T) {
	// prepare
	modPaths := []ModulePath{
		"go.opentelemetry.io/test/test1",
		"go.opentelemetry.io/test/test2",
		"go.opentelemetry.io/test3",
		"go.opentelemetry.io/root",
	}

	modPathMap := ModulePathMap{
		"go.opentelemetry.io/test/test1" : "root/path/to/mod/test/test1/go.mod",
		"go.opentelemetry.io/test/test2" : "root/path/to/mod/test/test2/go.mod",
		"go.opentelemetry.io/test3" : "root/test3/go.mod",
		"go.opentelemetry.io/root" : "root/go.mod",
		"go.opentelemetry.io/not-used" : "path/to/mod/not-used/go.mod",
	}

	repoRoot := "root"

	expected := []ModuleTagName{
		"path/to/mod/test/test1",
		"path/to/mod/test/test2",
		"test3",
		repoRootTag,
	}

	// test
	actual, err := ModulePathsToTagNames(modPaths, modPathMap, repoRoot)
	require.NoError(t, err)

	// verify
	assert.Equal(t, expected, actual)
}

func TestModulePathsToFilePaths(t *testing.T) {
	// prepare
	modPaths := []ModulePath{
		"go.opentelemetry.io/test/test1",
		"go.opentelemetry.io/test/test2",
		"go.opentelemetry.io/test3",
		"go.opentelemetry.io/root",
	}

	modPathMap := ModulePathMap{
		"go.opentelemetry.io/test/test1" : "root/path/to/mod/test/test1/go.mod",
		"go.opentelemetry.io/test/test2" : "root/path/to/mod/test/test2/go.mod",
		"go.opentelemetry.io/test3" : "root/test3/go.mod",
		"go.opentelemetry.io/root" : "root/go.mod",
		"go.opentelemetry.io/not-used" : "path/to/mod/not-used/go.mod",
	}

	expected := []ModuleFilePath{
		"root/path/to/mod/test/test1/go.mod",
		"root/path/to/mod/test/test2/go.mod",
		"root/test3/go.mod",
		"root/go.mod",
	}

	// test
	actual, err := modulePathsToFilePaths(modPaths, modPathMap)
	require.NoError(t, err)

	// verify
	assert.Equal(t, expected, actual)
}

func TestModulePathsToFilePaths_ModuleNotInPathMap(t *testing.T) {
	// prepare
	modPaths := []ModulePath{
		"go.opentelemetry.io/in_map",
		"go.opentelemetry.io/not_in_map",
	}

	modPathMap := ModulePathMap{
		"go.opentelemetry.io/in_map" : "root/path/go.mod",
	}

	// test
	_, err := modulePathsToFilePaths(modPaths, modPathMap)

	// verify
	assert.Error(t, err)
}

func TestModuleFilePathToTagName(t *testing.T) {
	// prepare
	repoRoot := "root"

	for _, tt := range []struct{
		ModFilePath ModuleFilePath
		ShouldError bool
		Expected	ModuleTagName
	}{
		{
			ModFilePath: "root/path/to/mod/test/test1/go.mod",
			ShouldError: false,
			Expected: ModuleTagName("path/to/mod/test/test1"),
		},
		{
			ModFilePath: "root/go.mod",
			ShouldError: false,
			Expected: repoRootTag,
		},
		{
			ModFilePath: "no/go/mod/in/path",
			ShouldError: true,
			Expected: "",
		},
		{
			ModFilePath: "not/in/root/go.mod",
			ShouldError: true,
			Expected: "",
		},
	}{
		// test
		actual, err := moduleFilePathToTagName(tt.ModFilePath, repoRoot)

		// verify
		if tt.ShouldError {
			assert.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Equal(t, tt.Expected, actual)
		}
	}
}

func TestModuleFilePathsToTagNames(t *testing.T) {
	// prepare
	modFilePaths := []ModuleFilePath{
		"root/path/to/mod/test/test1/go.mod",
		"root/path/to/mod/test/test2/go.mod",
		"root/test3/go.mod",
		"root/go.mod",
	}

	repoRoot := "root"

	expected := []ModuleTagName{
		"path/to/mod/test/test1",
		"path/to/mod/test/test2",
		"test3",
		repoRootTag,
	}

	// test
	actual, err := moduleFilePathsToTagNames(modFilePaths, repoRoot)
	require.NoError(t, err)

	// verify
	assert.Equal(t, expected, actual)
}

func TestModuleFilePathsToTagNames_Invalid(t *testing.T) {
	// prepare
	modFilePaths := []ModuleFilePath{
		"no/go/mod/in/path",
	}

	repoRoot := "root"

	// test
	actual, err := moduleFilePathsToTagNames(modFilePaths, repoRoot)

	// verify
	assert.Error(t, err)
	assert.Nil(t, actual)
}

func TestIsStableVersion(t *testing.T) {
	// prepare
	for _, tt := range []struct{
		Version	 string
		Expected bool
	}{
		{Version: "v1.0.0", Expected: true},
		{Version: "v1.2.3", Expected: true},
		{Version: "v1.0.0-RC1", Expected: true},
		{Version: "v1.0.0-RC2+MetaData", Expected: true},
		{Version: "v10.10.10", Expected: true},
		{Version: "v0.0.0", Expected: false},
		{Version: "v0.1.2", Expected: false},
		{Version: "v0.20.0", Expected: false},
		{Version: "v0.0.0-RC1", Expected: false},
		{Version: "not-valid-semver", Expected: false},
	}{
		// test
		actual := IsStableVersion(tt.Version)

		// verify
		assert.Equal(t, tt.Expected, actual)
	}
}

func TestFindRepoRoot(t *testing.T) {
	// prepare
	expected, _ := filepath.Abs("../../")

	// test
	actual, err := FindRepoRoot()

	// verify
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestChangeToRepoRoot(t *testing.T) {
	// prepare
	expected, _ := filepath.Abs("../../")

	// test
	actual, err := ChangeToRepoRoot()

	// verify
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	newDir, err := os.Getwd()
	if err != nil {
		t.Logf("could not get current working directory: %v", err)
	}
	assert.Equal(t, expected, newDir)
}
