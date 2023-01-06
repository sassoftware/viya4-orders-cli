// Copyright © 2020-2023, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"log"

	"github.com/sassoftware/viya4-orders-cli/lib/assetreqs"
	"github.com/spf13/cobra"
)

// deploymentAssetsCmd represents the deploymentAssets command
var deploymentAssetsCmd = &cobra.Command{
	Use: "deploymentAssets [order number] [cadence name] [cadence version]",
	Short: "Download deployment assets for the given order number at the given cadence name and version -" +
		" if version not specified, get the latest version of the given cadence name",
	Example: "viya4-orders-cli depassets 993456 stable 2020.0.3\n" +
		"viya4-orders-cli dep 993456 stable\n" +
		"viya4-orders-cli dep 993456 stable -p $HOME/sas -n depAssets_993456_stable",
	Aliases: []string{"depassets", "dep"},
	Args:    cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		cver := ""
		if len(args) == 3 {
			cver = args[2]
		}
		ar := assetreqs.New(token, "deploymentAssets", args[0], args[1], cver, assetFilePath, assetFileName, outFormat, allowUnsuppd)
		err := ar.GetAsset()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deploymentAssetsCmd)
}
