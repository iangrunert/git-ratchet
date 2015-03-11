package cmd

import (
	"testing"
	"os"
	"os/exec"
	"io/ioutil"
	"strings"
	"path/filepath"

	log "github.com/spf13/jwalterweatherman"
)

func TestCheck(t *testing.T) {
	if (testing.Verbose()) {
		log.SetLogThreshold(log.LevelInfo)
		log.SetStdoutThreshold(log.LevelInfo)
	}

	repo := createEmptyGitRepo(t)
	
	err := os.Chdir(repo)
	
	if err != nil {
		t.Fatalf("Failed to change to directory %s", repo)
	}
	
	runCheck(t, false, "")
	runCheck(t, false, "foo,5")
	runCheck(t, true, "foo,5")
	runCheck(t, false, "foo,5")

	t.Logf("Running check command w: %t i: %s", false, "foo,6")

	errCode := Check(false, strings.NewReader("foo,6"))

	if errCode != 50 {
		t.Fatalf("Check command passed unexpectedly!")
	}
}

func runCheck(t *testing.T, write bool, input string) {
	t.Logf("Running check command w: %t i: %s", write, input)

	errCode := Check(write, strings.NewReader(input))

	if errCode != 0 {
		t.Fatalf("Check command failed! Error code: %d", errCode)
	}
}

func createEmptyGitRepo(t *testing.T) string {
	testDir, err := ioutil.TempDir(os.TempDir(), "git-ratchet-test-")
		
	if err != nil {
		t.Fatalf("Failed to create directory %s", testDir)
	}

	err = os.Chdir(testDir)
	
	if err != nil {
		t.Fatalf("Failed to init repository %s", testDir)
	}
	
	runCommand(t, testDir, exec.Command("git", "init", testDir))
	runCommand(t, testDir, exec.Command("git", "config", "user.email", "test@example.com"))	
	runCommand(t, testDir, exec.Command("git", "config", "user.name", "Test Name"))	

	runCommand(t, testDir, exec.Command("git", "add", createFile(t, testDir, "README").Name()))
	runCommand(t, testDir, exec.Command("git", "commit", "-m", "First Commit"))
	runCommand(t, testDir, exec.Command("git", "add", createFile(t, testDir, "test.txt").Name()))
	runCommand(t, testDir, exec.Command("git", "commit", "-m", "Second Commit"))

	t.Logf("Init repository %s", testDir)
	
	return testDir
}

func runCommand(t *testing.T, testDir string, c *exec.Cmd) {
	t.Logf("Running command %s", strings.Join(c.Args, " "))
	
	output, err := c.CombinedOutput()
	
	if err != nil {
		t.Fatalf("Failed to init repository %s, %s, %s", testDir, err, output)
	}
}

func createFile(t *testing.T, repo string, filename string) *os.File {
	file, err := os.Create(filepath.Join(repo, filename))
	
	if err != nil {
		t.Fatalf("Failed to init repository %s", repo)
	}

	return file
}
