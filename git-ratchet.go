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
	var prefix string
	var slack int
	var inputType string

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

			err := ratchet.Check(prefix, slack, write, inputType, os.Stdin)
			if err != 0 {
				os.Exit(err)
			}
		},
	}

	checkCmd.Flags().BoolVarP(&write, "write", "w", false, "write values if no increase is detected. only use on your CI server.")
	checkCmd.Flags().IntVarP(&slack, "slack", "s", 0, "slack value, increase within the range of the slack is acceptable.")
	checkCmd.Flags().StringVarP(&inputType, "inputType", "i", "csv", "input type. csv and checkstyle available.")

	var measure string
	var excuse string

	var excuseCmd = &cobra.Command{
		Use:   "excuse",
		Short: "Write an excuse for a measurement increase.",
		Long:  `Write an excuse for a measurement increase. This will allow the check command to pass.`,
		Run: func(cmd *cobra.Command, args []string) {
			if verbose {
				log.SetLogThreshold(log.LevelInfo)
				log.SetStdoutThreshold(log.LevelInfo)
			}

			os.Exit(ratchet.Excuse(prefix, measure, excuse))
		},
	}

	excuseCmd.Flags().StringVarP(&measure, "name", "n", "", "names of the measures to excuse, comma separated list.")
	excuseCmd.Flags().StringVarP(&excuse, "excuse", "e", "", "excuse for the measure rising.")

	var dumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump a CSV file containing the measurement data over time.",
		Long:  `Dump a CSV file containing the measurement data over time.`,
		Run: func(cmd *cobra.Command, args []string) {
			if verbose {
				log.SetLogThreshold(log.LevelInfo)
				log.SetStdoutThreshold(log.LevelInfo)
			}

			err := ratchet.Dump(prefix, os.Stdout)

			if err != 0 {
				os.Exit(err)
			}
		},
	}

	var rootCmd = &cobra.Command{Use: "git-ratchet"}
	rootCmd.AddCommand(checkCmd, excuseCmd, dumpCmd)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "increase logging verbosity.")
	rootCmd.PersistentFlags().StringVarP(&prefix, "prefix", "p", "master", "prefix the ratchet notes. useful for storing multiple sets of values in the same repo.")

	rootCmd.Execute()
}
