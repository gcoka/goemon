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
		Use:   "goemon \"command to run\"",
		Short: "Filewatcher",
		Long:  `Filewatcher`,
		Args:  cobra.MinimumNArgs(1),
		// Uncomment the following line if your bare application
		// has an action associated with it:
		Run: func(cmd *cobra.Command, args []string) {
			opt.WatchInterval = viper.GetInt("delay")
			opt.Ext = viper.GetStringSlice("ext")

			fmt.Println(opt)
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

	cmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is ./goemon.yaml)")
	cmd.Flags().UintP("delay", "d", 200, "Delay")
	viper.BindPFlag("delay", cmd.Flags().Lookup("delay"))
	cmd.Flags().StringSliceP("ext", "e", []string{}, "specify extentions")
	viper.BindPFlag("ext", cmd.Flags().Lookup("ext"))

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

		viper.AddConfigPath(".")      // adding current directory as first search path
		viper.SetConfigName("goemon") // name of config file (without extension)
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
		fmt.Println(viper.AllSettings())
	}
}
