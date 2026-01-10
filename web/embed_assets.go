package web

import (
	"embed"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// AssetVersion is injected at build time (e.g. git sha) via:
// -ldflags "-X ehang.io/nps/web.AssetVersion=<sha>"
var AssetVersion string

//go:embed static/** views/**
var embeddedWeb embed.FS

func SyncAssetsToRunPath(runPath string) error {
	webRoot := filepath.Join(runPath, "web")

	tmpRoot := webRoot + ".tmp-" + strconv.Itoa(os.Getpid()) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	if err := os.RemoveAll(tmpRoot); err != nil {
		return err
	}
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmpRoot) }()

	if err := syncDir("views", tmpRoot); err != nil {
		return err
	}
	if err := syncDir("static", tmpRoot); err != nil {
		return err
	}

	version := AssetVersion
	if version == "" {
		version = "unknown"
	}
	if err := os.WriteFile(filepath.Join(tmpRoot, ".asset_version"), []byte(version+"\n"), 0o644); err != nil {
		return err
	}

	backupPath := webRoot + ".bak-" + time.Now().Format("20060102150405")
	if _, err := os.Stat(webRoot); err == nil {
		if err := os.RemoveAll(backupPath); err != nil {
			return err
		}
		if err := os.Rename(webRoot, backupPath); err != nil {
			if err := os.RemoveAll(webRoot); err != nil {
				return err
			}
		}
	}

	if err := os.Rename(tmpRoot, webRoot); err != nil {
		if _, statErr := os.Stat(backupPath); statErr == nil {
			_ = os.Rename(backupPath, webRoot)
		}
		return err
	}
	_ = os.RemoveAll(backupPath)
	return nil
}

func syncDir(root string, tmpWebRoot string) error {
	return fs.WalkDir(embeddedWeb, root, func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dstPath := filepath.Join(tmpWebRoot, filepath.FromSlash(path))
		if dirEntry.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return err
		}

		srcFile, err := embeddedWeb.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}
