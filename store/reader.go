package store

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	log "github.com/spf13/jwalterweatherman"
	"io"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func ParseInputType(input string) InputType {
	switch input {
	case "csv":
		return CSV
	case "checkstyle":
		return Checkstyle
	default:
		return Unknown
	}
}

func CommitMeasureCommand(prefix string) *exec.Cmd {
	return GitLog("git-ratchet-1-"+prefix, "HEAD", `%H,%ae,%at,"%N",`)
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

			measures, err := ParseMeasures(strings.NewReader(strings.Trim(record[3], "\\\"")), CSV)
			if err != nil {
				return CommitMeasure{}, err
			}

			if len(measures) > 0 {
				return CommitMeasure{CommitHash: strings.Trim(record[0], "'"),
					Committer: record[1],
					Timestamp: time.Unix(int64(timestamp), 0),
					Measures:  measures}, nil
			}
		}
	}, nil
}

func ParseMeasures(r io.Reader, t InputType) ([]Measure, error) {
	switch t {
	case CSV:
		return ParseMeasuresCSV(r)
	case Checkstyle:
		return ParseMeasuresCheckstyle(r)
	default:
		return nil, errors.New("Unknown input type")
	}
}

func ParseMeasuresCSV(r io.Reader) ([]Measure, error) {
	data := csv.NewReader(r)
	data.FieldsPerRecord = -1 // Variable number of fields per record

	measures := make([]Measure, 0)

	for {
		var baseline int

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

		if len(arr) > 2 {
			baseline, err = strconv.Atoi(arr[2])
			if err != nil {
				return nil, err
			}
		} else {
			baseline = value
		}

		measure := Measure{Name: arr[0], Value: value, Baseline: baseline}
		measures = append(measures, measure)
	}

	sort.Sort(ByName(measures))

	return measures, nil
}

func ParseMeasuresCheckstyle(r io.Reader) ([]Measure, error) {
	decoder := xml.NewDecoder(r)
	errors := 0

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "error" {
				errors++
			}
		}
	}

	return []Measure{{Name: "errors", Value: errors, Baseline: errors}}, nil
}

func CompareMeasures(prefix string, hash string, storedm []Measure, computedm []Measure, slack int, zeroOnMissing bool) ([]Measure, error) {
	if len(storedm) == 0 {
		return computedm, errors.New("No stored measures to compare against.")
	}

	excuses, err := GetExclusions(prefix, hash)

	if err != nil {
		return computedm, err
	}

	log.INFO.Printf("Total excuses %s", excuses)

	failing := make([]*Measure, 0)
	zeroMes := make([]Measure, 0)

	i := 0
	j := 0

	exc := 0

	for i < len(storedm) && j < len(computedm) {
		stored := storedm[i]
		computed := computedm[j]

		log.INFO.Printf("Checking meaures: %s %s", stored.Name, computed.Name)
		if stored.Name < computed.Name {
			log.ERROR.Printf("Missing computed value for stored measure: %s", stored.Name)
			if zeroOnMissing {
				zeroMes = append(zeroMes, Measure{Name: stored.Name, Value: 0, Baseline: 0})
			} else {
				failing = append(failing, &stored)
			}
			i++
		} else if computed.Name < stored.Name {
			log.WARN.Printf("New measure found: %s", computed.Name)
			j++
		} else {
			if computed.Baseline > stored.Baseline {
				computed.Baseline = stored.Baseline
				computedm[j].Baseline = stored.Baseline
			}

			// Compare the value
			if computed.Value > (stored.Baseline + slack) {
				log.ERROR.Printf("Measure rising: %s, delta %d", computed.Name, (computed.Value - stored.Baseline))

				if exc < len(excuses) {
					ex := excuses[exc]

					log.INFO.Printf("Checking excuses: %s %s", ex, computed)
					if ex < computed.Name {
						log.WARN.Printf("Exclusion found for not failing measure: %s", ex)
						exc++
						failing = append(failing, &computed)
					} else if computed.Name < ex {
						log.ERROR.Printf("No exclusion for failing measure: %s", computed.Name)
						failing = append(failing, &computed)
					} else {
						log.WARN.Printf("Exclusion found for failing measure: %s", computed.Name)
						computed.Baseline = computed.Value
						computedm[j].Baseline = computed.Value
						exc++
					}
				} else {
					failing = append(failing, &computed)
				}

			}
			i++
			j++
		}
	}

	for i < len(storedm) {
		stored := storedm[i]
		log.ERROR.Printf("Missing computed value for stored measure: %s", stored.Name)
		if zeroOnMissing {
			zeroMes = append(zeroMes, Measure{Name: stored.Name, Value: 0, Baseline: 0})
		} else {
			failing = append(failing, &stored)
		}
		i++
	}

	for j < len(computedm) {
		computed := computedm[i]
		log.WARN.Printf("New measure found: %s", computed.Name)
		j++
	}

	if len(failing) > 0 {
		return computedm, errors.New("One or more metrics currently failing.")
	}

	computedm = append(computedm, zeroMes...)
	sort.Sort(ByName(computedm))

	return computedm, nil
}

func GetExclusions(prefix string, hash string) ([]string, error) {
	ref := "git-ratchet-excuse-1-" + prefix

	gitlog := GitLog(ref, hash+"^1..HEAD", "%N")

	stdout, err := gitlog.StdoutPipe()
	if err != nil {
		return []string{}, err
	}

	scanner := bufio.NewScanner(stdout)

	err = gitlog.Start()
	if err != nil {
		return []string{}, err
	}

	exclusions := make([]string, 0)

	for scanner.Scan() {
		record := strings.Trim(scanner.Text(), "'")

		if len(record) == 0 {
			continue
		}

		measures, err := ParseExclusion(record)

		if err != nil && err != io.EOF {
			return []string{}, err
		}

		exclusions = append(exclusions, measures...)
	}

	if err = scanner.Err(); err != nil {
		return []string{}, err
	}

	stdout.Close()

	err = gitlog.Wait()

	if err != nil && err != syscall.EPIPE {
		return []string{}, err
	}

	sort.Strings(exclusions)

	return exclusions, nil
}

func ParseExclusion(ex string) ([]string, error) {
	log.INFO.Printf("Exclusion %s", ex)

	var m Exclusion
	err := json.Unmarshal([]byte(strings.Trim(ex, "'")), &m)

	if err != nil {
		return []string{}, err
	}

	return m.Measure, nil
}
