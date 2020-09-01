//Copyright © 2020, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
//SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sassoftware/viya4-orders-cli/lib/authen"
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

//rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: "1.1",
	Use:   "viya4-orders-cli",
	Short: "viya4-orders-cli is a CLI to the SAS Viya Orders API",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Panic(err.Error())
	}
}

func init() {
	//Authentication is required for all commands.
	cobra.OnInitialize(initConfig, auth)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		"config file (default is $HOME/.viya4-orders-cli)")
	//All flags are global because they all apply to all commands and it seems silly to repeat them in
	//each commmand's source.
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
		//Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		//Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Panic(err.Error())
		}

		//Search config in home directory with name ".viya4-orders-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".viya4-orders-cli")
		//If they provide a config file with no extension if must be in yaml format
		viper.SetConfigType("yaml")
	}

	//If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Panic(err.Error())
	}

	//Read in environment vars
	viper.AutomaticEnv()

	//Bind flags from the command line to the Viper framework
	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		log.Panic(err)
	}

	setOptions()
}

//Get option values from Viper and validate where appropriate. In general, those options set on the command line override those set
//in the environment which override those set in the config.
func setOptions() {
	assetFileName = viper.GetString("file-name")
	assetFilePath = viper.GetString("file-path")
	if assetFilePath != "" {
		//Make sure the given path exists and is a directory
		if chk, err := os.Stat(assetFilePath); err == nil {
			//It exists, but is it a directory?
			if !chk.Mode().IsDir() {
				usageError(assetFilePath + " is not a directory and therefore is not a valid value for -p, --file-path!")
			}
		} else if os.IsNotExist(err) {
			// path/to/whatever does *not* exist
			usageError(assetFilePath + " does not exist and therefore is not a valid value for -p, --file-path!")
		}
	}

	outFormat := viper.GetString("output")
	//Validate output flag value.
	if outFormat != "text" && outFormat != "t" && outFormat != "json" && outFormat != "j" {
		usageError("Invalid value " + outFormat + " specified for -o, --output option!")
	}
}

func usageError(message string) {
	rootCmd.Help()
	fmt.Println("Error: " + message)
	os.Exit(1)
}

//Get Bearer token.
func auth() {
	token = authen.GetBearerToken()
}
