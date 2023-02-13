package plugin_test

import (
	"encoding/json"
	"testing"

	"github.com/wisepythagoras/honeyshell/plugin"
)

const testVfs = "{\"root\":{\"t\":1,\"n\":\"\",\"f\":{\"home\":{\"t\":1,\"n\":\"home\",\"f\":{\"{}\":{\"t\":1,\"n\":\"{}\",\"f\":{\"test.txt\":{\"t\":2,\"n\":\"test.txt\",\"o\":\"{}\",\"c\":\"This is a test file\",\"m\":432}},\"o\":\"{}\",\"m\":2147484157}},\"o\":\"root\",\"m\":2147484141},\"etc\":{\"t\":1,\"n\":\"etc\",\"f\":{\"hostname\":{\"t\":2,\"n\":\"hostname\",\"o\":\"root\",\"c\":\"test-hostname\",\"m\":420},\"issue\":{\"t\":2,\"n\":\"issue\",\"o\":\"root\",\"c\":\"Ubuntu 22.04\",\"m\":420}},\"o\":\"root\",\"m\":2147484141}},\"o\":\"root\",\"m\":2147484141},\"home\":\"/home/{}\"}"

func TestPathResolution(t *testing.T) {
	vfs := &plugin.VFS{}
	err := json.Unmarshal([]byte(testVfs), vfs)
	vfs.User = &plugin.User{
		Username: "{}",
		Group:    "{}",
	}

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
	vfs.User = &plugin.User{
		Username: "{}",
		Group:    "{}",
	}

	if err != nil {
		t.Errorf("Error: %s", err)
	}

	session := plugin.Session{
		VFS: vfs,
		User: &plugin.User{
			Username: "test",
			Group:    "test",
		},
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
