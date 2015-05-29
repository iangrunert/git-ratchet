package cmd

import (
	"bytes"
	"github.com/iangrunert/git-ratchet/store"
	log "github.com/spf13/jwalterweatherman"
	"io"
)

func Check(prefix string, slack int, write bool, input io.Reader) int {
	// Parse the measures from stdin
	log.INFO.Println("Parsing measures from stdin")
	passedMeasures, err := store.ParseMeasures(input)
	log.INFO.Println("Finished parsing measures from stdin")
	log.INFO.Println(passedMeasures)
	if err != nil {
		log.FATAL.Println(err)
		return 10
	}

	log.INFO.Println("Reading measures stored in git")
	gitlog := store.CommitMeasureCommand(prefix)
	var stderr bytes.Buffer
	gitlog.Stderr = &stderr

	readStoredMeasure, err := store.CommitMeasures(gitlog)
	if err != nil {
		log.FATAL.Println(err)
		return 20
	}

	commitmeasure, err := readStoredMeasure()

	// Empty state of the repository - no stored metrics. Let's store one if we can.
	if err == io.EOF {
		log.INFO.Println("No measures found.")
		if write {
			log.INFO.Println("Writing initial measure values.")
			err = store.PutMeasures(prefix, passedMeasures)
			if err != nil {
				log.FATAL.Println(err)
				return 30
			}
			log.INFO.Println("Successfully written initial measures.")
		}
	} else if err != nil {
		log.FATAL.Println(err)
		return 40
	} else {
		log.INFO.Println("Checking passed measure against stored value")
		finalMeasures, compareErr := store.CompareMeasures(prefix, commitmeasure.CommitHash, commitmeasure.Measures, passedMeasures, slack)

		if write {
			log.INFO.Println("Writing measure values.")
			err = store.PutMeasures(prefix, finalMeasures)
			if err != nil {
				log.FATAL.Println(err)
				return 30
			}
			log.INFO.Println("Successfully written measures.")
		}
		if compareErr != nil {
			log.FATAL.Println(compareErr)
			return 50
		} else {
			log.INFO.Println("Metrics passing!")
		}
	}

	log.INFO.Println("Finished reading measures stored in git")
	return 0
}
