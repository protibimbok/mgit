package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func Ask(label, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

func Select(label string, options []string) (int, error) {
	fmt.Printf("%s:\n", label)
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}
	for {
		fmt.Printf("Enter choice [1-%d]: ", len(options))
		input, err := reader.ReadString('\n')
		if err != nil {
			return 0, err
		}
		var idx int
		if _, err := fmt.Sscanf(strings.TrimSpace(input), "%d", &idx); err == nil {
			if idx >= 1 && idx <= len(options) {
				return idx - 1, nil
			}
		}
		fmt.Println("Invalid choice, try again.")
	}
}

func Confirm(label string) (bool, error) {
	fmt.Printf("%s [y/N]: ", label)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes", nil
}
