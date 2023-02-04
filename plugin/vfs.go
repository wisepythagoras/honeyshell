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
	Type     int                `json:"t"`
	Name     string             `json:"n"`
	Files    map[string]VFSFile `json:"f"`
	Contents string             `json:"c"`
	Mode     os.FileMode        `json:"m"`
	Owner    string             `json:"o"`
	Group    string             `json:"g"`
	ModTime  time.Time          `json:"mt"`
	LinkTo   string             `json:"lt"`
	NLink    int                `json:"nl"`
	CmdFn    CommandFn          `json:"-"`
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

func (f *VFSFile) CanAccess(user *User) Perm {
	var buf [32]bool
	w := 0

	// 13, one for each of "dalTLDpSugct?"
	for i := 0; i < 13; i++ {
		if f.Mode&(1<<uint(32-1-i)) != 0 {
			buf[w] = true
			w++
		}
	}

	if w == 0 {
		buf[w] = false
		w++
	}

	// 9, one for each of "rwxrwxrwx"
	for i := 0; i < 9; i++ {
		if f.Mode&(1<<uint(9-1-i)) != 0 {
			buf[w] = true
		} else {
			buf[w] = false
		}

		w++
	}

	rawPerms := buf[:w]
	perm := Perm{}

	if user.Username == f.Owner {
		perm.Read = rawPerms[1]
		perm.Write = rawPerms[2]
		perm.Exec = rawPerms[3]
	} else if user.Username != f.Owner && user.Group == f.Group {
		perm.Read = rawPerms[4]
		perm.Write = rawPerms[5]
		perm.Exec = rawPerms[6]
	} else if user.Username != f.Owner && user.Group != f.Group {
		perm.Read = rawPerms[7]
		perm.Write = rawPerms[8]
		perm.Exec = rawPerms[9]
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

func (vfs *VFS) Mkdir(path string, mode os.FileMode) (*VFSFile, error) {
	_, file, err := vfs.FindFile(filepath.Dir(path))

	if err != nil {
		return nil, err
	}

	if file.Type != T_DIR {
		return nil, fmt.Errorf("cannot create directory ‘%s‘: No such file or directory", path)
	}

	realUser := *vfs.User

	if vfs.User.Username == realUser.Username {
		realUser = User{
			Username: "{}",
			Group:    "{}",
		}
	}

	perms := file.CanAccess(&realUser)

	if !perms.Write {
		return nil, fmt.Errorf("cannot create directory ‘%s‘: Permission denied", path)
	}

	base := filepath.Base(path)

	if _, ok := file.Files[base]; ok {
		return nil, fmt.Errorf("file %q already exists", path)
	}

	if mode == 0 {
		mode = 0775
	}

	newFile := VFSFile{
		Name:    base,
		Type:    T_DIR,
		Mode:    mode | os.ModeDir,
		Files:   make(map[string]VFSFile),
		Owner:   "{}",
		Group:   "{}",
		ModTime: time.Now(),
	}

	file.Files[base] = newFile

	return &newFile, nil
}

func (vfs *VFS) Rmfile(path string) error {
	_, parentFolder, err := vfs.FindFile(filepath.Dir(path))

	if err != nil {
		return err
	}

	realUser := *vfs.User

	if vfs.User.Username == realUser.Username {
		realUser = User{
			Username: "{}",
			Group:    "{}",
		}
	}

	var file VFSFile
	var ok bool

	base := filepath.Base(path)

	if file, ok = parentFolder.Files[base]; !ok {
		return fmt.Errorf("no such file or directory")
	}

	perms := file.CanAccess(&realUser)

	if !perms.Write {
		return fmt.Errorf("permission denied")
	}

	delete(parentFolder.Files, base)

	return nil
}

func (vfs *VFS) WriteFile(path, contents string) error {
	_, parentFolder, err := vfs.FindFile(filepath.Dir(path))

	if err != nil {
		return err
	}

	realUser := *vfs.User

	if vfs.User.Username == realUser.Username {
		realUser = User{
			Username: "{}",
			Group:    "{}",
		}
	}

	base := filepath.Base(path)
	perms := parentFolder.CanAccess(&realUser)

	if !perms.Write {
		return fmt.Errorf("permission denied")
	}

	file, ok := parentFolder.Files[base]

	if ok && file.Type == T_DIR {
		return fmt.Errorf("file is a directory")
	} else if ok && !file.CanAccess(&realUser).Write {
		return fmt.Errorf("permission denied")
	}

	// TODO: Implement all modes, read, write, append.
	if ok {
		file.Contents = contents
	} else {
		file = VFSFile{
			Type:     T_FILE,
			Mode:     0664,
			Name:     base,
			Contents: contents,
			Owner:    "{}",
			Group:    "{}",
			ModTime:  time.Now(),
		}
	}

	parentFolder.Files[base] = file

	return nil
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
