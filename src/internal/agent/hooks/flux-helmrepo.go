// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package hooks contains the mutation hooks for the Zarf agent.
package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/defenseunicorns/pkg/helpers/v2"
	"github.com/fluxcd/pkg/apis/meta"
	flux "github.com/fluxcd/source-controller/api/v1"
	"github.com/zarf-dev/zarf/src/config"
	"github.com/zarf-dev/zarf/src/config/lang"
	"github.com/zarf-dev/zarf/src/internal/agent/operations"
	"github.com/zarf-dev/zarf/src/pkg/cluster"
	"github.com/zarf-dev/zarf/src/pkg/logger"
	"github.com/zarf-dev/zarf/src/pkg/transform"
	v1 "k8s.io/api/admission/v1"
)

// NewHelmRepositoryMutationHook creates a new instance of the helm repo mutation hook.
func NewHelmRepositoryMutationHook(ctx context.Context, cluster *cluster.Cluster) operations.Hook {
	return operations.Hook{
		Create: func(r *v1.AdmissionRequest) (*operations.Result, error) {
			return mutateHelmRepo(ctx, r, cluster)
		},
		Update: func(r *v1.AdmissionRequest) (*operations.Result, error) {
			return mutateHelmRepo(ctx, r, cluster)
		},
	}
}

// mutateHelmRepo mutates the repository url to point to the repository URL defined in the ZarfState.
func mutateHelmRepo(ctx context.Context, r *v1.AdmissionRequest, cluster *cluster.Cluster) (*operations.Result, error) {
	l := logger.From(ctx)

	src := &flux.HelmRepository{}
	if err := json.Unmarshal(r.Object.Raw, &src); err != nil {
		return nil, fmt.Errorf(lang.ErrUnmarshal, err)
	}

	// If we see a type of helm repo other than OCI we should flag a warning and return
	if strings.ToLower(src.Spec.Type) != "oci" {
		l.Warn("skipping HelmRepository mutation because the type is not OCI", "type", src.Spec.Type)
		return &operations.Result{Allowed: true}, nil
	}

	zarfState, err := cluster.LoadState(ctx)
	if err != nil {
		return nil, err
	}

	// Get the registry service info if this is a NodePort service to use the internal kube-dns
	registryAddress, err := cluster.GetServiceInfoFromRegistryAddress(ctx, zarfState.RegistryInfo.Address)
	if err != nil {
		return nil, err
	}

	l.Info("using the Zarf registry URL to mutate the Flux HelmRepository",
		"name", src.Name,
		"registry", registryAddress)

	patchedURL := src.Spec.URL

	var (
		isPatched bool

		isCreate = r.Operation == v1.Create
		isUpdate = r.Operation == v1.Update
	)

	// Check if this is an update operation and the hostname is different from what we have in the zarfState
	// NOTE: We mutate on updates IF AND ONLY IF the hostname in the request is different than the hostname in the zarfState
	// NOTE: We are checking if the hostname is different before because we do not want to potentially mutate a URL that has already been mutated.
	if isUpdate {
		zarfStateAddress := helpers.OCIURLPrefix + registryAddress
		isPatched, err = helpers.DoHostnamesMatch(zarfStateAddress, src.Spec.URL)
		if err != nil {
			return nil, fmt.Errorf(lang.AgentErrHostnameMatch, err)
		}
	}

	// Mutate the helm repo URL if necessary
	if isCreate || (isUpdate && !isPatched) {
		patchedSrc, err := transform.ImageTransformHost(registryAddress, src.Spec.URL)
		if err != nil {
			return nil, fmt.Errorf("unable to transform the HelmRepo URL: %w", err)
		}

		patchedRefInfo, err := transform.ParseImageRef(patchedSrc)
		if err != nil {
			return nil, fmt.Errorf("unable to parse the HelmRepo URL: %w", err)
		}
		patchedURL = helpers.OCIURLPrefix + patchedRefInfo.Name
	}

	l.Debug("mutating the Flux HelmRepository URL to the Zarf URL", "original", src.Spec.URL, "mutated", patchedURL)

	var patches []operations.PatchOperation

	patches = populateHelmRepoPatchOperations(patchedURL, zarfState.RegistryInfo.IsInternal())
	patches = append(patches, getLabelPatch(src.Labels))

	return &operations.Result{
		Allowed:  true,
		PatchOps: patches,
	}, nil
}

func populateHelmRepoPatchOperations(repoURL string, isInternal bool) []operations.PatchOperation {
	var patches []operations.PatchOperation
	patches = append(patches, operations.ReplacePatchOperation("/spec/url", repoURL))

	if isInternal {
		patches = append(patches, operations.ReplacePatchOperation("/spec/insecure", true))
	}

	patches = append(patches, operations.AddPatchOperation("/spec/secretRef", meta.LocalObjectReference{Name: config.ZarfImagePullSecretName}))

	return patches
}
