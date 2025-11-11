package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var exitFunc = os.Exit

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		exitFunc(1)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: coverage <file>")
	}

	coverage, err := calculate(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("coverage: %.1f%%\n", coverage)
	return nil
}

func calculate(path string) (float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open coverage file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var covered, total float64

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		statements, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return 0, fmt.Errorf("parse statements: %w", err)
		}
		count, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return 0, fmt.Errorf("parse count: %w", err)
		}
		total += statements
		if count > 0 {
			covered += statements
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scan coverage: %w", err)
	}

	if total == 0 {
		return 0, nil
	}
	return (covered / total) * 100, nil
}
