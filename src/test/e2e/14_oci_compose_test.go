// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package test provides e2e tests for Zarf.
package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/defenseunicorns/pkg/helpers/v2"
	"github.com/defenseunicorns/pkg/oci"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zarf-dev/zarf/src/api/v1alpha1"
	"github.com/zarf-dev/zarf/src/pkg/logger"
	"github.com/zarf-dev/zarf/src/pkg/packager/layout"
	"github.com/zarf-dev/zarf/src/pkg/transform"
	"github.com/zarf-dev/zarf/src/pkg/utils"
	"github.com/zarf-dev/zarf/src/pkg/zoci"
	"github.com/zarf-dev/zarf/src/test"
	"github.com/zarf-dev/zarf/src/test/testutil"
	corev1 "k8s.io/api/core/v1"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

type PublishCopySkeletonSuite struct {
	suite.Suite
	*require.Assertions
	Reference   registry.Reference
	PackagesDir string
}

var (
	importEverything      = filepath.Join("src", "test", "packages", "14-import-everything")
	importEverythingPath  string
	importception         = filepath.Join("src", "test", "packages", "14-import-everything", "inception")
	importceptionPath     string
	importRemoteResources = filepath.Join("src", "test", "packages", "14-import-everything", "remote-resources")
)

func (suite *PublishCopySkeletonSuite) SetupSuite() {
	suite.Assertions = require.New(suite.T())

	// This port must match the registry URL in 14-import-everything/zarf.yaml
	suite.Reference.Registry = testutil.SetupInMemoryRegistry(testutil.TestContext(suite.T()), suite.T(), 31888)
	suite.PackagesDir = suite.T().TempDir()
	// Setup the package paths after e2e has been initialized
	importEverythingPath = filepath.Join(suite.PackagesDir, fmt.Sprintf("zarf-package-import-everything-%s-0.0.1.tar.zst", e2e.Arch))
	importceptionPath = filepath.Join(suite.PackagesDir, fmt.Sprintf("zarf-package-importception-%s-0.0.1.tar.zst", e2e.Arch))
}

func (suite *PublishCopySkeletonSuite) TearDownSuite() {
	err := os.RemoveAll(filepath.Join("src", "test", "packages", "14-import-everything", "charts", "local"))
	suite.NoError(err)
	err = os.RemoveAll(importEverythingPath)
	suite.NoError(err)
	err = os.RemoveAll(importceptionPath)
	suite.NoError(err)
}

func (suite *PublishCopySkeletonSuite) Test_0_Publish_Skeletons() {
	suite.T().Log("E2E: Skeleton Package Publish oci://")
	ref := suite.Reference.String()

	helmCharts := filepath.Join("examples", "helm-charts")
	_, _, err := e2e.Zarf(suite.T(), "package", "publish", helmCharts, "oci://"+ref, "--plain-http")
	suite.NoError(err)

	composable := filepath.Join("src", "test", "packages", "09-composable-packages")
	_, _, err = e2e.Zarf(suite.T(), "package", "publish", composable, "oci://"+ref, "--plain-http")
	suite.NoError(err)

	_, _, err = e2e.Zarf(suite.T(), "package", "publish", importRemoteResources, "oci://"+ref, "--plain-http")
	suite.NoError(err)

	_, _, err = e2e.Zarf(suite.T(), "package", "publish", importEverything, "oci://"+ref, "--plain-http")
	suite.NoError(err)

	_, _, err = e2e.Zarf(suite.T(), "package", "inspect", "definition", "oci://"+ref+"/import-everything:0.0.1", "--plain-http", "-a", "skeleton")
	suite.NoError(err)

	_, _, err = e2e.Zarf(suite.T(), "package", "pull", "oci://"+ref+"/import-everything:0.0.1", "-o", suite.PackagesDir, "--plain-http", "-a", "skeleton")
	suite.NoError(err)

	_, _, err = e2e.Zarf(suite.T(), "package", "pull", "oci://"+ref+"/helm-charts:0.0.1", "-o", suite.PackagesDir, "--plain-http", "-a", "skeleton")
	suite.NoError(err)

	_, _, err = e2e.Zarf(suite.T(), "package", "pull", "oci://"+ref+"/test-compose-package:0.0.1", "-o", suite.PackagesDir, "--plain-http", "-a", "skeleton")
	suite.NoError(err)
}

func (suite *PublishCopySkeletonSuite) Test_1_Compose_Everything_Inception() {
	suite.T().Log("E2E: Skeleton Package Compose oci://")

	_, _, err := e2e.Zarf(suite.T(), "package", "create", importEverything, "-o", suite.PackagesDir, "--plain-http", "--confirm")
	suite.NoError(err)

	_, _, err = e2e.Zarf(suite.T(), "package", "create", importception, "-o", suite.PackagesDir, "--plain-http", "--confirm")
	suite.NoError(err)

	stdOut, _, err := e2e.Zarf(suite.T(), "package", "inspect", "definition", importEverythingPath)
	suite.NoError(err)

	targets := []string{
		"file-imports == file-imports",
		"local-chart-import == local-chart-import",
	}

	for _, target := range targets {
		suite.Contains(stdOut, target)
	}
}

func (suite *PublishCopySkeletonSuite) Test_2_FilePaths() {
	suite.T().Log("E2E: Skeleton + Package File Paths")

	pkgTars := []string{
		filepath.Join(suite.PackagesDir, fmt.Sprintf("zarf-package-import-everything-%s-0.0.1.tar.zst", e2e.Arch)),
		filepath.Join(suite.PackagesDir, "zarf-package-import-everything-skeleton-0.0.1.tar.zst"),
		filepath.Join(suite.PackagesDir, fmt.Sprintf("zarf-package-importception-%s-0.0.1.tar.zst", e2e.Arch)),
		filepath.Join(suite.PackagesDir, "zarf-package-helm-charts-skeleton-0.0.1.tar.zst"),
		filepath.Join(suite.PackagesDir, "zarf-package-test-compose-package-skeleton-0.0.1.tar.zst"),
	}

	for _, pkgTar := range pkgTars {
		// Wrap in a fn to ensure our defers cleanup resources on each iteration
		func() {
			var pkg v1alpha1.ZarfPackage

			unpacked := strings.TrimSuffix(pkgTar, ".tar.zst")
			_, _, err := e2e.Zarf(suite.T(), "tools", "archiver", "decompress", pkgTar, unpacked)
			suite.NoError(err)
			suite.DirExists(unpacked)

			// Cleanup resources
			defer func() {
				suite.NoError(os.RemoveAll(unpacked))
			}()
			defer func() {
				suite.NoError(os.RemoveAll(pkgTar))
			}()

			// Verify skeleton contains kustomize-generated manifests.
			if strings.HasSuffix(pkgTar, "zarf-package-test-compose-package-skeleton-0.0.1.tar.zst") {
				kustomizeGeneratedManifests := []string{
					"kustomization-connect-service-0.yaml",
					"kustomization-connect-service-1.yaml",
					"kustomization-connect-service-two-0.yaml",
				}
				ctx := context.Background()
				pkgLayout, err := layout.LoadFromDir(ctx, unpacked, layout.PackageLayoutOptions{
					IsPartial: true,
				})
				suite.NoError(err)
				tmpdir := suite.T().TempDir()
				manifestDir, err := pkgLayout.GetComponentDir(ctx, tmpdir, "test-compose-package", layout.ManifestsComponentDir)
				suite.NoError(err)
				for _, manifest := range kustomizeGeneratedManifests {
					manifestPath := filepath.Join(manifestDir, manifest)
					suite.FileExists(manifestPath, "expected to find kustomize-generated manifest: %q", manifestPath)
					var configMap corev1.ConfigMap
					err := utils.ReadYaml(manifestPath, &configMap)
					suite.NoError(err)
					suite.Equal("ConfigMap", configMap.Kind, "expected manifest %q to be of kind ConfigMap", manifestPath)
				}
			}

			err = utils.ReadYaml(filepath.Join(unpacked, layout.ZarfYAML), &pkg)
			suite.NoError(err)
			suite.NotNil(pkg)

			components := pkg.Components
			suite.NotNil(components)

			isSkeleton := strings.Contains(pkgTar, "-skeleton-")
			suite.verifyComponentPaths(unpacked, components, isSkeleton)
		}()
	}
}

func (suite *PublishCopySkeletonSuite) Test_3_Copy() {
	t := suite.T()
	tmpdir := t.TempDir()

	stdOut, stdErr, err := e2e.Zarf(t, "package", "create", "examples/helm-charts", "-o", tmpdir, "--skip-sbom")
	suite.NoError(err, stdOut, stdErr)
	example := filepath.Join(tmpdir, fmt.Sprintf("zarf-package-helm-charts-%s-0.0.1.tar.zst", e2e.Arch))
	stdOut, stdErr, err = e2e.Zarf(t, "package", "publish", example, "oci://"+suite.Reference.Registry, "--plain-http")
	suite.NoError(err, stdOut, stdErr)

	suite.Reference.Repository = "helm-charts"
	suite.Reference.Reference = "0.0.1"
	ref := suite.Reference.String()

	dstRegistry := testutil.SetupInMemoryRegistry(testutil.TestContext(t), t, 31890)
	dstRef := strings.Replace(ref, suite.Reference.Registry, dstRegistry, 1)
	ctx := testutil.TestContext(t)
	ctx = logger.WithContext(ctx, test.GetLogger(t))

	src, err := zoci.NewRemote(ctx, ref, oci.PlatformForArch(e2e.Arch), oci.WithPlainHTTP(true))
	suite.NoError(err)

	dst, err := zoci.NewRemote(ctx, dstRef, oci.PlatformForArch(e2e.Arch), oci.WithPlainHTTP(true))
	suite.NoError(err)

	reg, err := remote.NewRegistry(strings.Split(dstRef, "/")[0])
	suite.NoError(err)
	reg.PlainHTTP = true
	attempt := 0
	for attempt <= 5 {
		err = reg.Ping(ctx)
		if err == nil {
			break
		}
		attempt++
		time.Sleep(2 * time.Second)
	}
	require.Less(t, attempt, 5, "failed to ping registry")

	err = zoci.CopyPackage(ctx, src, dst, 5)
	suite.NoError(err)

	srcRoot, err := src.FetchRoot(ctx)
	suite.NoError(err)

	for _, layer := range srcRoot.Layers {
		ok, err := dst.Repo().Exists(ctx, layer)
		suite.True(ok)
		suite.NoError(err)
	}
}

func (suite *PublishCopySkeletonSuite) DirOrFileExists(path string) {
	suite.T().Helper()

	invalid := helpers.InvalidPath(path)
	suite.Falsef(invalid, "path specified does not exist: %s", path)
}

func (suite *PublishCopySkeletonSuite) verifyComponentPaths(unpackedPath string, components []v1alpha1.ZarfComponent, isSkeleton bool) {
	suite.T().Helper()
	if isSkeleton {
		suite.NoDirExists(filepath.Join(unpackedPath, "images"))
		suite.NoDirExists(filepath.Join(unpackedPath, "sboms"))
	}
	ctx := context.Background()
	pkgLayout, err := layout.LoadFromDir(ctx, unpackedPath, layout.PackageLayoutOptions{
		IsPartial: isSkeleton,
	})
	suite.NoError(err)

	for _, component := range components {
		if len(component.Charts) == 0 && len(component.Files) == 0 && len(component.Manifests) == 0 && len(component.DataInjections) == 0 && len(component.Repos) == 0 {
			// component has no files to check
			continue
		}

		tmpdir := suite.T().TempDir()

		if isSkeleton && component.DeprecatedCosignKeyPath != "" {
			componentsPath := filepath.Join(unpackedPath, "components")
			base := filepath.Join(unpackedPath, "components", component.Name)
			_, _, err = e2e.Zarf(suite.T(), "tools", "archiver", "decompress", fmt.Sprintf("%s.tar", base), componentsPath)
			suite.NoError(err)
			suite.FileExists(filepath.Join(base, filepath.Base(component.DeprecatedCosignKeyPath)))
		}

		var containsChart bool
		for _, chart := range component.Charts {
			if isSkeleton && chart.URL != "" {
				continue
			}
			containsChart = true
		}
		var chartDir string
		if containsChart {
			chartDir, err = pkgLayout.GetComponentDir(ctx, tmpdir, component.Name, layout.ChartsComponentDir)
			suite.NoError(err)
		}
		for chartIdx, chart := range component.Charts {
			if isSkeleton && chart.URL != "" {
				continue
			} else if isSkeleton {
				dir := fmt.Sprintf("%s-%d", chart.Name, chartIdx)
				suite.DirExists(filepath.Join(chartDir, dir))
				continue
			}
			tgz := fmt.Sprintf("%s-%s.tgz", chart.Name, chart.Version)
			suite.FileExists(filepath.Join(chartDir, tgz))
		}

		var containsFiles bool
		for _, file := range component.Files {
			if isSkeleton && helpers.IsURL(file.Source) {
				continue
			}
			containsFiles = true
		}
		var filesDir string
		if containsFiles {
			filesDir, err = pkgLayout.GetComponentDir(ctx, tmpdir, component.Name, layout.FilesComponentDir)
			suite.NoError(err)
		}
		for filesIdx, file := range component.Files {
			if isSkeleton && helpers.IsURL(file.Source) {
				continue
			}
			path := filepath.Join(filesDir, strconv.Itoa(filesIdx), filepath.Base(file.Target))
			suite.DirOrFileExists(path)
		}

		var containsDataInjections bool
		for _, data := range component.DataInjections {
			if isSkeleton && helpers.IsURL(data.Source) {
				continue
			}
			containsDataInjections = true
		}
		var dataInjectionsDir string
		if containsDataInjections {
			dataInjectionsDir, err = pkgLayout.GetComponentDir(ctx, tmpdir, component.Name, layout.DataComponentDir)
			suite.NoError(err)
		}
		for dataIdx, data := range component.DataInjections {
			if isSkeleton && helpers.IsURL(data.Source) {
				continue
			}
			path := filepath.Join(dataInjectionsDir, strconv.Itoa(dataIdx), filepath.Base(data.Target.Path))
			suite.DirOrFileExists(path)
		}

		var manifestsDir string
		if len(component.Manifests) > 0 {
			manifestsDir, err = pkgLayout.GetComponentDir(ctx, tmpdir, component.Name, layout.ManifestsComponentDir)
			suite.NoError(err)
		}
		for _, manifest := range component.Manifests {
			if isSkeleton {
				suite.Nil(manifest.Kustomizations)
			}
			for filesIdx, path := range manifest.Files {
				if isSkeleton && helpers.IsURL(path) {
					continue
				}
				suite.FileExists(filepath.Join(manifestsDir, fmt.Sprintf("%s-%d.yaml", manifest.Name, filesIdx)))
			}
			for kustomizeIdx := range manifest.Kustomizations {
				path := filepath.Join(manifestsDir, fmt.Sprintf("kustomization-%s-%d.yaml", manifest.Name, kustomizeIdx))
				suite.FileExists(path)
			}
		}

		if !isSkeleton {
			var reposDir string
			if len(component.Repos) > 0 {
				reposDir, err = pkgLayout.GetComponentDir(ctx, tmpdir, component.Name, layout.RepoComponentDir)
				suite.NoError(err)
			}
			for _, repo := range component.Repos {
				dir, err := transform.GitURLtoFolderName(repo)
				suite.NoError(err)
				suite.DirExists(filepath.Join(reposDir, dir))
			}
		}
	}
}

func TestSkeletonSuite(t *testing.T) {
	suite.Run(t, new(PublishCopySkeletonSuite))
}
