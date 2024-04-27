// Copyright © 2020-2023, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"log"
	"sync"

	"github.com/sassoftware/viya4-orders-cli/lib/assetreqs"
	"github.com/spf13/cobra"
)

// getallCmd represents the getall command
var getallCmd = &cobra.Command{
	Use:   "getall [order number] [cadence name] [cadence version]",
	Short: "Download all downloadable objects (assets + license + certs) for the given order number at the given cadence name and version	",
	Example: "viya4-orders-cli getall 993456 stable 2020.0.3\n" +
		"viya4-orders-cli getall 993456 stable 2020.0.3 -p $HOME/sas -n license_993456_stable_2020.0.3",
	Aliases: []string{"all"},
	Args:    cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {

		var wg sync.WaitGroup

		wg.Add(3)

		go func() {
			ar := assetreqs.New(token, "license", args[0], args[1], args[2], assetFilePath, "", outFormat, allowUnsuppd)
			err := ar.GetAsset()
			if err != nil {
				log.Fatalln(err)
			}
			wg.Done()
		}()

		go func() {
			ar := assetreqs.New(token, "deploymentAssets", args[0], args[1], args[2], assetFilePath, "", outFormat, allowUnsuppd)
			err := ar.GetAsset()
			if err != nil {
				log.Fatalln(err)
			}
			wg.Done()
		}()

		go func() {
			ar := assetreqs.New(token, "certificates", args[0], "", "", assetFilePath, "", outFormat, allowUnsuppd)
			err := ar.GetAsset()
			if err != nil {
				log.Fatalln(err)
			}
			wg.Done()
		}()

		wg.Wait()

	},
}

func init() {
	rootCmd.AddCommand(getallCmd)
}
