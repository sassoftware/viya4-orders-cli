// Copyright © 2020-2023, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"log"

	"github.com/sassoftware/viya4-orders-cli/lib/assetreqs"
	"github.com/spf13/cobra"
)

// certificatesCmd represents the certificates command
var certificatesCmd = &cobra.Command{
	Use:   "certificates [order number]",
	Short: "Download certificates for the given order number",
	Example: "viya4-orders-cli certs 993456\n" +
		"viya4-orders-cli certs 993456 -p $HOME/sas",
	Aliases: []string{"certs", "cer"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Cadence is not a factor in certs, so we hard-code allowUnsuppd to false for the last argument.
		ar := assetreqs.New(token, "certificates", args[0], "", "",
			assetFilePath, assetFileName, outFormat, false)
		err := ar.GetAsset()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(certificatesCmd)
}
