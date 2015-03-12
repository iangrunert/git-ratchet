package cmd

import (
	"testing"
	"os/exec"
	"bytes"
	"bufio"
	"strings"

	log "github.com/spf13/jwalterweatherman"
)

func TestDump(t *testing.T) {
	if (testing.Verbose()) {
		log.SetLogThreshold(log.LevelInfo)
		log.SetStdoutThreshold(log.LevelInfo)
	}
	
	repo := createEmptyGitRepo(t)
	
	runCheckP(t, "foo", true, "foo,5")
	runCommand(t, repo, exec.Command("git", "add", createFile(t, repo, "bar.txt").Name()))
	runCommand(t, repo, exec.Command("git", "commit", "-m", "Third Commit"))
	runCheckP(t, "foo", true, "foo,4")
	
	dump := bufio.NewScanner(bytes.NewReader(runDump(t, "foo").Bytes()))
	
	dump.Scan()
	
	checkString(t, "foo,4", dump.Text())
	
	dump.Scan()

	checkString(t, "foo,5", dump.Text())

	if len(runDump(t, "bar").Bytes()) > 0 {
		t.Fatalf("Should be no data under prefix bar")
	}
}

func checkString(t *testing.T, expected string, actual string) {
	if !strings.HasSuffix(actual, expected) {
		t.Fatalf("Dump incorrect. Expected suffix %s got %s", expected, actual)
	}	
}

func runDump(t *testing.T, prefix string) *bytes.Buffer {
	t.Logf("Running dump command")

	buf := new(bytes.Buffer)

	errCode := Dump(prefix, buf)

	if errCode != 0 {
		t.Fatalf("Dump command failed! Error code: %d", errCode)
	}
	
	return buf
}
