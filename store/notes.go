package store

import (
	"encoding/csv"
	"errors"
	log "github.com/spf13/jwalterweatherman"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Measure struct {
	Name  string
	Value int
}

type CommitMeasure struct {
	CommitHash string
	Timestamp  time.Time
	Committer  string
	Measures   []Measure
}

func (cm *CommitMeasure) String() string {
	return cm.CommitHash
}

func CommitMeasureCommand() *exec.Cmd {
	gitlog := exec.Command("git", "log", "--show-notes=git-ratchet", `--pretty=format:'%H,%an <%ae>,%at,"%N",'`)
	log.INFO.Println(strings.Join(gitlog.Args, " "))
	return gitlog
}

func CommitMeasures(gitlog *exec.Cmd) (func() (CommitMeasure, error), error) {
	stdout, err := gitlog.StdoutPipe()
	if err != nil {
		return nil, err
	}

	output := csv.NewReader(stdout)
	output.TrailingComma = true

	err = gitlog.Start()
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
			if len(record[len(record)-1]) == 0 {
				continue
			}

			timestamp, err := strconv.Atoi(strings.Trim(record[2], "\\\""))
			if err != nil {
				return CommitMeasure{}, err
			}

			measures, err := ParseMeasures(strings.NewReader(strings.Trim(record[3], "\\\"")))
			if err != nil {
				return CommitMeasure{}, err
			}

			if len(measures) > 0 {
				return CommitMeasure{CommitHash: record[0],
					Committer: record[1],
					Timestamp: time.Unix(int64(timestamp), 0),
					Measures:  measures}, nil
			}
		}
	}, nil
}

func PutMeasures(m []Measure) error {
	// Create a temporary file
	notepath := ".git-ratchet-note"

	tempfile, err := os.Create(notepath)
	if err != nil {
		return err
	}
	defer os.Remove(notepath)

	err = WriteMeasures(m, tempfile)
	if err != nil {
		return err
	}

	err = tempfile.Close()
	if err != nil {
		return err
	}

	writenotes := exec.Command("git", "notes", "--ref=git-ratchet", "add", "-f", "-F", notepath)

	log.INFO.Println(strings.Join(writenotes.Args, " "))

	return writenotes.Run()
}

type ByName []Measure

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

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

func ParseMeasures(r io.Reader) ([]Measure, error) {
	data := csv.NewReader(r)
	data.FieldsPerRecord = -1 // Variable number of fields per record

	measures := make([]Measure, 0)

	for {
		arr, err := data.Read()
		if err == io.EOF {
			break
		}

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

		measure := Measure{Name: arr[0], Value: value}
		measures = append(measures, measure)
	}

	sort.Sort(ByName(measures))

	return measures, nil
}

func CompareMeasures(storedm []Measure, computedm []Measure) error {
	if len(computedm) == 0 {
		return errors.New("No measures passed to git-ratchet to compare against.")
	}

	if len(storedm) == 0 {
		return errors.New("No stored measures to compare against.")
	}

	failing := false

	i := 0
	j := 0

	for i < len(storedm) && j < len(computedm) {
		stored := storedm[i]
		computed := computedm[j]
		log.INFO.Printf("Checking meaures: %s %s", stored.Name, computed.Name)
		if stored.Name < computed.Name {
			log.ERROR.Printf("Missing computed value for stored measure: %s", stored.Name)
			failing = true
			i++
		} else if computed.Name < stored.Name {
			log.WARN.Printf("New measure found: %s", computed.Name)
			j++
		} else {
			// Compare the value
			if computed.Value > stored.Value {
				log.ERROR.Printf("Measure rising: %s", computed.Name)
				failing = true
			}
			i++
			j++
		}
	}

	for i < len(storedm) {
		stored := storedm[i]
		log.ERROR.Printf("Missing computed value for stored measure: %s", stored.Name)
		failing = true
		i++
	}

	for j < len(computedm) {
		computed := computedm[i]
		log.WARN.Printf("New measure found: %s", computed.Name)
		j++
	}

	if failing {
		return errors.New("One or more metrics currently failing.")
	}

	return nil
}
