// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package cmd contains the CLI commands for Zarf.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/zarf-dev/zarf/src/cmd/helm"
	"github.com/zarf-dev/zarf/src/config/lang"
	"github.com/zarf-dev/zarf/src/pkg/logger"
	"helm.sh/helm/v3/pkg/action"
)

// ldflags github.com/zarf-dev/zarf/src/cmd.helmVersion=x.x.x
var helmVersion string

func newHelmCommand() *cobra.Command {
	actionConfig := new(action.Configuration)

	// Truncate Helm's arguments so that it thinks its all alone
	helmArgs := []string{}
	if len(os.Args) > 2 {
		helmArgs = os.Args[3:]
	}
	// The inclusion of Helm in this manner should be changed once https://github.com/helm/helm/pull/13617 is merged
	cmd, err := helm.NewRootCmd(actionConfig, os.Stdout, helmArgs)
	if err != nil {
		logger.Default().Debug("failed to initialize helm command", "error", err)
	}
	cmd.Short = lang.CmdToolsHelmShort
	cmd.Long = lang.CmdToolsHelmLong
	cmd.AddCommand(newToolsVersionCmd("helm", helmVersion))

	return cmd
}
