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

	errCode := Check("", 0, false, true, "csv", false, strings.NewReader("foo,6"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}

	errCode = Check("", 0, false, true, "csv", false, strings.NewReader("foo,6"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}
}

func TestZeroMissing(t *testing.T) {
	if testing.Verbose() {
		log.SetLogThreshold(log.LevelInfo)
		log.SetStdoutThreshold(log.LevelInfo)
	}

	createEmptyGitRepo(t)

	runCheck(t, true, "foo,5")

	t.Logf("Running check command w: %t i: %s", false, "")

	errCode := Check("", 0, false, true, "csv", false, strings.NewReader(""))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}

	t.Logf("Running check command zero on missing w: %t i: %s", false, "")

	errCode = Check("", 0, false, true, "csv", false, strings.NewReader(""))

	if errCode != 0 {
		t.Fatalf("Check command failed unexpectedly!")
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

	errCode := Check("foobar", 0, false, false, "csv", false, strings.NewReader("foo,6"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}
}

func TestCheckSlackValue(t *testing.T) {
	if testing.Verbose() {
		log.SetLogThreshold(log.LevelInfo)
		log.SetStdoutThreshold(log.LevelInfo)
	}

	createEmptyGitRepo(t)

	slack := 5.0
	usePercents := false

	runCheckPS(t, "pageweight", slack, usePercents, true, "gzippedjs,10")
	runCheckPS(t, "pageweight", slack, usePercents, false, "gzippedjs,15")

	t.Logf("Running check command p: %s w: %t i: %s", "pageweight", false, "gzippedjs,16")

	errCode := Check("pageweight", slack, usePercents, false, "csv", false, strings.NewReader("gzippedjs,16"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}
}

func TestCheckSlackPercent(t *testing.T) {
	if testing.Verbose() {
		log.SetLogThreshold(log.LevelInfo)
		log.SetStdoutThreshold(log.LevelInfo)
	}

	createEmptyGitRepo(t)

	slack := 20.0
	usePercents := true

	runCheckPS(t, "pageweight", slack, usePercents, true, "gzippedjs,100")
	runCheckPS(t, "pageweight", slack, usePercents, false, "gzippedjs,101")

	t.Logf("Running check command p: %s w: %t i: %s", "pageweight", false, "gzippedjs,120")

	errCode := Check("pageweight", slack, usePercents, false, "csv", false, strings.NewReader("gzippedjs,120"))

	if errCode != 0 {
		t.Fatalf("Check command failed unexpectedly!")
	}

	t.Logf("Running check command p: %s w: %t i: %s", "pageweight", false, "gzippedjs,160")

	errCode = Check("pageweight", slack, usePercents, false, "csv", false, strings.NewReader("gzippedjs,160"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}
}

func TestCheckExcuse(t *testing.T) {
	if testing.Verbose() {
		log.SetLogThreshold(log.LevelInfo)
		log.SetStdoutThreshold(log.LevelInfo)
	}

	repo := createEmptyGitRepo(t)

	runCheckP(t, "foobar", true, "foo,5")
	// Increase on "barfoo" prefix is okay
	runCheckP(t, "barfoo", true, "foo,6")

	t.Logf("Running check command p: %s w: %t i: %s", "foobar", false, "foo,6")

	errCode := Check("foobar", 0, false, false, "csv", false, strings.NewReader("foo,6"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}

	writeExcuse(t, "foobar", "foo", "PROD's down right now, I'll clean foo up later")

	runCheckP(t, "foobar", true, "foo,6")

	t.Logf("Running check command p: %s w: %t i: %s", "barfoo", false, "foo,7")

	errCode = Check("barfoo", 0, false, false, "csv", false, strings.NewReader("foo,7"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}

	runCommand(t, repo, exec.Command("git", "add", createFile(t, repo, "test2.txt").Name()))
	runCommand(t, repo, exec.Command("git", "commit", "-m", "Third Commit"))

	runCheckP(t, "foobar", true, "foo,6")

	runCommand(t, repo, exec.Command("git", "add", createFile(t, repo, "test3.txt").Name()))
	runCommand(t, repo, exec.Command("git", "commit", "-m", "Fourth Commit"))

	runCheckP(t, "foobar", true, "foo,6")
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

	errCode := Check("jshint", 0, false, true, "checkstyle", false, checkStyleFile)

	if errCode != 0 {
		t.Fatalf("Check command failed! Error code: %d", errCode)
	}

	t.Logf("Running check command p: %s w: %t i: %s", "jshint", false, "errors,951")

	errCode = Check("jshint", 0, false, false, "csv", false, strings.NewReader("errors,951"))

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
	runCheckPS(t, prefix, 0, false, write, input)
}

func runCheckPS(t *testing.T, prefix string, slack float64, usePercents bool, write bool, input string) {
	t.Logf("Running check command p: %s s: %g, sp: %t, w: %t i: %s", prefix, slack, usePercents, write, input)

	errCode := Check(prefix, slack, usePercents, write, "csv", false, strings.NewReader(input))

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
