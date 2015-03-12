package cmd

import (
	"github.com/iangrunert/git-ratchet/store"
	log "github.com/spf13/jwalterweatherman"
	"os"
	"strings"
)

func Excuse(prefix string, measure string, excuse string) {
	name, err := store.GetCommitterName()
	
	if err != nil {
		log.FATAL.Println("Error when fetching committer name")
		log.DEBUG.Println(err)
		os.Exit(10)
	}
	
	exclusion := store.Exclusion{Committer: name, Excuse: excuse, Measure: strings.Split(measure, ",")}
	
	err = store.WriteExclusion(exclusion)
	
	if err != nil {
		log.FATAL.Println("Error writing exclusion note.")
		log.DEBUG.Println(err)
		os.Exit(20)
	}
}
