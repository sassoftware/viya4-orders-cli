// Copyright © 2020-2023, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"log"

	"github.com/sassoftware/viya4-orders-cli/lib/assetreqs"
	"github.com/spf13/cobra"
)

// licenseCmd represents the license command
var licenseCmd = &cobra.Command{
	Use:   "license [order number] [cadence name] [cadence version]",
	Short: "Download a license for the given order number at the given cadence name and version	",
	Example: "viya4-orders-cli license 993456 stable 2020.0.3\n" +
		"viya4-orders-cli lic 993456 stable 2020.0.3 -p $HOME/sas -n license_993456_stable_2020.0.3",
	Aliases: []string{"lic"},
	Args:    cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		ar := assetreqs.New(token, "license", args[0], args[1], args[2], "", assetFilePath, assetFileName, outFormat, allowUnsuppd)
		err := ar.GetAsset()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(licenseCmd)
}
