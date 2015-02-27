package cmd

import (
	"github.com/iangrunert/git-ratchet/store"
	"encoding/csv"
	log "github.com/spf13/jwalterweatherman"
	"io"
	"os"
	"strconv"
)

func Dump() {
	log.INFO.Println("Reading measures stored in git")
	gitlog := store.CommitMeasureCommand()
	
	readStoredMeasure, err := store.CommitMeasures(gitlog)
	if err != nil {
		log.FATAL.Println(err)
		os.Exit(20)
	}
	
	for {
		cm, err := readStoredMeasure()
		
		// Empty state of the repository - no stored metrics.
		if err == io.EOF {
			break
		} else if err != nil {
			log.FATAL.Println(err)
			os.Exit(40)
		}
		
		out := csv.NewWriter(os.Stdout)
		
		for _, measure := range cm.Measures {
			out.Write([]string{cm.Timestamp.String(), measure.Name, strconv.Itoa(measure.Value)})
		}
		out.Flush()
	}
	
	err = gitlog.Wait()
	
	if err != nil {
		log.FATAL.Println(err)
		os.Exit(22)
	}
	
	log.INFO.Println("Finished reading measures stored in git")
}
