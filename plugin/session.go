package plugin

import (
	"golang.org/x/term"
)

type Session struct {
	VFS      *VFS
	Username string
	Term     *term.Terminal
	Manager  *PluginManager
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
