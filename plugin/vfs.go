package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const T_DIR = 1
const T_FILE = 2
const T_SYMLINK = 3
const T_ANY = 4

type VFSFile struct {
	Type     int                `json:"type"`
	Name     string             `json:"name"`
	Files    map[string]VFSFile `json:"files"`
	Contents string             `json:"contents"`
	Mode     os.FileMode        `json:"mode"`
	Owner    string             `json:"owner"`
}

func (f *VFSFile) findFile(name string, rest []string) (string, *VFSFile, error) {
	if name == f.Name && len(rest) == 0 {
		return name, f, nil
	} else if name == f.Name && len(rest) > 0 {
		if f.Type != T_DIR {
			return "", nil, fmt.Errorf("file not a directory")
		}

		if file, ok := f.Files[rest[0]]; ok {
			newPath, newFile, err := file.findFile(rest[0], rest[1:])

			if err != nil {
				return "", nil, err
			}

			if name == "" {
				name = "/"
			}

			return filepath.Join(name, newPath), newFile, nil
		}
	}

	return "", nil, fmt.Errorf("file not found")
}

func (f *VFSFile) ForEach(callback func(*VFSFile, int)) {
	if f.Type == T_FILE {
		callback(f, 0)
		return
	}

	i := 0

	for _, file := range f.Files {
		callback(&file, i)
		i++
	}
}

func (f *VFSFile) StrMode() string {
	return f.Mode.String()
}

type VFS struct {
	Root     VFSFile `json:"root"`
	Home     string  `json:"home"`
	PWD      string  `json:"-"`
	Username string  `json:"-"`
}

func (vfs *VFS) resolveDotPath(path string) string {
	if path == "." || path == "./" {
		return vfs.PWD
	}

	return filepath.Join(vfs.PWD, path)
}

func (vfs *VFS) FindFile(path string) (string, *VFSFile, error) {
	if path == "" {
		path = "/"
	} else if len(path) > 2 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	if strings.HasPrefix(path, "~") {
		if path == "~" || path == "~/" || path == "~/." {
			path = vfs.Home
		} else {
			path = filepath.Join(vfs.Home, strings.Replace(path, "~/", "", 1))
		}
	} else if path == "." || strings.HasPrefix(path, "./") {
		path = vfs.resolveDotPath(path)
	} else if path == ".." || strings.HasPrefix(path, "../") {
		path = filepath.Join(vfs.PWD, path)
	} else if !strings.HasPrefix(path, "/") {
		path = filepath.Join(vfs.PWD, path)
	}

	if path == "/" {
		return "/", &vfs.Root, nil
	}

	parts := strings.Split(path, "/")

	return vfs.Root.findFile(parts[0], parts[1:])
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
