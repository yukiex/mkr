package plugin

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mackerelio/mkr/logger"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
)

var CommandPlugin = cli.Command{
	Name:        "plugin",
	Usage:       "Manage mackerel plugin",
	Description: `Manage mackerel plugin`,
	Subcommands: []cli.Command{
		{
			Name:        "install",
			Usage:       "install mackerel plugin",
			Description: `WIP`,
			Action:      doPluginInstall,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "prefix",
					Usage: "plugin install location",
				},
			},
		},
	},
	Hidden: true,
}

func doPluginInstall(c *cli.Context) error {
	argInstallTarget := c.Args().First()
	if argInstallTarget == "" {
		return fmt.Errorf("Specify install name")
	}
	it, err := parseInstallTarget(argInstallTarget)
	if err != nil {
		return errors.Wrap(err, "failed to install plugin")
	}

	pluginDir, err := setupPluginDir(c.String("prefix"))
	if err != nil {
		return errors.Wrap(err, "failed to install plugin")
	}

	u, err := it.makeDownloadURL()
	if err != nil {
		return errors.Wrap(err, "failed to install plugin while making download url")
	}
	err = install(u, filepath.Join(pluginDir, "bin"))

	fmt.Println("do plugin install [wip]")
	return nil
}

func install(u, binPath string) error {
	logger.Log("", fmt.Sprintf("download %s\n", u))
	archivePath, err := download(u)
	if err != nil {
		return errors.Wrap(err, "failed to download")
	}
	tmpdir := filepath.Dir(archivePath)
	defer os.RemoveAll(tmpdir)

	workDir := filepath.Join(tmpdir, "work")
	os.MkdirAll(workDir, 0755)

	logger.Log("", fmt.Sprintf("extract %s\n", path.Base(u)))
	err = archiver.Zip.Open(archivePath, workDir)
	if err != nil {
		return errors.Wrap(err, "failed to extract")
	}
	fmt.Println(tmpdir)

	return pickupExecutable(workDir, binPath)
}

func download(u string) (fpath string, err error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		err = errors.Wrap(err, "failed to create request")
		return
	}
	req.Header.Set("User-Agent", "mkr-plugin-installer/0.0.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = errors.Wrap(err, "failed to create request")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("http response not OK. code: %d, url: %s", resp.StatusCode, u)
		return
	}
	archiveBase := path.Base(u)
	tempdir, err := ioutil.TempDir("", "mkr-plugin-installer-")
	if err != nil {
		err = errors.Wrap(err, "failed to create tempdir")
		return
	}
	defer func() {
		if err != nil {
			os.RemoveAll(tempdir)
		}
	}()
	fpath = filepath.Join(tempdir, archiveBase)
	f, err := os.OpenFile(filepath.Join(tempdir, archiveBase), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		err = errors.Wrap(err, "failed to open file")
		return
	}
	defer f.Close()
	// progressR := progbar(resp.Body, resp.ContentLength)
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		err = errors.Wrap(err, "failed to read response")
		return
	}
	return fpath, nil
}

func pickupExecutable(workdir, bindir string) error {
	return filepath.Walk(workdir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if name := info.Name(); (info.Mode()&0111) != 0 && looksLikePlugin(name) {
			return place(path, filepath.Join(bindir, name))
		}
		return nil
	})
}

func looksLikePlugin(name string) bool {
	return strings.HasPrefix(name, "check-") || strings.HasPrefix(name, "mackerel-plugin-")
}

func place(src, dest string) error {
	_, err := os.Stat(dest)
	if err == nil {
		logger.Log("", fmt.Sprintf("%s already exists. skip installing for now", dest))
		return nil
	}
	logger.Log("", fmt.Sprintf("install %s\n", filepath.Base(dest)))
	err = os.Rename(src, dest)
	// When `src` and `dest` are on different file systems, os.Rename sometimes fails
	if err != nil {
		return copyExecutable(src, dest)
	}
	return nil
}

func copyExecutable(srcName string, destName string) error {
	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(destName)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(srcName)
	if err != nil {
		return err
	}

	return os.Chmod(destName, fileInfo.Mode())
}

func setupPluginDir(prefix string) (string, error) {
	if prefix == "" {
		prefix = "/opt/mackerel-agent/plugins"
	}
	err := os.MkdirAll(filepath.Join(prefix, "bin"), 0755)
	if err != nil {
		return "", errors.Wrap(err, "failed to setup plugin directory")
	}
	return prefix, nil
}

type installTarget struct {
	owner      string
	repo       string
	pluginName string
	releaseTag string
}

func (it *installTarget) makeDownloadURL() (string, error) {
	if it.owner != "" && it.repo != "" {
		if it.releaseTag == "" {
			return "", fmt.Errorf("not implemented")
		}
		filename := fmt.Sprintf("%s_%s_%s.zip", it.repo, runtime.GOOS, runtime.GOARCH)
		return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
			it.owner, it.repo, it.releaseTag, filename), nil
	}
	return "", fmt.Errorf("not implemented")
}

func parseInstallTarget(target string) (*installTarget, error) {
	it := &installTarget{}

	ownerRepoAndReleaseTag := strings.Split(target, "@")
	var ownerRepo string
	switch len(ownerRepoAndReleaseTag) {
	case 1:
		ownerRepo = ownerRepoAndReleaseTag[0]
	case 2:
		ownerRepo = ownerRepoAndReleaseTag[0]
		it.releaseTag = ownerRepoAndReleaseTag[1]
	default:
		return nil, fmt.Errorf("Install target is invalid: %s", target)
	}

	ownerAndRepo := strings.Split(ownerRepo, "/")
	switch len(ownerAndRepo) {
	case 1:
		it.pluginName = ownerAndRepo[0]
	case 2:
		it.owner = ownerAndRepo[0]
		it.repo = ownerAndRepo[1]
	default:
		return nil, fmt.Errorf("Install target is invalid: %s", target)
	}

	return it, nil
}
