package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"teer/internal/domain"
)

const githubRepo = "triadmoko/teer"
const defaultAPIBase = "https://api.github.com"

type UpdaterService struct {
	version    string
	emitter    domain.EventEmitter
	httpClient *http.Client
	apiBaseURL string
	targetPath string
	onDone     func() // dipanggil setelah apply selesai; default: quit app
}

func NewUpdaterService(version string, emitter domain.EventEmitter) *UpdaterService {
	return &UpdaterService{
		version:    version,
		emitter:    emitter,
		apiBaseURL: defaultAPIBase,
	}
}

func (u *UpdaterService) client() *http.Client {
	if u.httpClient != nil {
		return u.httpClient
	}
	return &http.Client{Timeout: 15 * time.Second}
}

type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseURL     string `json:"releaseUrl"`
	DownloadURL    string `json:"downloadUrl"`
}

type ghRelease struct {
	TagName string    `json:"tag_name"`
	HTMLURL string    `json:"html_url"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func (u *UpdaterService) CheckUpdate() (*UpdateInfo, error) {
	base := u.apiBaseURL
	if base == "" {
		base = defaultAPIBase
	}
	url := fmt.Sprintf("%s/repos/%s/releases/latest", base, githubRepo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "teer/"+u.version)

	resp, err := u.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal cek update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &UpdateInfo{Available: false, CurrentVersion: u.version}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	current := strings.TrimPrefix(u.version, "v")

	info := &UpdateInfo{
		Available:      latest != current && current != "dev",
		CurrentVersion: u.version,
		LatestVersion:  rel.TagName,
		ReleaseURL:     rel.HTMLURL,
		DownloadURL:    pickAsset(rel.Assets),
	}
	return info, nil
}

func (u *UpdaterService) DownloadAndApply(downloadURL string) error {
	u.emitter.Emit("updater:progress", map[string]any{"stage": "downloading", "percent": 0})

	dlClient := u.httpClient
	if dlClient == nil {
		dlClient = &http.Client{Timeout: 5 * time.Minute}
	}
	resp, err := dlClient.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("gagal download: %w", err)
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", "teer-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := tmpFile.Write(buf[:n]); werr != nil {
				tmpFile.Close()
				return werr
			}
			downloaded += int64(n)
			if total > 0 {
				pct := int(downloaded * 100 / total)
				u.emitter.Emit("updater:progress", map[string]any{"stage": "downloading", "percent": pct})
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			tmpFile.Close()
			return rerr
		}
	}
	tmpFile.Close()

	u.emitter.Emit("updater:progress", map[string]any{"stage": "applying", "percent": 100})

	var selfPath string
	if u.targetPath != "" {
		selfPath = u.targetPath
	} else {
		selfPath, err = os.Executable()
		if err != nil {
			return err
		}
		selfPath, err = filepath.EvalSymlinks(selfPath)
		if err != nil {
			return err
		}
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return err
	}

	backupPath := selfPath + ".bak"
	if err := os.Rename(selfPath, backupPath); err != nil {
		return fmt.Errorf("gagal backup binary lama: %w", err)
	}

	if err := moveFile(tmpPath, selfPath); err != nil {
		// rollback
		_ = os.Rename(backupPath, selfPath)
		return fmt.Errorf("gagal replace binary: %w", err)
	}
	os.Remove(backupPath)

	u.emitter.Emit("updater:progress", map[string]any{"stage": "done", "percent": 100})

	done := u.onDone
	if done == nil {
		done = func() {
			time.Sleep(500 * time.Millisecond)
			application.Get().Quit()
		}
	}
	go done()

	return nil
}

// moveFile copies src ke dst, fallback dari rename (beda filesystem).
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func pickAsset(assets []ghAsset) string {
	os := runtime.GOOS     // "linux" | "windows" | "darwin"
	arch := runtime.GOARCH // "amd64" | "arm64"

	for _, a := range assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, os) && strings.Contains(name, arch) {
			return a.BrowserDownloadURL
		}
	}
	// fallback: cari platform saja tanpa arch
	for _, a := range assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, os) {
			return a.BrowserDownloadURL
		}
	}
	return ""
}
