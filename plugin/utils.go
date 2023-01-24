package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

func LoadPlugins(path string, db *gorm.DB) ([]*Plugin, error) {
	files, err := os.ReadDir(path)
	plugins := []*Plugin{}

	if err != nil {
		return plugins, err
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}

		extFiles, err := os.ReadDir(filepath.Join(path, f.Name()))

		if err != nil {
			return plugins, err
		}

		var mainFile os.DirEntry = nil

		for _, extFile := range extFiles {
			if extFile.IsDir() {
				continue
			}

			if extFile.Name() == "main.lua" {
				mainFile = extFile
			}
		}

		if mainFile == nil {
			return plugins, fmt.Errorf("extension folder %q doesn't have an entry point (main.lua)", f.Name())
		}

		extension := &Plugin{
			Path: path,
			Dir:  f,
			Main: mainFile,
			DB:   db,
		}

		plugins = append(plugins, extension)
	}

	return plugins, nil
}

func stringToBytes(str string) []byte {
	return []byte(str)
}

func bytesToString(b []byte) string {
	return string(b)
}
