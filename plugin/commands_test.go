package plugin

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
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

func TestParseInstallTarget(t *testing.T) {
	testCases := []struct {
		Name   string
		Input  string
		Output installTarget
	}{
		{
			Name:  "Plugin name only",
			Input: "mackerel-plugin-sample",
			Output: installTarget{
				pluginName: "mackerel-plugin-sample",
			},
		},
		{
			Name:  "Plugin name and release tag",
			Input: "mackerel-plugin-sample@v0.0.1",
			Output: installTarget{
				pluginName: "mackerel-plugin-sample",
				releaseTag: "v0.0.1",
			},
		},
		{
			Name:  "Owner and repo",
			Input: "mackerelio/mackerel-plugin-sample",
			Output: installTarget{
				owner: "mackerelio",
				repo:  "mackerel-plugin-sample",
			},
		},
		{
			Name:  "Owner and repo with release tag",
			Input: "mackerelio/mackerel-plugin-sample@v1.0.1",
			Output: installTarget{
				owner:      "mackerelio",
				repo:       "mackerel-plugin-sample",
				releaseTag: "v1.0.1",
			},
		},
	}

	for _, tc := range testCases {
		t.Logf("testing: %s\n", tc.Name)
		it, err := parseInstallTarget(tc.Input)
		if err != nil {
			t.Errorf("%s(err): error should be nil but: %+v", tc.Name, err)
			continue
		}
		if !reflect.DeepEqual(*it, tc.Output) {
			t.Errorf("%s(parse): \n out =%+v\n want %+v", tc.Name, *it, tc.Output)
		}
	}
}

func TestParseInstallTarget_error(t *testing.T) {
	testCases := []struct {
		Name   string
		Input  string
		Output string
	}{
		{
			Name:   "Too many @",
			Input:  "mackerel-plugin-sample@v0.0.1@v0.1.0",
			Output: "Install target is invalid: mackerel-plugin-sample@v0.0.1@v0.1.0",
		},
		{
			Name:   "Too many /",
			Input:  "mackerelio/hatena/mackerel-plugin-sample",
			Output: "Install target is invalid: mackerelio/hatena/mackerel-plugin-sample",
		},
	}

	for _, tc := range testCases {
		t.Logf("testing: %s\n", tc.Name)
		_, err := parseInstallTarget(tc.Input)
		if err == nil {
			t.Errorf("%s(err): err should be occurred but nil", tc.Name)
			continue
		}
		if err.Error() != tc.Output {
			t.Errorf("%s(parse_error): \n out =%s\n want %s", tc.Name, err.Error(), tc.Output)
		}
	}
}
