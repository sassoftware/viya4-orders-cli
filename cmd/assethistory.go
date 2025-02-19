// Copyright © 2020-2023, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"log"

	"github.com/sassoftware/viya4-orders-cli/lib/assetreqs"
	"github.com/spf13/cobra"
)

// assetHistoryCmd represents the assetHistory command
var assetHistoryCmd = &cobra.Command{
	Use:   "assetHistory [order number]",
	Short: "Get the list of completed asset downloads for the given order number",
	Example: "viya4-orders-cli assetHistory 993456\n" +
		"viya4-orders-cli ah 993456\n" +
		"viya4-orders-cli ah 993456 -p $HOME/sas -n ah_993456",
	Aliases: []string{"ah"},
	Args:    cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ar := assetreqs.New(token, "assetHistory", args[0], "", "", "", assetFilePath, assetFileName, outFormat, allowUnsuppd)
		err := ar.GetAsset()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(assetHistoryCmd)
}
