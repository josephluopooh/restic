//go:build darwin || freebsd || linux

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/restic/restic/internal/backend/all"
	"github.com/restic/restic/internal/global"
	"golang.org/x/sys/unix"
)

func TestValidateMountpointWithoutWriteAccess(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("non-root FUSE mounts require a writable mountpoint")
	}
	tempdir := t.TempDir()
	repo := filepath.Join(tempdir, "repo")
	mountpoint := filepath.Join(tempdir, "mount")
	if err := os.Mkdir(repo, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(mountpoint, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(mountpoint, 0700) })

	gopts := global.Options{Repo: repo, Backends: all.Backends()}
	if err := validateMountpoint(mountpoint, gopts); err == nil {
		t.Fatal("non-root user should not mount on a directory without write access")
	}
}

func TestValidateMountpointReadOnlyRoot(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root")
	}
	mountpoint := os.Getenv("RESTIC_TEST_READONLY_MOUNTPOINT")
	repo := os.Getenv("RESTIC_TEST_REPO_PATH")
	if mountpoint == "" || repo == "" {
		t.Skip("requires a read-only filesystem and a disjoint repository path")
	}
	if err := unix.Access(mountpoint, unix.W_OK|unix.X_OK); err == nil {
		t.Skip("mountpoint allows writes")
	}

	gopts := global.Options{Repo: repo, Backends: all.Backends()}
	if err := validateMountpoint(mountpoint, gopts); err != nil {
		t.Fatalf("root should be allowed to try a read-only mountpoint: %v", err)
	}
}
