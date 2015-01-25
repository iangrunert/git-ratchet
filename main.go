package main

import (
	   "github.com/spf13/cobra"
	   log "github.com/spf13/jwalterweatherman"
	   "io"
	   "./store"
	   "os"
)

func main() {
	var write bool
	var verbose bool

	var checkCmd = &cobra.Command{
		Use:   "check",
		Short: "Checks the values passed in against the most recent stored values.",
		Long:  `Checks the values passed in against the most recent stored values. 
The most recent stored values are found by walking up the commit graph and looking at the git-notes stored.`,
		Run: func(cmd *cobra.Command, args []string) {
			 if verbose {
			 	log.SetLogThreshold(log.LevelInfo)
			 	log.SetStdoutThreshold(log.LevelInfo)
			}
			 // Parse the measures from stdin
			 log.INFO.Println("Parsing measures from stdin")
			 passedMeasures, err := store.ParseMeasures(os.Stdin)
			 log.INFO.Println("Finished parsing measures from stdin")
			 log.INFO.Println(passedMeasures)
			 if err != nil {
			 	log.FATAL.Println(err)
				os.Exit(10)
			 }
			 
			 log.INFO.Println("Reading measures stored in git")
			 gitlog := store.CommitMeasureCommand()

			 log.INFO.Println(gitlog.Args)

			 readStoredMeasure, err := store.CommitMeasures(gitlog)
			 if err != nil {
			 	log.FATAL.Println(err)
				os.Exit(20)
			 }

			 measure, err := readStoredMeasure()

			 // Empty state of the repository - no stored metrics. Let's store one if we can.
			 if err == io.EOF {
			 	log.INFO.Println("No measures found.")
			 	if write {
				   log.INFO.Println("Writing initial measure values.")
				   err = store.PutMeasures(passedMeasures)
				   if err != nil {
				   	  log.FATAL.Println(err)
					  os.Exit(30)
				   }
				   log.INFO.Println("Successfully written initial measures.")
				}
			 } else if err != nil {	 	
			 	log.FATAL.Println(err)
				os.Exit(40)
			 } else {
			    log.INFO.Println(measure)
			 }
			 
			 err = gitlog.Wait()

			 if err != nil {
			 	log.FATAL.Println(err)
				os.Exit(22)
			 }

			 log.INFO.Println("Finished reading measures stored in git")
		},
	}

	checkCmd.Flags().BoolVarP(&write, "write", "w", false, "write values if no increase is detected. only use on your CI server.")
	checkCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "logging verbosity.")

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
