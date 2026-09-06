// Command tailwind baixa o CLI standalone do Tailwind CSS na versão fixada
// abaixo, confere o SHA-256 contra os valores fixados e grava em .tools/.
//
// É a única fonte de verdade da versão do Tailwind no projeto: Taskfile, CI e
// Dockerfile chamam `go run ./tools/tailwind`. Trocar de versão significa
// atualizar a constante e os checksums (publicados em sha256sums.txt de cada
// release), nunca só um dos dois.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const version = "v4.3.3"

// checksums vêm de https://github.com/tailwindlabs/tailwindcss/releases/download/<version>/sha256sums.txt
var checksums = map[string]string{
	"linux-arm64":     "55fd0b241214eff3de1e8ee4f22796662f2d2e7a49bcfca7477cfd0bac398195",
	"linux-x64":       "dc61b3ac6b8c9ca874c0cc4c57b2409791a64c5540404ca5f5367360babc313a",
	"macos-arm64":     "cdf646702987a743464dff4d9c60fd4480d1c1e73dd819a9a67f1078815dce9d",
	"macos-x64":       "7922e0953f2110c05976e3bf58f14e643d90427575e766b7d433f5f80cbee7e1",
	"windows-x64.exe": "e0e260ce048014e9268f6237ff18f8ccf02cef521cbd0ae04e82c2cdf7aa3955",
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "tools/tailwind:", err)
		os.Exit(1)
	}
}

func run() error {
	asset, err := assetSuffix(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	want := checksums[asset]

	dest := filepath.Join(".tools", "tailwindcss")
	if runtime.GOOS == "windows" {
		dest += ".exe"
	}

	if got, err := fileSHA256(dest); err == nil && got == want {
		fmt.Printf("tailwindcss %s já presente em %s\n", version, dest)
		return nil
	}

	url := fmt.Sprintf("https://github.com/tailwindlabs/tailwindcss/releases/download/%s/tailwindcss-%s", version, asset)
	fmt.Printf("baixando tailwindcss %s (%s)\n", version, asset)
	if err := download(url, dest, want); err != nil {
		return err
	}
	fmt.Printf("tailwindcss %s pronto em %s\n", version, dest)
	return nil
}

func assetSuffix(goos, goarch string) (string, error) {
	arch := map[string]string{"amd64": "x64", "arm64": "arm64"}[goarch]
	if arch == "" {
		return "", fmt.Errorf("arquitetura sem build do Tailwind: %s", goarch)
	}
	switch goos {
	case "linux":
		return "linux-" + arch, nil
	case "darwin":
		return "macos-" + arch, nil
	case "windows":
		if arch != "x64" {
			return "", fmt.Errorf("sem build do Tailwind para windows/%s", goarch)
		}
		return "windows-x64.exe", nil
	default:
		return "", fmt.Errorf("sistema sem build do Tailwind: %s", goos)
	}
}

// download grava em arquivo temporário, confere o hash e só então renomeia:
// um download corrompido ou adulterado nunca fica em .tools/.
func download(url, dest, want string) (err error) {
	if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
		return err
	}

	client := &http.Client{Timeout: 3 * time.Minute}
	resp, err := client.Get(url) //nolint:noctx // ferramenta de linha de comando; o timeout do client basta
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, resp.Body.Close()) }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dest), "tailwindcss-*.part")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, hash), resp.Body); err != nil {
		return errors.Join(err, tmp.Close())
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if got := hex.EncodeToString(hash.Sum(nil)); got != want {
		return fmt.Errorf("checksum não confere para %s:\n  esperado %s\n  obtido   %s", url, want, got)
	}
	// É um executável: precisa do bit de execução. Fica em .tools/, fora do git.
	if err := os.Chmod(tmp.Name(), 0o755); err != nil { //nolint:gosec // G302: binário executável
		return err
	}
	return os.Rename(tmp.Name(), dest)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
