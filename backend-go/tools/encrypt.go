package main

import (
	"fmt"
	"os"
	"stock-monitor/pkg/envcrypt"
)

// Build-time passphrase used for env file encryption.
// This is obfuscated by splitting into multiple literals.
func passphrase() string {
	return "Inv" + "est" + "ment" + "Asst" + "2026" + "Secure"
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run tools/encrypt.go <path/to/.env>")
		os.Exit(1)
	}
	inPath := os.Args[1]
	outPath := inPath + ".enc"

	plain, err := os.ReadFile(inPath)
	if err != nil {
		fmt.Println("read failed:", err)
		os.Exit(1)
	}

	enc, err := envcrypt.Encrypt(plain, passphrase())
	if err != nil {
		fmt.Println("encrypt failed:", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, enc, 0644); err != nil {
		fmt.Println("write failed:", err)
		os.Exit(1)
	}
	fmt.Println("Encrypted:", inPath, "->", outPath)
}
