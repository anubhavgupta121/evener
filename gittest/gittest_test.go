package gittest

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// ambient git state must not reach a fixture: a hostile environment that
// selects another repository or injects config is ignored by Run.
func TestRunIgnoresAmbientGitEnvironment(t *testing.T) {
	other := t.TempDir()
	t.Setenv("GIT_DIR", filepath.Join(other, "nope"))
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "user.name")
	t.Setenv("GIT_CONFIG_VALUE_0", "ambient")
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(other, "missing"))

	dir := t.TempDir()
	Run(t, dir, "init", "-q")
	Run(t, dir, "commit", "-q", "--allow-empty", "-m", "init")
	if got := Run(t, dir, "log", "-1", "--format=%an"); got == "ambient" {
		t.Fatalf("ambient GIT_CONFIG_* reached the fixture: author %q", got)
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Fatalf("repository not created in dir despite ambient GIT_DIR: %v", err)
	}
}

func TestNewLinkedWorktreeSharesMainRepository(t *testing.T) {
	main, wt := NewLinkedWorktree(t)
	common := Run(t, wt, "rev-parse", "--path-format=absolute", "--git-common-dir")
	want, err := filepath.EvalSymlinks(filepath.Join(main, ".git"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := filepath.EvalSymlinks(common)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("linked worktree common dir = %q, want main repo %q", got, want)
	}
	if branch := Run(t, wt, "rev-parse", "--abbrev-ref", "HEAD"); branch != "feat" {
		t.Fatalf("linked worktree branch = %q, want feat", branch)
	}
}

func TestQuiescentCommitArgsInsertsBeforeCommit(t *testing.T) {
	got := QuiescentCommitArgs([]string{"-c", "x=1", "commit", "-m", "m"})
	want := []string{"-c", "x=1", "-c", "maintenance.auto=false", "commit", "-m", "m"}
	if !slices.Equal(got, want) {
		t.Fatalf("QuiescentCommitArgs = %v, want %v", got, want)
	}
	if got := QuiescentCommitArgs([]string{"status"}); !slices.Equal(got, []string{"status"}) {
		t.Fatalf("non-commit args changed: %v", got)
	}
}

func TestMaintenanceProbeUsesEffectiveBoolean(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		active bool
	}{
		{name: "false", args: []string{"-c", "maintenance.auto=false", "commit"}},
		{name: "zero", args: []string{"-c", "maintenance.auto=0", "commit"}},
		{name: "no", args: []string{"-c", "maintenance.auto=no", "commit"}},
		{name: "off", args: []string{"-c", "maintenance.auto=off", "commit"}},
		{name: "case insensitive false", args: []string{"-c", "maintenance.auto=FaLsE", "commit"}},
		{name: "last false", args: []string{"-c", "maintenance.auto=true", "-c", "maintenance.auto=0", "commit"}},
		{name: "true", args: []string{"-c", "maintenance.auto=true", "commit"}, active: true},
		{name: "last true", args: []string{"-c", "maintenance.auto=0", "-c", "maintenance.auto=on", "commit"}, active: true},
		{name: "absent", args: []string{"commit"}, active: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			probe := NewMaintenanceProbe()
			dir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(dir, ".git", "objects"), 0o755); err != nil {
				t.Fatalf("mkdir objects: %v", err)
			}
			var gotArgs []string
			wrapped := probe.Run(func(_ string, args ...string) (string, error) {
				gotArgs = slices.Clone(args)
				return "", nil
			})
			result := make(chan error, 1)
			go func() {
				_, err := wrapped(dir, test.args...)
				result <- err
			}()
			if test.active {
				<-probe.Ready
				probe.Published()
				if err := <-result; err != nil {
					t.Fatalf("probe run: %v", err)
				}
				<-probe.Exited
				if probe.MutatorErr != nil {
					t.Fatalf("maintenance mutator: %v", probe.MutatorErr)
				}
				if !probe.PublishedWhileActive {
					t.Fatal("expected publication overlap")
				}
				if !slices.Contains(gotArgs, "maintenance.auto=false") {
					t.Fatalf("real runner args = %v, missing quiescent config", gotArgs)
				}
				return
			}
			if err := <-result; err != nil {
				t.Fatalf("probe run: %v", err)
			}
			select {
			case <-probe.Ready:
				t.Fatal("suppressed config launched synthetic maintenance")
			default:
			}
			if !slices.Equal(gotArgs, test.args) {
				t.Fatalf("real runner args = %v, want %v", gotArgs, test.args)
			}
		})
	}
}
