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

		// check if "file-name" flag is used and if so, log a message this parameter is ignored with the getall command
		//if cmd.Flags().Lookup("file-name").Changed {
		if assetFileName != "" {
			log.Println("The getall command ignores the --file-name option. Default file names will be used instead.")
		}

		var wg sync.WaitGroup
		var assetTypes []string = []string{"license", "deploymentAssets", "certificates"}

		for _, v := range assetTypes {

			wg.Add(1)

			// cannot reference "v" directly inside func see https://stackoverflow.com/questions/39207608/how-does-golang-share-variables-between-goroutines
			// even though it was supposed to work fine in Go 1.22 https://go.dev/blog/go1.22
			go func(assetType string) {

				defer func() {
					recover()
					wg.Done()
				}()

				var p2, p3 string

				switch assetType {
				case "certificates":
					p2, p3 = "", ""
				default:
					p2, p3 = args[1], args[2]
				}

				ar := assetreqs.New(token, assetType, args[0], p2, p3, assetFilePath, "", outFormat, allowUnsuppd)
				err := ar.GetAsset()
				if err != nil {
					log.Panicln("Error ocurred while getting asset type", assetType, "error is:", err)
				}
			}(v)

		}

		log.Println("Waiting for downloads to complete.")
		wg.Wait()
		log.Println("Downloads are complete.")

	},
}

func init() {
	rootCmd.AddCommand(getallCmd)
}
