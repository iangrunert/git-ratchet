package store

import (
	"encoding/csv"
	"encoding/json"
	"bufio"
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

func GitLog(ref string, commitrange string, format string) *exec.Cmd {
	gitlog := exec.Command("git", "log", "--show-notes=" + ref, `--pretty=format:'` + format + `'`, commitrange)
	log.INFO.Println(strings.Join(gitlog.Args, " "))
	return gitlog
}

func CommitMeasureCommand() *exec.Cmd {
	return GitLog("git-ratchet", "HEAD", `%H,%an <%ae>,%at,"%N",`)
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
				return CommitMeasure{CommitHash: strings.Trim(record[0], "'"),
					Committer: record[1],
					Timestamp: time.Unix(int64(timestamp), 0),
					Measures:  measures}, nil
			}
		}
	}, nil
}

func WriteNotes(writef func(io.Writer) error, ref string) error {
	// Create a temporary file
	notepath := ".git-ratchet-note"

	tempfile, err := os.Create(notepath)
	if err != nil {
		return err
	}
	defer os.Remove(notepath)

	err = writef(tempfile)
	if err != nil {
		return err
	}

	err = tempfile.Close()
	if err != nil {
		return err
	}

	writenotes := exec.Command("git", "notes", "--ref=" + ref, "add", "-f", "-F", notepath)

	log.INFO.Println(strings.Join(writenotes.Args, " "))

	err = writenotes.Run()
	if err != nil {
		return err
	}
	
	pushnotes := exec.Command("git", "push", "origin", "refs/notes/" + ref)

	log.INFO.Println(strings.Join(pushnotes.Args, " "))

	return pushnotes.Run()
}

func PutMeasures(m []Measure) error {
	writef := func(tempfile io.Writer) error { 
		err := WriteMeasures(m, tempfile)
		if err != nil {
			return err
		}
		return nil
	}

	return WriteNotes(writef, "git-ratchet")
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

func CompareMeasures(hash string, storedm []Measure, computedm []Measure) error {
	if len(computedm) == 0 {
		return errors.New("No measures passed to git-ratchet to compare against.")
	}

	if len(storedm) == 0 {
		return errors.New("No stored measures to compare against.")
	}

	failing := make([]string, 0)

	i := 0
	j := 0

	for i < len(storedm) && j < len(computedm) {
		stored := storedm[i]
		computed := computedm[j]
		log.INFO.Printf("Checking meaures: %s %s", stored.Name, computed.Name)
		if stored.Name < computed.Name {
			log.ERROR.Printf("Missing computed value for stored measure: %s", stored.Name)
			failing = append(failing, stored.Name)
			i++
		} else if computed.Name < stored.Name {
			log.WARN.Printf("New measure found: %s", computed.Name)
			j++
		} else {
			// Compare the value
			if computed.Value > stored.Value {
				log.ERROR.Printf("Measure rising: %s", computed.Name)
				failing = append(failing, computed.Name)
			}
			i++
			j++
		}
	}

	for i < len(storedm) {
		stored := storedm[i]
		log.ERROR.Printf("Missing computed value for stored measure: %s", stored.Name)
		failing = append(failing, stored.Name)
		i++
	}

	for j < len(computedm) {
		computed := computedm[i]
		log.WARN.Printf("New measure found: %s", computed.Name)
		j++
	}

	if len(failing) > 0 {
		log.INFO.Printf("Checking for excuses")

		exclusions, err := GetExclusions(hash)

		if err != nil {
			return err
		}
	
		log.INFO.Printf("Total excuses %s", exclusions)
	
		i = 0
		j = 0
	
		missingexclusion := false
	
		for i < len(exclusions) && j < len(failing) {
			ex := exclusions[i]
			fail := failing[j]
			log.INFO.Printf("Checking excuses: %s %s", ex, fail)
			if ex < fail {
				log.WARN.Printf("Exclusion found for not failing measure: %s", ex)
				i++
			} else if fail < ex {
				log.ERROR.Printf("No exclusion for failing measure: %s", fail)
				missingexclusion = true
				j++
			} else {
				i++
				j++
			}		
		}
		
		if missingexclusion || j < len(failing) {
			return errors.New("One or more metrics currently failing.")
		}
	}

	return nil
}

type Exclusion struct {
	Committer string
	Excuse string
	Measure []string
}

func GetExclusions(hash string) ([]string, error) {
	gitlog := GitLog("git-ratchet-excuse", hash + "^1..HEAD", "%N")
	
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
			break
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
	
	err = gitlog.Wait()
	
	if err != nil {
		return []string{}, err
	}
	
	sort.Strings(exclusions)
	
	return exclusions, nil
}

func ParseExclusion(ex string) ([]string, error){
	log.INFO.Printf("Exclusion %s", ex)

	var m Exclusion
	err := json.Unmarshal([]byte(strings.Trim(ex, "'")), &m)
	
	if err != nil {
		return []string{}, err
	}
	
	return m.Measure, nil
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

func GetCommitterName() (string, error) {
	getname := exec.Command("git", "config", "--get", "user.name")
	
	name, err := getname.Output()

	if err != nil {
		log.ERROR.Printf("fucked %s", err)
		return "", err
	}

	return strings.Trim(string(name), "\n"), nil
}
