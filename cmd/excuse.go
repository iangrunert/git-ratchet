package cmd

import (
	"github.com/iangrunert/git-ratchet/store"
	log "github.com/spf13/jwalterweatherman"
	"strings"
)

func Excuse(prefix string, measure string, excuse string) int {
	name, err := store.GetCommitterName()
	
	if err != nil {
		log.FATAL.Println("Error when fetching committer name")
		log.DEBUG.Println(err)
		return 10
	}
	
	exclusion := store.Exclusion{Committer: name, Excuse: excuse, Measure: strings.Split(measure, ",")}
	
	err = store.WriteExclusion(prefix, exclusion)
	
	if err != nil {
		log.FATAL.Println("Error writing exclusion note %s", err)
		return 20
	}

	return 0
}
