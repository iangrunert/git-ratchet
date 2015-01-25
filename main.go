package main

import (
	   "github.com/spf13/cobra"
	   "./store"
	   "log"
)

func main() {
	var write bool

	log.SetFlags(0)

	var checkCmd = &cobra.Command{
		Use:   "check",
		Short: "Checks the values passed in against the most recent stored values.",
		Long:  `Checks the values passed in against the most recent stored values. 
The most recent stored values are found by walking up the commit graph and looking at the git-notes stored.`,
		Run: func(cmd *cobra.Command, args []string) {
			 readMeasure, err := store.CommitMeasures()
			 if err != nil {
			 	log.Fatal(err)
			 }
			 measure, err := readMeasure()
			 if err != nil {
			 	log.Fatal(err)
			 }
			 log.Print(measure)
		},
	}

	checkCmd.Flags().BoolVarP(&write, "write", "w", false, "write values if no increase is detected. only use on your CI server.")

	var excuseCmd = &cobra.Command{
		Use:   "excuse",
		Short: "Write an excuse for a measurement increase, so that the check command will ignore an increase.",
		Long:  `Write an excuse for a measurement increase. This will allow the check command to pass.`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	var dumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump a CSV file containing the measurement data over time.",
		Long:  `Dump a CSV file containing the measurement data over time.`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	var leaderboardCmd = &cobra.Command{
		Use:   "leaderboard",
		Short: "Show a sorted leaderboard of which developers contributed to metric decreases.",
		Long:  `Dump a CSV file containing the measurement data over time.
If multiple developers have committed between subsequent runs, they'll share the points 50/50.`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	
	var rootCmd = &cobra.Command{Use: "git-ratchet"}
    rootCmd.AddCommand(checkCmd, excuseCmd, dumpCmd, leaderboardCmd)
    rootCmd.Execute()
}
