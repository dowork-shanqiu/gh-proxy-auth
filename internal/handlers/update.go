package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// UpdateHandler handles version checking and self-update.
type UpdateHandler struct{}

func NewUpdateHandler() *UpdateHandler {
	return &UpdateHandler{}
}

// RestartChan is used to signal main that the binary has been replaced and a
// graceful restart should begin.
var RestartChan = make(chan struct{}, 1)

// updateMu prevents concurrent update attempts.
var updateMu sync.Mutex

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// currentVersionFn is a variable so it can be overridden by the version
// injected at build time (see version.go in the root package).
// Handlers package cannot import main, so we use a setter.
var currentVersionFn func() string

// SetVersionFunc registers a function that returns the running version string.
func SetVersionFunc(fn func() string) {
	currentVersionFn = fn
}

func currentVersion() string {
	if currentVersionFn != nil {
		return currentVersionFn()
	}
	return "dev"
}

// CheckUpdate returns the latest GitHub release info and whether an update is
// available.
func (h *UpdateHandler) CheckUpdate(c *gin.Context) {
	rel, err := fetchLatestRelease()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取最新版本失败: " + err.Error()})
		return
	}

	cur := currentVersion()
	hasUpdate := rel.TagName != cur && cur != "dev"

	assetName := fmt.Sprintf("gh-proxy-auth-%s-%s", runtime.GOOS, runtime.GOARCH)
	downloadURL := ""
	for _, a := range rel.Assets {
		if strings.EqualFold(a.Name, assetName) {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"current_version": cur,
		"latest_version":  rel.TagName,
		"has_update":      hasUpdate,
		"download_url":    downloadURL,
	})
}

// ApplyUpdate downloads the latest release binary, replaces the current
// executable, then triggers a graceful restart via RestartChan.
func (h *UpdateHandler) ApplyUpdate(c *gin.Context) {
	if !updateMu.TryLock() {
		c.JSON(http.StatusConflict, gin.H{"error": "更新正在进行中，请稍后"})
		return
	}

	rel, err := fetchLatestRelease()
	if err != nil {
		updateMu.Unlock()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取最新版本失败: " + err.Error()})
		return
	}

	assetName := fmt.Sprintf("gh-proxy-auth-%s-%s", runtime.GOOS, runtime.GOARCH)
	downloadURL := ""
	for _, a := range rel.Assets {
		if strings.EqualFold(a.Name, assetName) {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		updateMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("未找到适合当前系统的安装包 (%s)", assetName)})
		return
	}

	// Return immediately so the client gets a response before the server restarts.
	c.JSON(http.StatusOK, gin.H{
		"message": "开始下载更新，下载完成后服务将自动重启",
		"version": rel.TagName,
	})

	// Perform the download and restart in background.
	go func() {
		defer updateMu.Unlock()
		if err := downloadAndReplace(downloadURL); err != nil {
			// Nothing we can do here except log – the response is already sent.
			fmt.Fprintf(os.Stderr, "[update] 更新失败: %v\n", err)
			return
		}
		// Signal main goroutine to restart.
		RestartChan <- struct{}{}
	}()
}

// downloadAndReplace downloads the binary at url and atomically replaces the
// current executable with it.
func downloadAndReplace(url string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取当前程序路径失败: %w", err)
	}
	// Resolve symlinks so we operate on the real file.
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("解析程序路径失败: %w", err)
	}

	// Download to a temporary file in the same directory to enable atomic rename.
	dir := filepath.Dir(execPath)
	tmp, err := os.CreateTemp(dir, ".gh-proxy-auth-update-*")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		// Remove temp file on failure (ignored if already renamed).
		os.Remove(tmpName)
	}()

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url) //nolint:noctx
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，HTTP 状态码: %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("关闭临时文件失败: %w", err)
	}

	// Copy permissions from current binary.
	info, err := os.Stat(execPath)
	if err != nil {
		return fmt.Errorf("获取当前程序权限失败: %w", err)
	}
	if err := os.Chmod(tmpName, info.Mode()); err != nil {
		return fmt.Errorf("设置临时文件权限失败: %w", err)
	}

	// Atomic replace.
	if err := os.Rename(tmpName, execPath); err != nil {
		return fmt.Errorf("替换程序文件失败: %w", err)
	}

	return nil
}

// fetchLatestRelease queries the GitHub API for the latest release.
func fetchLatestRelease() (*githubRelease, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet,
		"https://api.github.com/repos/dowork-shanqiu/gh-proxy-auth/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回 %d", resp.StatusCode)
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// DoRestart performs a graceful exec-restart: it replaces the current process
// image with the (already-replaced) binary on disk. Called from main after the
// HTTP server has been shut down.
func DoRestart() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}
	return syscall.Exec(execPath, os.Args, os.Environ())
}
