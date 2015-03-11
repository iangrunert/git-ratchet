package main

import (
	ratchet "github.com/iangrunert/git-ratchet/cmd"
	"github.com/spf13/cobra"
	log "github.com/spf13/jwalterweatherman"
	"os"
)

func main() {
	var write bool
	var verbose bool

	var checkCmd = &cobra.Command{
		Use:   "check",
		Short: "Checks the values passed in against the most recent stored values.",
		Long: `Checks the values passed in against the most recent stored values. 
The most recent stored values are found by walking up the commit graph and looking at the git-notes stored.`,
		Run: func(cmd *cobra.Command, args []string) {
			if verbose {
				log.SetLogThreshold(log.LevelInfo)
				log.SetStdoutThreshold(log.LevelInfo)
			}

			err := ratchet.Check(write, os.Stdin)
			if err != 0 {
				os.Exit(err)
			}
		},
	}

	checkCmd.Flags().BoolVarP(&write, "write", "w", false, "write values if no increase is detected. only use on your CI server.")
	checkCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "increase logging verbosity.")

	var measure string
	var excuse string

	var excuseCmd = &cobra.Command{
		Use:   "excuse",
		Short: "Write an excuse for a measurement increase, so that the check command will ignore an increase.",
		Long:  `Write an excuse for a measurement increase. This will allow the check command to pass.`,
		Run: func(cmd *cobra.Command, args []string) {
			if verbose {
				log.SetLogThreshold(log.LevelInfo)
				log.SetStdoutThreshold(log.LevelInfo)
			}

			ratchet.Excuse(measure, excuse)
		},
	}

	excuseCmd.Flags().StringVarP(&measure, "name", "n", "", "names of the measures to excuse, comma separated list")
	excuseCmd.Flags().StringVarP(&excuse, "excuse", "e", "", "excuse for the measure rising")
	excuseCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "increase logging verbosity.")

	var dumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump a CSV file containing the measurement data over time.",
		Long:  `Dump a CSV file containing the measurement data over time.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := ratchet.Dump(os.Stdout)

			if err != 0 {
				os.Exit(err)
			}
		},
	}

	var rootCmd = &cobra.Command{Use: "git-ratchet"}
	rootCmd.AddCommand(checkCmd, excuseCmd, dumpCmd)
	rootCmd.Execute()
}
