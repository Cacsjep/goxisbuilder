package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

type ProgressBar struct {
	current   int
	total     int
	prefix    string
	spinner   []string
	spinIndex int
	mu        sync.Mutex // Protects access to current and spinIndex
	done      bool
}

func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		total:   total,
		spinner: []string{"-", "/", "|", "\\"},
		prefix:  "Processing...",
	}
}

func (p *ProgressBar) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current < p.total {
		p.current++
	}
}

func (p *ProgressBar) SetTotal(total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.total = total
}

func (p *ProgressBar) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = p.total
	p.prefix = "Complete"
	p.done = true
}

func (p *ProgressBar) SetPrefix(prefix string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.prefix = prefix
}

func (p *ProgressBar) Render() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.total <= 0 {
		fmt.Printf("\r%s Total must be greater than 0 to calculate progress.\n", p.prefix)
		return
	}

	percent := float64(p.current) / float64(p.total) * 100
	bar := "\033[34m"
	for i := 0; i < p.total; i++ {
		if i < p.current {
			bar += "="
		} else {
			bar += " "
		}
	}
	bar += "\033[0m"
	spinChar := p.spinner[p.spinIndex%len(p.spinner)]

	maxLineLength := getTerminalWidth()
	progress := fmt.Sprintf("%s %s [%s] %d%% Steps: \033[34m%d\033[0m/%d", spinChar, p.prefix, bar, int(percent), p.current, p.total)
	padding := strings.Repeat(" ", max(maxLineLength-len(progress), 0))

	fmt.Printf("\r%s%s", progress, padding)
}

func (p *ProgressBar) StartSpinner() {
	go func() {
		for {
			p.mu.Lock()
			if p.done {
				p.mu.Unlock()
				return
			}
			p.spinIndex++
			p.mu.Unlock()

			p.Render()                         // Update progress bar to show spinner frame
			time.Sleep(100 * time.Millisecond) // Adjust the speed as needed
		}
	}()
}

// Helper functions remain unchanged
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 150 // Fallback to a default value if there's an error
	}
	return width
}
