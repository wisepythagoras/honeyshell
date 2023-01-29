package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const T_DIR = 1
const T_FILE = 2
const T_SYMLINK = 3
const T_ANY = 4

type Perm struct {
	Read  bool
	Write bool
	Exec  bool
}

type VFSFile struct {
	Type     int                `json:"type"`
	Name     string             `json:"name"`
	Files    map[string]VFSFile `json:"files"`
	Contents string             `json:"contents"`
	Mode     os.FileMode        `json:"mode"`
	Owner    string             `json:"owner"`
	Group    string             `json:"group"`
	ModTime  time.Time          `json:"mod_time"`
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

func (f *VFSFile) resolvePerms(op uint32) Perm {
	perm := Perm{}

	if op&4 == 4 {
		perm.Read = true
	}

	if op&2 == 2 {
		perm.Write = true
	}

	if op&1 == 1 {
		perm.Exec = true
	}

	return perm
}

func (f *VFSFile) CanAccess(user *User) Perm {
	if user == nil {
		return Perm{}
	}

	up := uint32(f.Mode / 100)
	gp := uint32(f.Mode % 100 / 10)
	op := uint32(f.Mode % 100 % 10)

	perm := Perm{}

	if user.Username == f.Owner {
		perm = f.resolvePerms(up)
	} else if user.Username != f.Owner && user.Group == f.Group {
		perm = f.resolvePerms(gp)
	} else if user.Username != f.Owner && user.Group != f.Group {
		perm = f.resolvePerms(op)
	}

	return perm
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
	Root VFSFile `json:"root"`
	Home string  `json:"home"`
	PWD  string  `json:"-"`
	User *User   `json:"-"`
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
	} else if strings.HasPrefix(path, "/home/"+vfs.User.Username) {
		path = strings.Replace(path, "/home/"+vfs.User.Username, "/home/{}", 1)
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
