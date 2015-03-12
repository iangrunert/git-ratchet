package store

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"sort"
	"strconv"

	log "github.com/spf13/jwalterweatherman"
)

func PutMeasures(prefix string, m []Measure) error {
	writef := func(tempfile io.Writer) error {
		err := WriteMeasures(m, tempfile)
		if err != nil {
			return err
		}
		return nil
	}

	return WriteNotes(writef, "git-ratchet-1-"+prefix)
}

func WriteMeasures(measures []Measure, w io.Writer) error {
	out := csv.NewWriter(w)
	sort.Sort(ByName(measures))
	for _, m := range measures {
		err := out.Write([]string{m.Name, strconv.Itoa(m.Value), strconv.Itoa(m.Baseline)})
		if err != nil {
			return err
		}
	}
	out.Flush()
	return nil
}

func WriteExclusion(prefix string, ex Exclusion) error {
	ref := "git-ratchet-excuse-1-" + prefix

	writef := func(tempfile io.Writer) error {
		b, err := json.Marshal(ex)

		if err != nil {
			return err
		}

		tempfile.Write(b)
		return nil
	}

	err := WriteNotes(writef, ref)

	if err != nil {
		return err
	}

	err = PushNotes(ref)

	if err != nil {
		log.ERROR.Printf("Error while pushing notes: %s", err)
	}

	return nil
}
