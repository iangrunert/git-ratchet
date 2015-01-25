package store

import (
	   "encoding/csv"
	   "errors"
	   "io"
	   "os/exec"
	   "strconv"
	   "strings"
	   "time"
)

type Measure struct {
	 Name string
	 Value int
	 Rising bool
}

type CommitMeasure struct {
	 CommitHash string
	 Timestamp time.Time
	 Committer string
	 Measures []Measure
}

func (cm *CommitMeasure) String() string {
	 return cm.CommitHash
}

func CommitMeasures() (func() (CommitMeasure, error), error) {
	log := exec.Command(`git`, `log --show-notes=git-ratchet --pretty=format:"%H,%an <%ae>,%at,%N"`)
	stdout, err := log.StdoutPipe()
	if err != nil {
	   return nil, err
	}
	output := csv.NewReader(stdout)
	err = log.Start()
	if err != nil {
	   return nil, err
	}

	return func() (CommitMeasure, error) {
		for {
			// The log is of the form commithash,committer,timestamp,note
			// If note is empty, there's no set of Measures
			record, err := output.Read()
			if err != nil {
			   return CommitMeasure{}, err
			}
			// The note needs to be non-empty to contain measures.
			if len(record[len(record) - 1]) == 0 {
			   continue
			}
			timestamp, err := strconv.Atoi(record[1])
			if err != nil {
			   return CommitMeasure{}, err
			}
			measures, err := ParseMeasures(strings.NewReader(record[3]))
			if err != nil {
			   return CommitMeasure{}, err
			}
			return CommitMeasure{CommitHash: record[0], 
				   				  Timestamp: time.Unix(int64(timestamp), 0),
								  Committer: record[2],
								  Measures: measures}, nil
		}
	}, nil
}

func ParseMeasures(r io.Reader) ([]Measure, error) {
	 data := csv.NewReader(r)
	 
	 measures := make([]Measure, 1)

	 for {
	 	 arr, err := data.Read()
		 if err != nil {
		 	return nil, err
		 }
		 if len(arr) < 2 {
		 	return nil, errors.New("Badly formatted measures")
		 }
		 value, err := strconv.Atoi(arr[1])
		 if err != nil {
		 	return nil, err
		 }
		 measure := Measure{Name: arr[0], Value: value, Rising: (len(arr) > 2)}
		 measures = append(measures, measure)
	 }
	 
	 return measures, nil
}