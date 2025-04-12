package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func editor(flPth string) error {
	cmd := exec.Command(editorEnvVar, flPth)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type tmpFileFunc func(tmpFl *os.File) error

func withTempFile(pattern string, tff tmpFileFunc) error {
	tmpFl, err := os.CreateTemp(config.TemporaryDirectory, pattern)
	if err != nil {
		return err
	}
	defer func() {
		tmpFlName := tmpFl.Name()
		err := tmpFl.Close()
		if err != nil {
			fmt.Println(err)
		}
		err = os.Remove(tmpFlName)
		if err != nil {
			fmt.Println(err)
		}
	}()
	return tff(tmpFl)
}

func getConfigEnvironmentVar(cg string) string {
	if !strings.HasPrefix(cg, "$") {
		return cg
	}

	return os.Getenv(strings.TrimPrefix(cg, "$"))
}

func getConf(cfgName string) (*configType, error) {
	conf := configType{}

	usrConfDir, err := os.UserConfigDir()
	if err != nil {
		return &conf, err
	}

	viper.AddConfigPath(".") // first file takes presidence
	viper.AddConfigPath(filepath.Join(usrConfDir, cfgName))
	viper.SetConfigName(cfgName)
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		return &conf, err
	}

	err = viper.Unmarshal(&conf)

	for hdr, val := range conf.Headers {
		conf.Headers[hdr] = getConfigEnvironmentVar(val)
	}

	return &conf, err
}
