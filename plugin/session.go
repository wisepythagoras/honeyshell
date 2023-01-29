package plugin

import (
	"golang.org/x/term"
)

type Session struct {
	VFS     *VFS
	Term    *term.Terminal
	Manager *PluginManager
	pwd     string
	User    *User
}

func (s *Session) AutoCompleteCallback(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
	if key == 9 {
		_, matchingCmds := s.Manager.MatchCommand(line)

		for _, cmd := range matchingCmds {
			s.Term.Write([]byte(cmd))
		}

		s.Term.Write([]byte("\n"))
	}

	return line, pos, ok
}

func (s *Session) TermWrite(data ...string) {
	for _, v := range data {
		s.Term.Write([]byte(v))
	}
}

func (s *Session) Chdir(newPath string) error {
	path, _, err := s.VFS.FindFile(newPath)

	if err != nil {
		return err
	}

	s.VFS.PWD = path
	s.pwd = path

	return nil
}

func (s *Session) GetPWD() string {
	return s.pwd
}
