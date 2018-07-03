package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gcoka/goemon/goemon"
)

var cfgFile string

// NewCmdRoot initialize the root command
func NewCmdRoot() *cobra.Command {
	opt := &goemon.Option{}
	opt.Default()

	cmd := &cobra.Command{
		Use:   "goemon",
		Short: "Filewatcher",
		Long:  `Filewatcher`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		Run: func(cmd *cobra.Command, args []string) {
			g := goemon.New(args, opt)
			done := make(chan error)
			go func() {
				err := g.Start()
				if err != nil {
					log.Println(err)
				}
				close(done)
			}()

			sig := make(chan os.Signal)
			signal.Notify(sig, os.Interrupt, os.Kill)

			select {
			case <-sig:
			case <-done:
			}

			g.Close()
		},
	}
	cobra.OnInitialize(initConfig)

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./goemon.yaml)")
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	cmd.Flags().StringSliceVarP(opt.Ext, "ext", "e", []string{}, "specify extentions")

	return cmd
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cmd := NewCmdRoot()
	if err := cmd.Execute(); err != nil {
		cmd.SetOutput(os.Stderr)
		cmd.Println(err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	} else {

		viper.SetConfigName("goemon") // name of config file (without extension)
		viper.AddConfigPath(".")      // adding current directory as first search path
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
