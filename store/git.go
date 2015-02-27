package store

import (
	log "github.com/spf13/jwalterweatherman"
	"io"
	"os"
	"os/exec"
	"strings"
)

func GitLog(ref string, commitrange string, format string) *exec.Cmd {
	gitlog := exec.Command("git", "log", "--show-notes=" + ref, `--pretty=format:'` + format + `'`, commitrange)
	log.INFO.Println(strings.Join(gitlog.Args, " "))
	return gitlog
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

func WriteNotes(writef func(io.Writer) error, ref string) error {
	// Create a temporary file to store the note data
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
