package store

import (
	"fmt"
	log "github.com/spf13/jwalterweatherman"
	"io"
	"os"
	"os/exec"
	"strings"
)

func GitLog(ref string, commitrange string, format string) *exec.Cmd {
	gitlog := exec.Command("git", "--no-pager", "log", "--show-notes="+ref, `--pretty=format:'`+format+`'`, commitrange)
	log.INFO.Println(strings.Join(gitlog.Args, " "))
	return gitlog
}

func GetCommitterName() (string, error) {
	getname := exec.Command("git", "config", "--get", "user.name")

	name, err := getname.CombinedOutput()

	if err != nil {
		log.ERROR.Printf("Get committer name failed %s : %s", err, name)
		return "", err
	}

	return strings.Trim(string(name), "\n"), nil
}

func WriteNotes(writef func(io.Writer) error, ref string) error {
	// Create a temporary file to store the note data
	notepath := ".git-ratchet-note"

	tempfile, err := os.Create(notepath)
	if err != nil {
		return fmt.Errorf("Error creating file .git-ratchet-note %s", err)
	}
	defer os.Remove(notepath)

	err = writef(tempfile)
	if err != nil {
		return fmt.Errorf("Error writing notes to .git-ratchet-note %s", err)
	}

	err = tempfile.Close()
	if err != nil {
		return fmt.Errorf("Error closing .git-ratchet-note %s", err)
	}
	
	writenotes := exec.Command("git", "notes", "--ref="+ref, "add", "-f", "-F", notepath)

	log.INFO.Println(strings.Join(writenotes.Args, " "))

	bytes, err := writenotes.CombinedOutput()

	if err != nil {
		return fmt.Errorf("Error writing notes %s, %s", err, string(bytes))
	}

	return err
}

func PushNotes(ref string) error {
	pushnotes := exec.Command("git", "push", "origin", "refs/notes/"+ref)
	log.INFO.Println(strings.Join(pushnotes.Args, " "))

	bytes, err := pushnotes.CombinedOutput()

	if err != nil {
		return fmt.Errorf("Error pushing notes %s, %s", err, bytes)
	}

	return err
}
