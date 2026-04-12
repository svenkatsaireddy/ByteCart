package promo

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"unicode"
)

// Validator checks promo codes: length 8–10 and substring present in ≥2 corpora.
// Large gzip files are not fully loaded at startup; each new code streams the files once and results are cached.
type Validator struct {
	paths   []string   // gzip paths (production)
	memFile []string   // raw text (tests)
	mu      sync.RWMutex
	cache   map[string]bool
}

// NewValidatorFromGzipFiles checks that paths exist and returns a streaming validator.
func NewValidatorFromGzipFiles(paths []string) (*Validator, error) {
	if len(paths) != 3 {
		return nil, fmt.Errorf("expected 3 coupon gzip paths, got %d", len(paths))
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			return nil, fmt.Errorf("coupon file %s: %w", p, err)
		}
	}
	return &Validator{
		paths: paths,
		cache: make(map[string]bool),
	}, nil
}

// NewValidatorFromStringContents builds a validator from three in-memory texts (tests).
func NewValidatorFromStringContents(files []string) (*Validator, error) {
	if len(files) != 3 {
		return nil, fmt.Errorf("expected 3 files, got %d", len(files))
	}
	return &Validator{
		memFile: files,
		cache:   make(map[string]bool),
	}, nil
}

// Valid reports whether code is allowed (length 8–10 and substring in ≥2 files).
func (v *Validator) Valid(code string) bool {
	if v == nil {
		return false
	}
	if len(v.memFile) == 0 && len(v.paths) == 0 {
		return false
	}
	n := len(code)
	if n < 8 || n > 10 {
		return false
	}
	if !isAlnum(code) {
		return false
	}

	v.mu.RLock()
	if res, ok := v.cache[code]; ok {
		v.mu.RUnlock()
		return res
	}
	v.mu.RUnlock()

	cnt := v.countInFiles(code)
	res := cnt >= 2

	v.mu.Lock()
	v.cache[code] = res
	v.mu.Unlock()
	return res
}

func (v *Validator) countInFiles(code string) int {
	if len(v.memFile) > 0 {
		n := 0
		for _, s := range v.memFile {
			if strings.Contains(s, code) {
				n++
				if n >= 2 {
					return n
				}
			}
		}
		return n
	}
	n := 0
	for _, p := range v.paths {
		ok, err := gzipContainsSubstring(p, code)
		if err != nil || !ok {
			continue
		}
		n++
		if n >= 2 {
			return n
		}
	}
	return n
}

func gzipContainsSubstring(path, needle string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return false, err
	}
	defer zr.Close()
	return readerContains(zr, []byte(needle))
}

func readerContains(r io.Reader, needle []byte) (bool, error) {
	if len(needle) == 0 {
		return false, nil
	}
	const chunkSize = 256 * 1024
	buf := make([]byte, chunkSize)
	overlap := len(needle) - 1
	if overlap < 0 {
		overlap = 0
	}
	var carry []byte
	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			data := append(carry, chunk...)
			if bytes.Contains(data, needle) {
				return true, nil
			}
			if len(data) > overlap {
				carry = append(carry[:0], data[len(data)-overlap:]...)
			} else {
				carry = append(carry[:0], data...)
			}
		}
		if err == io.EOF {
			return false, nil
		}
		if err != nil {
			return false, err
		}
	}
}

func isAlnum(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
