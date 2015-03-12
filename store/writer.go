package store

import (
	"encoding/json"
	"encoding/csv"
	"io"
	"sort"
	"strconv"
)

func PutMeasures(prefix string, m []Measure) error {
	writef := func(tempfile io.Writer) error { 
		err := WriteMeasures(m, tempfile)
		if err != nil {
			return err
		}
		return nil
	}

	return WriteNotes(writef, "git-ratchet-1-" + prefix)
}

func WriteMeasures(measures []Measure, w io.Writer) error {
	out := csv.NewWriter(w)
	sort.Sort(ByName(measures))
	for _, m := range measures {
		err := out.Write([]string{m.Name, strconv.Itoa(m.Value)})
		if err != nil {
			return err
		}
	}
	out.Flush()
	return nil
}

func WriteExclusion(ex Exclusion) error {
	writef := func(tempfile io.Writer) error { 
		b, err := json.Marshal(ex)
	
		if err != nil {
			return err
		}
		
		tempfile.Write(b)
		return nil
	}

	return WriteNotes(writef, "git-ratchet-excuse")
}
