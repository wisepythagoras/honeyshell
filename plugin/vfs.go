package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const T_DIR = 1
const T_FILE = 2
const T_SYMLINK = 3
const T_ANY = 4

type VFSFile struct {
	Type     int         `json:"type"`
	Name     string      `json:"name"`
	Files    []VFSFile   `json:"files"`
	Contents string      `json:"contents"`
	Mode     os.FileMode `json:"mode"`
	Owner    string      `json:"owner"`
}

// Find will check the current file or its descendants for the file with name `name` and type `t`.
func (f *VFSFile) Find(name, path string, t int) (string, *VFSFile) {
	if name == path+f.Name && (f.Type == t || t == T_ANY) {
		return path + f.Name, f
	}

	if f.Type == T_DIR {
		for _, file := range f.Files {
			newPath := path + f.Name + "/"

			if filePath, foundFile := file.Find(name, newPath, t); foundFile != nil {
				return filePath, foundFile
			}
		}
	}

	return "", nil
}

func (f *VFSFile) StrMode() string {
	return f.Mode.String()
}

type VFS struct {
	Root VFSFile `json:"root"`
	Home string  `json:"home"`
	PWD  string  `json:"-"`
}

// Find tries to find a specific file in the virtual file system.
func (vfs *VFS) Find(name string, t int) (string, *VFSFile) {
	if strings.HasPrefix(name, "~") {
		homePath, homeDir := vfs.Find(vfs.Home, T_DIR)

		if name == "~" || name == "~/" {
			return strings.ReplaceAll(homePath, "/home/{}", "~"), homeDir
		}

		// TODO: Handle the edge case of ~/.. here.
		sanitizedPath := strings.ReplaceAll(name, "~/", "{}/")
		foundPath, foundFile := homeDir.Find(sanitizedPath, "", t)

		if foundFile != nil {
			foundPath = strings.ReplaceAll(foundPath, "{}/", "~/")
		}

		return foundPath, foundFile
	} else if name == "." || strings.HasPrefix(name, "./") {
		cwdPath, cwdDir := vfs.Find(vfs.PWD, T_DIR)

		if name == "." || name == "./" {
			return cwdPath, cwdDir
		}

		// fullPath := filepath.Join(vfs.PWD, name)
		sanitizedPath := strings.ReplaceAll(name, "./", "{}/")
		foundPath, foundFile := cwdDir.Find(sanitizedPath, "", t)

		if foundFile != nil {
			foundPath = strings.ReplaceAll(foundPath, "{}/", "./")
		}

		return foundPath, foundFile
	} else if name == ".." || strings.HasPrefix(name, "../") {
		fullPath := filepath.Join(vfs.PWD, name)
		return vfs.Find(fullPath, T_ANY)
	}

	sanitizedPath := filepath.Join(name)

	if sanitizedPath == "/" {
		return "", &vfs.Root
	}

	return vfs.Root.Find(sanitizedPath, "", t)
}

// ReadVFSJSONFile reads the JSON file which contains the the virtual file system model.
func ReadVFSJSONFile(path string) (*VFS, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	vfs := &VFS{}

	if err = json.Unmarshal(data, vfs); err != nil {
		return nil, err
	}

	return vfs, nil
}
