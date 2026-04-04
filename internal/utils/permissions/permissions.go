package permissions

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var (
	// Dangerous files that should be protected from reading/editing
	DangerousFiles = []string{
		".bashrc", ".zshrc", ".bash_profile", ".zprofile", ".profile",
		".gitconfig", ".gitmodules", ".mcp.json", ".claude.json",
	}

	// Dangerous directories that should be protected
	DangerousDirectories = []string{
		".git", ".vscode", ".idea", ".claude",
	}

	// Windows short name pattern (e.g., GIT~1)
	shortNameRegex = regexp.MustCompile(`~\d`)

	// Dangerous commands list (simplified)
	DangerousCommands = []string{
		"rm", "fdisk", "format", "mkfs", "dd", "shred",
	}

	// Redirection pattern to detect attempts to write via shell
	// Matches >, >>, 2>, &>, etc. followed by a path
	redirectRegex = regexp.MustCompile(`(?:[12&]?>{1,2})\s*([^\s|;&]+)`)
)

// IsCommandSafe checks if a shell command is safe to execute
func IsCommandSafe(command string, cwd string) (bool, string) {
	lowerCmd := strings.ToLower(command)

	// 1. Detect base command if possible (very simple split for now)
	parts := strings.Fields(command)
	if len(parts) > 0 {
		base := strings.ToLower(filepath.Base(parts[0]))
		for _, dc := range DangerousCommands {
			if base == dc {
				// Special check: rm -rf / or similar broad deletes
				if base == "rm" && (strings.Contains(lowerCmd, "/") || strings.Contains(lowerCmd, "*")) {
					// Disallow deletions of root, parent directories, or broad wildcard deletes
					// Regex to catch: rm -rf /, rm -rf /*, rm -rf ./*, rm -rf ../, rm -rf/, etc.
					// But allow targeted ones like: rm -rf ./node_modules, rm -rf data/
					dangerousPattern := regexp.MustCompile(`\s+(\/|\*|\.\/\*|\.\.\/|\.\/(\s|$))|-[a-z]*[\/\*]`)
					if dangerousPattern.MatchString(lowerCmd) || strings.HasSuffix(lowerCmd, " /") || strings.HasSuffix(lowerCmd, " *") {
						return false, fmt.Sprintf("Command '%s' with broad arguments is restricted for security reasons", base)
					}
				}
			}
		}
	}

	// 2. Detect Redirection to forbidden files
	matches := redirectRegex.FindAllStringSubmatch(command, -1)
	for _, match := range matches {
		if len(match) > 1 {
			targetPath := match[1]
			// Check if targetPath is allowed for writing
			allowed, msg := IsWriteAllowed(targetPath, cwd)
			if !allowed {
				return false, fmt.Sprintf("Restricted redirection detected: %s", msg)
			}
		}
	}

	// 3. Detect suspicious Windows patterns in the command string itself (defense in depth)
	if hasSuspiciousWindowsPattern(command, cwd) {
		return false, "Command contains suspicious Windows path patterns (NTFS streams or shortnames)"
	}

	// 4. Detect forbidden files/directories in the command string with boundary checks
	for _, file := range DangerousFiles {
		// Matches the file name if it stands alone or is part of a path, but not as a substring of another name
		// Example: blocks "cat .bashrc", "ls ./.bashrc", but allows "vim .bashrc.bak"
		pattern := regexp.MustCompile(`(?i)(^|[\s/\\"'|;&])` + regexp.QuoteMeta(file) + `($|[\s/\\"'|;&])`)
		if pattern.MatchString(command) {
			return false, fmt.Sprintf("Command mentions restricted file: %s", file)
		}
	}
	for _, dir := range DangerousDirectories {
		// Example: blocks "ls .git/", "cd .git", allows "git config" or "my.git.repo"
		pattern := regexp.MustCompile(`(?i)(^|[\s/\\"'|;&])` + regexp.QuoteMeta(dir) + `(/|\\|$|[\s"'|;&])`)
		if pattern.MatchString(command) {
			return false, fmt.Sprintf("Command mentions restricted directory: %s", dir)
		}
	}

	return true, ""
}

// IsReadAllowed checks if a file path is safe for reading
func IsReadAllowed(path string, cwd string) (bool, string) {
	// First check raw path for trailing dots/spaces/streams before Abs() cleans it
	if hasSuspiciousWindowsPattern(path, cwd) {
		return false, "Suspicious Windows path pattern detected (NTFS stream, 8.3 shortname, or trailing dot/space)"
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, "Invalid path"
	}

	// Also check absPath for patterns like shortnames and streams
	if hasSuspiciousWindowsPattern(absPath, cwd) {
		return false, "Suspicious Windows path pattern detected in absolute path"
	}

	// Denied Directories check
	if isInsideDangerousDir(absPath) {
		return false, "Access to sensitive directory is denied"
	}

	// Dangerous Files check
	if isDangerousFile(absPath) {
		return false, "Access to sensitive configuration file is denied"
	}

	return true, ""
}

// IsWriteAllowed checks if a file path is safe for writing (CWD restricted)
func IsWriteAllowed(path string, cwd string) (bool, string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, "Invalid path"
	}

	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return false, "Invalid CWD"
	}

	// Restriction: Must be inside CWD
	normPath := filepath.ToSlash(strings.ToLower(absPath))
	normCwd := filepath.ToSlash(strings.ToLower(absCwd))
	if !strings.HasPrefix(normPath, normCwd) {
		return false, "Writing outside current project directory is restricted"
	}

	// Also perform read-level safety checks (denied files/dirs)
	allowed, msg := IsReadAllowed(path, cwd)
	if !allowed {
		return false, msg
	}

	return true, ""
}

func isInsideDangerousDir(path string) bool {
	// Use ToSlash for consistent segment splitting
	normalizedPath := filepath.ToSlash(path)
	segments := strings.Split(normalizedPath, "/")
	for _, segment := range segments {
		lowerSegment := strings.ToLower(segment)
		for _, dir := range DangerousDirectories {
			if lowerSegment == strings.ToLower(dir) {
				return true
			}
		}
	}
	return false
}

func isDangerousFile(path string) bool {
	filename := filepath.Base(path)
	lowerFilename := strings.ToLower(filename)
	for _, file := range DangerousFiles {
		if lowerFilename == strings.ToLower(file) {
			return true
		}
	}
	return false
}

func hasSuspiciousWindowsPattern(path string, cwd string) bool {
	// Standardize for consistent pattern matching across platforms
	normPath := filepath.ToSlash(strings.ToLower(path))

	// 1. Check for trailing dots or spaces - dangerous/confusing on many platforms
	if strings.HasSuffix(normPath, ".") || strings.HasSuffix(normPath, " ") {
		return true
	}

	// 2. Check for NTFS Alternate Data Streams (typically contains ':')
	// Even on Unix, ':' in a filename is rare/suspicious for an agent to be accessing.
	if len(normPath) > 3 {
		// On Windows paths like C:/, the ':' at index 1 is allowed
		remaining := normPath
		if len(normPath) >= 2 && normPath[1] == ':' {
			remaining = normPath[2:]
		}
		if strings.Contains(remaining, ":") {
			return true
		}
	}

	// 3. Check for 8.3 short names (e.g., ~1)
	if shortNameRegex.MatchString(normPath) {
		// Only whitelist system shortnames if we are actually on Windows.
		// On non-Windows, ~N in a filename is just suspicious for security purposes.
		if runtime.GOOS == "windows" {
			// Check against CWD
			if cwd != "" {
				absCwd, _ := filepath.Abs(cwd)
				normCwd := filepath.ToSlash(strings.ToLower(absCwd))
				if strings.Contains(normPath, normCwd) {
					return false
				}
			}

			// Check against TEMP directory
			tempDir := filepath.ToSlash(strings.ToLower(os.TempDir()))
			if strings.Contains(normPath, tempDir) {
				return false
			}

			// Check against USERPROFILE directory
			userProfile := filepath.ToSlash(strings.ToLower(os.Getenv("USERPROFILE")))
			if userProfile != "" && strings.Contains(normPath, userProfile) {
				return false
			}
		}

		return true
	}

	return false
}
