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

func TestEmptyRepository(t *testing.T) {
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

	if errCode == 0 {
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
	
	gitInit := exec.Command("git", "init", testDir)
	
	_, err = gitInit.CombinedOutput()
	
	if err != nil {
		t.Fatalf("Failed to init repository %s", testDir)
	}
	
	file, err := os.Create(filepath.Join(testDir, "README"))
	
	if err != nil {
		t.Fatalf("Failed to init repository %s", testDir)
	}

	err = os.Chdir(testDir)
	
	if err != nil {
		t.Fatalf("Failed to init repository %s", testDir)
	}
	
	gitAdd := exec.Command("git", "add", file.Name())
	
	_, err = gitAdd.CombinedOutput()
	
	if err != nil {
		t.Fatalf("Failed to init repository %s", testDir)
	}
	
	gitCommit := exec.Command("git", "commit", "-m", "First Commit")

	_, err = gitCommit.CombinedOutput()
	
	if err != nil {
		t.Fatalf("Failed to init repository %s", testDir)
	}
	
	t.Logf("Init repository %s", testDir)
	
	return testDir
}
