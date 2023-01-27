package plugin_test

import (
	"encoding/json"
	"testing"

	"github.com/wisepythagoras/honeyshell/plugin"
)

const testVfs = "{\"root\":{\"type\":1,\"name\":\"\",\"files\":{\"home\":{\"type\":1,\"name\":\"home\",\"files\":{\"{}\":{\"type\":1,\"name\":\"{}\",\"files\":{\"test.txt\":{\"type\":2,\"name\":\"test.txt\",\"owner\":\"{}\",\"contents\":\"This is a test file\",\"mode\":432}},\"owner\":\"{}\",\"mode\":2147484157}},\"owner\":\"root\",\"mode\":2147484141},\"etc\":{\"type\":1,\"name\":\"etc\",\"files\":{\"hostname\":{\"type\":2,\"name\":\"hostname\",\"owner\":\"root\",\"contents\":\"test-hostname\",\"mode\":420},\"issue\":{\"type\":2,\"name\":\"issue\",\"owner\":\"root\",\"contents\":\"Ubuntu 22.04\",\"mode\":420}},\"owner\":\"root\",\"mode\":2147484141}},\"owner\":\"root\",\"mode\":2147484141},\"home\":\"/home/{}\"}"

func TestPathResolution(t *testing.T) {
	vfs := &plugin.VFS{}
	err := json.Unmarshal([]byte(testVfs), vfs)
	vfs.Username = "{}"

	if err != nil {
		t.Errorf("Error: %s", err)
	}

	p, f, err := vfs.FindFile("/etc/issue")

	if err != nil {
		t.Errorf("Error: %s", err)
	}

	if p != "/etc/issue" {
		t.Errorf("Invalid path %q", p)
	}

	if f.Contents != "Ubuntu 22.04" {
		t.Errorf("Invalid contents %q", f.Contents)
	}

	_, _, err = vfs.FindFile("/home/{}/file-that-doesnt-exist.txt")

	if err == nil {
		t.Error("File found when there should be one")
	}

	_, _, err = vfs.FindFile("/home/{}/")

	if err != nil {
		t.Error("Error:", err)
	}

	_, _, err = vfs.FindFile("/")

	if err != nil {
		t.Error("Error:", err)
	}
}

func TestChdir(t *testing.T) {
	vfs := &plugin.VFS{}
	err := json.Unmarshal([]byte(testVfs), vfs)
	vfs.Username = "{}"

	if err != nil {
		t.Errorf("Error: %s", err)
	}

	session := plugin.Session{
		VFS:      vfs,
		Username: "test",
	}
	session.Chdir(vfs.Home)

	pwd := session.GetPWD()

	if pwd != vfs.Home {
		t.Errorf("%q != %q", pwd, vfs.Home)
	}

	p, f, err := vfs.FindFile("test.txt")

	if err != nil {
		t.Errorf("Error: %s", err)
	}

	if p != "/home/{}/test.txt" {
		t.Errorf("%q != \"/home/{}/test.txt\"", p)
	}

	if f.Contents != "This is a test file" {
		t.Errorf("%q != \"This is a test file\"", f.Contents)
	}

	p, f, err = vfs.FindFile("../../../etc/hostname")

	if err != nil {
		t.Errorf("Error: %s", err)
	}

	if p != "/etc/hostname" {
		t.Errorf("%q != \"/etc/hostname\"", p)
	}

	if f.Contents != "test-hostname" {
		t.Errorf("%q != \"test-hostname\"", f.Contents)
	}

	session.Chdir("/")
	p, f, err = vfs.FindFile("~/test.txt")

	if err != nil {
		t.Errorf("Error: %s", err)
	}

	if p != "/home/{}/test.txt" {
		t.Errorf("%q != \"/home/{}/test.txt\"", p)
	}

	if f.Contents != "This is a test file" {
		t.Errorf("%q != \"This is a test file\"", f.Contents)
	}

	_, _, err = vfs.FindFile("~/doesnt-exist")

	if err == nil {
		t.Error("File found where there should't be one")
	}
}
