// Copyright © 2020, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sassoftware/viya4-orders-cli/lib/authn"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

var assetFileName string
var assetFilePath string
var cfgFile string
var outFormat string
var token string

// Version is set by the build.
var version string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Version: version,
	Use:   "viya4-orders-cli",
	Short: "SAS Viya Orders CLI is a CLI to the SAS Viya Orders API",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Authentication is required for all commands.
	cobra.OnInitialize(initConfig, auth)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		"config file (default is $HOME/.viya4-orders-cli)")
	rootCmd.PersistentFlags().StringVarP(&assetFileName, "file-name", "n", "",
		"name of the file where you want the downloaded order asset stored\n"+
			"(defaults:\n\tcerts - SASiyaV4_<order number>_certs.zip\n\tlicense and depassets - SASiyaV4_<order number>_<renewal sequence>_<cadence information>_<asset name>_<date time stamp>."+
			"<asset extension>\n)")
	rootCmd.PersistentFlags().StringVarP(&assetFilePath, "file-path", "p", "",
		"path to where you want the downloaded order asset stored (default is path to your current working directory)")
	rootCmd.PersistentFlags().StringVarP(&outFormat, "output", "o", "text",
		"output format - valid values:\n"+
			"\tj, json\n\tt, text\n")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalln("ERROR: homedir.Dir() returned: " + err.Error())
		}

		// Search config in home directory with name ".viya4-orders-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".viya4-orders-cli")
		// If they provide a config file with no extension if must be in yaml format.
		viper.SetConfigType("yaml")
	}

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalln("ERROR: problem parsing config file " + viper.ConfigFileUsed() + ": " + err.Error())
		}
	}

	if outFormat != "j" && outFormat != "json" {
		if viper.ConfigFileUsed() != "" {
			log.Println("INFO: using config file:", viper.ConfigFileUsed())
		} else {
			log.Println("INFO: no config file found")
		}
	}

	// Read in environment vars.
	viper.AutomaticEnv()

	// Bind flags from the command line to the Viper framework.
	err = viper.BindPFlags(rootCmd.Flags())
	if err != nil {
		log.Fatalln("ERROR: viper.BindPFlags() returned: " + err.Error())
	}

	setOptions()
}

// Get option values from Viper and validate where appropriate. In general, those options set on the command line override those set
// in the environment which override those set in the config.
func setOptions() {
	assetFileName = viper.GetString("file-name")
	assetFilePath = viper.GetString("file-path")
	if assetFilePath != "" {
		// Make sure the given path exists and is a directory.
		if chk, err := os.Stat(assetFilePath); err == nil {
			// It exists, but is it a directory?
			if !chk.Mode().IsDir() {
				usageError(assetFilePath + " is not a directory and therefore is not a valid value for -p, --file-path!")
			}
		} else if os.IsNotExist(err) {
			// path/to/whatever does *not* exist
			usageError(assetFilePath + " does not exist and therefore is not a valid value for -p, --file-path!")
		}
	}

	outFormat := viper.GetString("output")
	// Validate output flag value.
	if outFormat != "text" && outFormat != "t" && outFormat != "json" && outFormat != "j" {
		usageError("invalid value " + outFormat + " specified for -o, --output option!")
	}
}

func usageError(message string) {
	log.Fatalln("Error: " + message)
	rootCmd.Help()
}

// Get Bearer token.
func auth() {
	var err error
	token, err = authn.GetBearerToken()
	if err != nil {
		log.Fatalln(err.Error())
	}
}
