/*
 * Copyright The Project-HAMi Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gcu

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// hostPath joins hostRoot with an absolute-looking path, treating path as
// relative to hostRoot even when it starts with "/".
func hostPath(hostRoot, path string) string {
	if hostRoot == "" || hostRoot == "/" {
		return path
	}
	return filepath.Join(hostRoot, strings.TrimPrefix(path, "/"))
}

func readTrimmed(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func readTrimmedOptional(path string) string {
	v, err := readTrimmed(path)
	if err != nil {
		return ""
	}
	return v
}

// parseDevID decodes sysfs dev_info/dev_id into a UUID string.
// The file contains space-separated decimal byte values that form a
// NUL-padded ASCII serial (e.g. "84 82 54 ..." → "TR6W18220404").
func parseDevID(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty dev_id")
	}

	// Already a printable serial (tests / alternate kernels).
	if !strings.Contains(raw, " ") && isPrintableASCII(raw) {
		return raw, nil
	}

	fields := strings.Fields(raw)
	buf := make([]byte, 0, len(fields))
	for _, f := range fields {
		n, err := strconv.ParseUint(f, 10, 8)
		if err != nil {
			return "", fmt.Errorf("parse dev_id byte %q: %w", f, err)
		}
		buf = append(buf, byte(n))
	}
	uuid := string(bytes.TrimRight(buf, "\x00"))
	if uuid == "" {
		return "", fmt.Errorf("dev_id decoded to empty string")
	}
	if !isPrintableASCII(uuid) {
		return "", fmt.Errorf("dev_id decoded to non-printable value %q", uuid)
	}
	return uuid, nil
}

func isPrintableASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 || s[i] > 0x7e {
			return false
		}
	}
	return true
}

// parseMemoryBytes parses sysfs hbm_size (hex like 0xbfe000000 or decimal).
func parseMemoryBytes(raw string) (uint64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("empty hbm_size")
	}
	base := 10
	if strings.HasPrefix(strings.ToLower(raw), "0x") {
		base = 0 // let ParseUint detect 0x
	}
	n, err := strconv.ParseUint(raw, base, 64)
	if err != nil {
		return 0, fmt.Errorf("parse hbm_size %q: %w", raw, err)
	}
	return n, nil
}

func parseNUMANode(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return -1
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return -1
	}
	return n
}

// findPCIeRoot walks up from a sysfs device directory looking for a pciXXXX:YY
// root-complex directory name.
func findPCIeRoot(deviceDir string) string {
	cur := deviceDir
	for i := 0; i < 16 && cur != "" && cur != "/" && cur != "."; i++ {
		base := filepath.Base(cur)
		if strings.HasPrefix(base, "pci") && strings.Contains(base, ":") {
			return base
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return ""
}

func parseFirmwareVersion(raw string) string {
	raw = strings.TrimSpace(raw)
	const prefix = "firmware version:"
	if strings.HasPrefix(strings.ToLower(raw), prefix) {
		return strings.TrimSpace(raw[len(prefix):])
	}
	return raw
}

// parseGCUIndex extracts N from names like "gcu0". Returns false for
// "gcu0vid0", "gcuctl", or other non-matching entries.
func parseGCUIndex(name string) (int, bool) {
	if !strings.HasPrefix(name, "gcu") {
		return 0, false
	}
	rest := name[len("gcu"):]
	if rest == "" {
		return 0, false
	}
	for i := 0; i < len(rest); i++ {
		if rest[i] < '0' || rest[i] > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(rest)
	if err != nil {
		return 0, false
	}
	return n, true
}
