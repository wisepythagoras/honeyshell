package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/wisepythagoras/honeyshell/plugin"
)

func readDir(path, basePath string) (map[string]plugin.VFSFile, error) {
	files, err := os.ReadDir(path)

	if err != nil {
		return nil, err
	}

	vfsMap := make(map[string]plugin.VFSFile)

	for _, f := range files {
		info, err := f.Info()

		if err != nil {
			return nil, err
		}

		stat := info.Sys().(*syscall.Stat_t)

		uid := stat.Uid
		gid := stat.Gid
		u := strconv.FormatUint(uint64(uid), 10)
		g := strconv.FormatUint(uint64(gid), 10)

		usr, err := user.LookupId(u)

		if err != nil {
			return nil, err
		}

		group, err := user.LookupGroupId(g)

		if err != nil {
			return nil, err
		}

		fmt.Println(filepath.Join(path, f.Name()), usr.Username, group.Name, f.IsDir())

		vfsFile := plugin.VFSFile{
			Name:    f.Name(),
			Owner:   usr.Username,
			ModTime: info.ModTime(),
			Group:   group.Name,
			NLink:   int(stat.Nlink),
		}

		newFilePath := filepath.Join(path, f.Name())
		shouldInclude := !strings.HasSuffix(newFilePath, "/bin") &&
			!strings.HasSuffix(newFilePath, "/lib") &&
			!strings.HasSuffix(newFilePath, "/lib64") &&
			!strings.HasSuffix(newFilePath, "/dev") &&
			!strings.HasSuffix(newFilePath, "/sys")

		if !f.IsDir() && (info.Mode()&os.ModeSymlink == 0) {
			buff, err := os.ReadFile(filepath.Join(path, f.Name()))

			if err != nil {
				return nil, err
			}

			vfsFile.Type = plugin.T_FILE
			vfsFile.Contents = string(buff)
			vfsFile.Mode = fs.FileMode(stat.Mode)
		} else if info.Mode()&os.ModeSymlink != 0 {
			flPath, err := os.Readlink(newFilePath)
			flPath = strings.Replace(flPath, basePath, "", 1)
			fmt.Println(newFilePath, "->", flPath, err)

			vfsFile.Type = plugin.T_SYMLINK
			vfsFile.Mode = fs.FileMode(stat.Mode) | os.ModeSymlink
			vfsFile.LinkTo = flPath
		} else if shouldInclude {
			vfsFiles, err := traverseDir(f, newFilePath, basePath)

			if err != nil {
				return nil, err
			}

			vfsFile.Files = vfsFiles
			vfsFile.Type = plugin.T_DIR
			vfsFile.Mode = fs.FileMode(stat.Mode) | os.ModeDir
		}

		vfsMap[f.Name()] = vfsFile
	}

	return vfsMap, nil
}

func traverseDir(file os.DirEntry, path, basePath string) (map[string]plugin.VFSFile, error) {
	if !file.IsDir() {
		return nil, fmt.Errorf("%q not a directory", path)
	}

	return readDir(path, basePath)
}

func main() {
	path := flag.String("path", "", "The path to the directory structure to clone")
	home := flag.String("home", "/home/{}", "Specify the home directory (add '{}' in place of the username)")
	out := flag.String("out", "out.json", "Where to write the VFS to")

	flag.Parse()

	if len(*path) == 0 {
		fmt.Println("A path is required")
		os.Exit(1)
	}

	files, err := readDir(*path, *path)

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	vfsRoot := plugin.VFSFile{
		Type:  plugin.T_DIR,
		Name:  "",
		Mode:  0755,
		Owner: "root",
		Files: files,
	}
	vfs := plugin.VFS{
		Root: vfsRoot,
		Home: *home,
		User: &plugin.User{
			Username: "{}",
			Group:    "{}",
		},
	}

	_, homeFolder, err := vfs.FindFile("/home")

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(3)
	}

	homeFolder.Files["{}"] = plugin.VFSFile{
		Type:  plugin.T_DIR,
		Name:  "{}",
		Owner: "{}",
		Group: "{}",
		Mode:  0775 | os.ModeDir,
		Files: make(map[string]plugin.VFSFile),
	}

	j, err := json.Marshal(vfs)

	if err != nil {
		fmt.Println("JSON Error:", err)
		os.Exit(4)
	}

	if err = os.WriteFile(*out, j, 0770); err != nil {
		fmt.Println(err)
	}
}
