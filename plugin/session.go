package plugin

import (
	"golang.org/x/term"
)

type Session struct {
	VFS      *VFS
	Username string
	Term     *term.Terminal
	Manager  *PluginManager
	pwd      string
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

func (s *Session) Chdir(newPath string) {
	s.VFS.PWD = newPath
	s.pwd = newPath
}

func (s *Session) GetPWD() string {
	return s.pwd
}
