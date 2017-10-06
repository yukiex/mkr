package plugin

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func tempd(t *testing.T) string {
	tmpd, err := ioutil.TempDir("", "mkr-plugin-install")
	if err != nil {
		t.Fatal(err)
	}
	return tmpd
}

func TestSetupPluginDir(t *testing.T) {
	{
		tmpd := tempd(t)
		defer os.RemoveAll(tmpd)
		err := setupPluginDir(tmpd)
		if err != nil {
			t.Errorf("error should be nil but: %s", err)
		}
		fi, err := os.Stat(filepath.Join(tmpd, "bin"))
		if !(err == nil && fi.IsDir()) {
			t.Errorf("plugin directory should be created but isn't")
		}
	}

	{
		tmpd := tempd(t)
		defer os.RemoveAll(tmpd)
		err := os.Chmod(tmpd, 0500)
		if err != nil {
			t.Errorf("error occured while chmod directory: %s", err)
		}
		err = setupPluginDir(tmpd)
		if err == nil {
			t.Errorf("error should be occured while manipulate unpermitted directory, but it is nil")
		}
	}
}
