package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var errSelectCancelled = errors.New("selection cancelled")

const (
	escReset   = "\033[0m"
	escBold    = "\033[1m"
	escDim     = "\033[2m"
	escCyan    = "\033[36m"
	escGreen   = "\033[32m"
	escHideCur = "\033[?25l"
	escShowCur = "\033[?25h"
)

// multiSelect menampilkan checklist interaktif di terminal dan mengembalikan
// index item yang dipilih. Ketik untuk mencari (dicocokkan ke searchKeys),
// spasi untuk memilih.
func multiSelect(title string, items, searchKeys []string) ([]int, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer term.Restore(fd, oldState)

	s := &selector{title: title, items: items, keys: searchKeys, checked: make([]bool, len(items))}
	s.applyFilter()

	fmt.Print(escHideCur)
	defer fmt.Print(escShowCur)

	buf := make([]byte, 64)
	for {
		s.render()

		n, err := os.Stdin.Read(buf)
		if err != nil {
			s.clear()
			return nil, err
		}

		done, cancelled := s.handleKey(buf[:n])
		if cancelled {
			s.clear()
			return nil, errSelectCancelled
		}
		if done {
			s.clear()
			var picked []int
			for i, c := range s.checked {
				if c {
					picked = append(picked, i)
				}
			}
			return picked, nil
		}
	}
}

type selector struct {
	title    string
	items    []string
	keys     []string
	checked  []bool
	filter   string
	visible  []int
	cursor   int
	offset   int
	rendered int
}

func (s *selector) applyFilter() {
	s.visible = s.visible[:0]
	q := strings.ToLower(s.filter)
	for i, key := range s.keys {
		if q == "" || strings.Contains(strings.ToLower(key), q) {
			s.visible = append(s.visible, i)
		}
	}
	s.cursor, s.offset = 0, 0
}

func (s *selector) move(delta int) {
	if len(s.visible) == 0 {
		return
	}
	s.cursor += delta
	if s.cursor < 0 {
		s.cursor = 0
	}
	if s.cursor >= len(s.visible) {
		s.cursor = len(s.visible) - 1
	}
}

func (s *selector) toggleAllVisible() {
	allChecked := true
	for _, i := range s.visible {
		if !s.checked[i] {
			allChecked = false
			break
		}
	}
	for _, i := range s.visible {
		s.checked[i] = !allChecked
	}
}

func (s *selector) handleKey(key []byte) (done, cancelled bool) {
	page := s.pageSize()

	switch string(key) {
	case "\x03", "\x1b": // ctrl+c, esc
		return false, true
	case "\r", "\n":
		return true, false
	case "\x1b[A", "\x1bOA":
		s.move(-1)
	case "\x1b[B", "\x1bOB":
		s.move(1)
	case "\x1b[5~":
		s.move(-page)
	case "\x1b[6~":
		s.move(page)
	case "\x1b[H", "\x1bOH":
		s.move(-len(s.visible))
	case "\x1b[F", "\x1bOF":
		s.move(len(s.visible))
	case " ":
		if len(s.visible) > 0 {
			i := s.visible[s.cursor]
			s.checked[i] = !s.checked[i]
		}
	case "\t":
		if len(s.visible) > 0 {
			i := s.visible[s.cursor]
			s.checked[i] = !s.checked[i]
			s.move(1)
		}
	case "\x01": // ctrl+a
		s.toggleAllVisible()
	case "\x7f", "\x08": // backspace
		if s.filter != "" {
			s.filter = s.filter[:len(s.filter)-1]
			s.applyFilter()
		}
	case "\x15": // ctrl+u
		s.filter = ""
		s.applyFilter()
	default:
		if key[0] == 0x1b {
			return false, false
		}
		changed := false
		for _, b := range key {
			if b > ' ' && b < 0x7f {
				s.filter += string(b)
				changed = true
			}
		}
		if changed {
			s.applyFilter()
		}
	}
	return false, false
}

func (s *selector) termSize() (width, height int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 24
	}
	return w, h
}

func (s *selector) pageSize() int {
	_, h := s.termSize()
	size := h - 5
	if size < 3 {
		size = 3
	}
	return size
}

func (s *selector) clear() {
	if s.rendered > 0 {
		fmt.Printf("\033[%dA", s.rendered)
	}
	fmt.Print("\r\033[J")
	s.rendered = 0
}

func (s *selector) render() {
	width, _ := s.termSize()
	page := s.pageSize()

	if s.cursor < s.offset {
		s.offset = s.cursor
	}
	if s.cursor >= s.offset+page {
		s.offset = s.cursor - page + 1
	}

	selected := 0
	for _, c := range s.checked {
		if c {
			selected++
		}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("%s%s%s %s(%d dipilih)%s", escBold, s.title, escReset, escGreen, selected, escReset))
	lines = append(lines, fmt.Sprintf("%sCari:%s %s", escCyan, escReset, s.filter))

	if len(s.visible) == 0 {
		lines = append(lines, escDim+"  (tidak ada user yang cocok)"+escReset)
	}

	end := s.offset + page
	if end > len(s.visible) {
		end = len(s.visible)
	}
	for pos := s.offset; pos < end; pos++ {
		i := s.visible[pos]
		pointer, box := "  ", "[ ]"
		if s.checked[i] {
			box = escGreen + "[x]" + escReset
		}
		label := truncate(s.items[i], width-8)
		if pos == s.cursor {
			pointer = escCyan + "❯ " + escReset
			label = escBold + label + escReset
		}
		lines = append(lines, fmt.Sprintf("%s%s %s", pointer, box, label))
	}

	lines = append(lines, fmt.Sprintf("%s%d-%d dari %d · ↑/↓ pindah · spasi pilih · ctrl+a pilih semua · ketik untuk cari · enter lanjut · esc batal%s",
		escDim, min(s.offset+1, end), end, len(s.visible), escReset))

	s.clear()
	// Mode raw tidak menerjemahkan \n menjadi \r\n, jadi tulis \r\n manual.
	for _, l := range lines {
		fmt.Print(truncateANSIWidth(l, width), "\r\n")
	}
	s.rendered = len(lines)
}

func truncate(text string, max int) string {
	r := []rune(text)
	if max < 1 || len(r) <= max {
		return text
	}
	return string(r[:max-1]) + "…"
}

// truncateANSIWidth memotong baris yang lebih lebar dari terminal agar tidak
// wrap (wrap membuat perhitungan clear() meleset). Escape ANSI tidak dihitung.
func truncateANSIWidth(line string, width int) string {
	if width < 1 {
		return line
	}
	var b strings.Builder
	visible, inEsc := 0, false
	for _, r := range line {
		if r == '\033' {
			inEsc = true
		}
		if inEsc {
			b.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		if visible >= width-1 {
			b.WriteString(escReset)
			break
		}
		b.WriteRune(r)
		visible++
	}
	return b.String()
}
