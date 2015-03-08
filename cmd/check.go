package cmd

import (
	"github.com/iangrunert/git-ratchet/store"
	log "github.com/spf13/jwalterweatherman"
	"io"
	"os"
)

func Check(write bool) {
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
	
	readStoredMeasure, err := store.CommitMeasures(gitlog)
	if err != nil {
		log.FATAL.Println(err)
		os.Exit(20)
	}

	commitmeasure, err := readStoredMeasure()
	
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
		log.INFO.Println(commitmeasure.Measures)
		log.INFO.Println("Checking passed measure against stored value")
		err = store.CompareMeasures(commitmeasure.CommitHash, commitmeasure.Measures, passedMeasures)
		if err != nil {
			log.FATAL.Println(err)
			os.Exit(50)
		} else if write {
			log.INFO.Println("Writing measure values.")
			err = store.PutMeasures(passedMeasures)
			if err != nil {
				log.FATAL.Println(err)
				os.Exit(30)
			}
			log.INFO.Println("Successfully written measures.")
		} else {
			log.INFO.Println("Metrics passing!")
		}
	}
	
	err = gitlog.Wait()
	
	if err != nil {
		log.FATAL.Println(err)
		os.Exit(22)
	}
	
	log.INFO.Println("Finished reading measures stored in git")
}