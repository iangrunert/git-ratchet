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
	
	runCheck(t, true, "foo,5")
	runCommand(t, repo, exec.Command("git", "add", createFile(t, repo, "bar.txt").Name()))
	runCommand(t, repo, exec.Command("git", "commit", "-m", "Third Commit"))
	runCheck(t, true, "foo,4")
	
	dump := bufio.NewScanner(bytes.NewReader(runDump(t).Bytes()))
	
	dump.Scan()
	
	checkString(t, "foo,4", dump.Text())
	
	dump.Scan()

	checkString(t, "foo,5", dump.Text())
}

func checkString(t *testing.T, expected string, actual string) {
	if !strings.HasSuffix(actual, expected) {
		t.Fatalf("Dump incorrect. Expected suffix %s got %s", expected, actual)
	}	
}

func runDump(t *testing.T) *bytes.Buffer {
	t.Logf("Running dump command")

	buf := new(bytes.Buffer)

	errCode := Dump(buf)

	if errCode != 0 {
		t.Fatalf("Check command failed! Error code: %d", errCode)
	}
	
	return buf
}
