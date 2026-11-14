package bridge

import (
	"path/filepath"
	"testing"

	"omp-telegram/internal/config"
	"omp-telegram/internal/i18n"
	"omp-telegram/internal/store"
	"omp-telegram/internal/telegram"
)

func lockedWorker(t *testing.T, locked bool) *worker {
	t.Helper()
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	b := testBridge(t, &Bridge{
		db:  db,
		bot: telegram.User{ID: 99},
		cfg: config.Config{WorkspaceRoot: t.TempDir(), WorkspaceLocked: locked},
	})
	w := testWorker(t, &worker{b: b})
	w.key = target{chat: -10, thread: 11}
	return w
}

func TestNewWorkspaceLockedRejectsArgument(t *testing.T) {
	w := lockedWorker(t, true)
	if _, err := w.newWorkspace(t.TempDir()); err == nil ||
		w.errorText(err) != w.tr(i18n.Msg.WorkspaceLockedCannotChange) {
		t.Fatalf("locked /new <path> error = %v", err)
	}
}

func TestNewWorkspaceLockedBareReusesBinding(t *testing.T) {
	w := lockedWorker(t, true)
	workspace := filepath.Join(t.TempDir(), "bound")
	if err := w.b.db.Save(store.Binding{
		Bot: w.b.bot.ID, Chat: w.key.chat, Thread: w.key.thread,
		Workspace: workspace, Generation: 1,
	}); err != nil {
		t.Fatal(err)
	}
	got, err := w.newWorkspace("")
	if err != nil || got != workspace {
		t.Fatalf("bare /new = %q, %v; want bound workspace %q", got, err, workspace)
	}
}

func TestNewWorkspaceUnlockedAcceptsArgument(t *testing.T) {
	w := lockedWorker(t, false)
	target := t.TempDir()
	got, err := w.newWorkspace(target)
	if err != nil || got != target {
		t.Fatalf("unlocked /new <path> = %q, %v; want %q", got, err, target)
	}
}
