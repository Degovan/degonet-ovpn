package output

import (
	"fmt"
	"strings"
)

type Row []string

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

func Header(text string) {
	fmt.Printf("\n%s%s%s\n", colorBold, text, colorReset)
	fmt.Println(strings.Repeat("─", len(text)))
}

func Success(text string) {
	fmt.Printf("\n%s✓ %s%s\n", colorGreen, text, colorReset)
}

func Error(text string) {
	fmt.Printf("%s✗ %s%s\n", colorRed, text, colorReset)
}

func Warning(text string) {
	fmt.Printf("%s⚠ %s%s\n", colorYellow, text, colorReset)
}

func Table(rows []Row) {
	if len(rows) == 0 {
		return
	}
	keyWidth := 0
	for _, r := range rows {
		if len(r) > 0 && len(r[0]) > keyWidth {
			keyWidth = len(r[0])
		}
	}
	for _, r := range rows {
		if len(r) >= 2 {
			fmt.Printf("  %s%-*s%s : %s\n", colorCyan, keyWidth, r[0], colorReset, r[1])
		}
	}
}

func TableWithHeaders(headers []string, rows []Row) {
	if len(headers) == 0 || len(rows) == 0 {
		return
	}

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}
	for _, r := range rows {
		for i, cell := range r {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	printRow := func(row []string, bold bool) {
		parts := make([]string, 0, len(row))
		for i, cell := range row {
			if i < len(colWidths) {
				parts = append(parts, fmt.Sprintf("%-*s", colWidths[i], cell))
			}
		}
		line := "  " + strings.Join(parts, "  ")
		if bold {
			fmt.Printf("%s%s%s\n", colorBold, line, colorReset)
		} else {
			fmt.Println(line)
		}
	}

	sep := strings.Repeat("─", sum(colWidths)+len(colWidths)*2-2)
	fmt.Println("  " + sep)
	printRow(headers, true)
	fmt.Println("  " + sep)
	for _, r := range rows {
		printRow(r, false)
	}
	fmt.Println("  " + sep)
}

func sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}
