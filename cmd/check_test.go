package cmd

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	log "github.com/spf13/jwalterweatherman"
)

var checkStyleFile *os.File
var checkStyleFileErr error

func TestMain(m *testing.M) {
	checkStyleFile, checkStyleFileErr = os.Open("./testdata/output.xml")

	os.Exit(m.Run()) 
}

func TestCheck(t *testing.T) {
	if testing.Verbose() {
		log.SetLogThreshold(log.LevelInfo)
		log.SetStdoutThreshold(log.LevelInfo)
	}
	
	createEmptyGitRepo(t)

	runCheck(t, false, "")
	runCheck(t, false, "foo,5")
	runCheck(t, true, "foo,5")
	runCheck(t, false, "foo,5")

	t.Logf("Running check command w: %t i: %s", false, "foo,6")

	errCode := Check("", 0, true, "csv", strings.NewReader("foo,6"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}

	errCode = Check("", 0, true, "csv", strings.NewReader("foo,6"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}
}

func TestCheckPrefix(t *testing.T) {
	if testing.Verbose() {
		log.SetLogThreshold(log.LevelInfo)
		log.SetStdoutThreshold(log.LevelInfo)
	}

	createEmptyGitRepo(t)

	runCheckP(t, "foobar", true, "foo,5")
	// Running a check against a different prefix should still work
	runCheckP(t, "barfoo", true, "foo,6")

	t.Logf("Running check command p: %s w: %t i: %s", "foobar", false, "foo,6")

	errCode := Check("foobar", 0, false, "csv", strings.NewReader("foo,6"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}
}

func TestCheckSlack(t *testing.T) {
	if testing.Verbose() {
		log.SetLogThreshold(log.LevelInfo)
		log.SetStdoutThreshold(log.LevelInfo)
	}

	createEmptyGitRepo(t)

	slack := 5

	runCheckPS(t, "pageweight", slack, true, "gzippedjs,10")
	runCheckPS(t, "pageweight", slack, false, "gzippedjs,15")

	t.Logf("Running check command p: %s w: %t i: %s", "pageweight", false, "gzippedjs,16")

	errCode := Check("pageweight", slack, false, "csv", strings.NewReader("gzippedjs,16"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}
}

func TestCheckExcuse(t *testing.T) {
	if testing.Verbose() {
		log.SetLogThreshold(log.LevelInfo)
		log.SetStdoutThreshold(log.LevelInfo)
	}

	createEmptyGitRepo(t)

	runCheckP(t, "foobar", true, "foo,5")
	// Increase on "barfoo" prefix is okay
	runCheckP(t, "barfoo", true, "foo,6")

	t.Logf("Running check command p: %s w: %t i: %s", "foobar", false, "foo,6")

	errCode := Check("foobar", 0, false, "csv", strings.NewReader("foo,6"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}

	writeExcuse(t, "foobar", "foo", "PROD's down right now, I'll clean foo up later")

	runCheckP(t, "foobar", true, "foo,6")

	t.Logf("Running check command p: %s w: %t i: %s", "barfoo", false, "foo,7")

	errCode = Check("barfoo", 0, false, "csv", strings.NewReader("foo,7"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}
}

func TestCheckWithCheckstyleInput(t *testing.T) {
	if testing.Verbose() {
		log.SetLogThreshold(log.LevelInfo)
		log.SetStdoutThreshold(log.LevelInfo)
	}
	
	if checkStyleFileErr != nil {
		t.Fatalf("Failure opening test data", checkStyleFileErr)
	}

	createEmptyGitRepo(t)

	t.Logf("Running check command p: %s w: %t i: %s", "jshint", true, checkStyleFile)
	
	errCode := Check("jshint", 0, true, "checkstyle", checkStyleFile)

	if errCode != 0 {
		t.Fatalf("Check command failed! Error code: %d", errCode)
	}

	t.Logf("Running check command p: %s w: %t i: %s", "jshint", false, "errors,951")

	errCode = Check("jshint", 0, false, "csv", strings.NewReader("errors,951"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}
}

func writeExcuse(t *testing.T, prefix string, measure string, excuse string) {
	t.Logf("Running excuse command p: %s m: %s, e: %s", prefix, measure, excuse)

	errCode := Excuse(prefix, measure, excuse)

	if errCode != 0 {
		t.Fatalf("Excuse command failed! Error code: %d", errCode)
	}
}

func runCheck(t *testing.T, write bool, input string) {
	runCheckP(t, "", write, input)
}

func runCheckP(t *testing.T, prefix string, write bool, input string) {
	runCheckPS(t, prefix, 0, write, input)
}

func runCheckPS(t *testing.T, prefix string, slack int, write bool, input string) {
	t.Logf("Running check command p: %s s: %d, w: %t i: %s", prefix, slack, write, input)

	errCode := Check(prefix, slack, write, "csv", strings.NewReader(input))

	if errCode != 0 {
		t.Fatalf("Check command failed! Error code: %d", errCode)
	}
}

func createEmptyGitRepo(t *testing.T) string {
	repo, err := ioutil.TempDir(os.TempDir(), "git-ratchet-test-")

	if err != nil {
		t.Fatalf("Failed to create directory %s", repo)
	}

	err = os.Chdir(repo)

	if err != nil {
		t.Fatalf("Failed to init repository %s", repo)
	}

	runCommand(t, repo, exec.Command("git", "init", repo))
	runCommand(t, repo, exec.Command("git", "config", "user.email", "test@example.com"))
	runCommand(t, repo, exec.Command("git", "config", "user.name", "Test Name"))

	runCommand(t, repo, exec.Command("git", "add", createFile(t, repo, "README").Name()))
	runCommand(t, repo, exec.Command("git", "commit", "-m", "First Commit"))
	runCommand(t, repo, exec.Command("git", "add", createFile(t, repo, "test.txt").Name()))
	runCommand(t, repo, exec.Command("git", "commit", "-m", "Second Commit"))

	t.Logf("Init repository %s", repo)

	return repo
}

func runCommand(t *testing.T, repo string, c *exec.Cmd) {
	t.Logf("Running command %s", strings.Join(c.Args, " "))

	output, err := c.CombinedOutput()

	if err != nil {
		t.Fatalf("Failed to init repository %s, %s, %s", repo, err, output)
	}
}

func createFile(t *testing.T, repo string, filename string) *os.File {
	file, err := os.Create(filepath.Join(repo, filename))

	if err != nil {
		t.Fatalf("Failed to init repository %s", repo)
	}

	return file
}
