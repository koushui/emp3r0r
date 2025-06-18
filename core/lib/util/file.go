package util

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jm33-m0/emp3r0r/core/lib/logging"
)

// Dentry Directory entry
type Dentry struct {
	Name       string `json:"name"`  // filename
	Ftype      string `json:"ftype"` // file/dir
	Size       string `json:"size"`  // 100
	Date       string `json:"date"`  // 2021-01-01
	Owner      string `json:"owner"` // jm33
	Permission string `json:"perm"`  // -rwxr-xr-x
}

// FileStat stat info of a file
type FileStat struct {
	Name       string `json:"name"`
	Permission string `json:"permission"`
	Checksum   string `json:"checksum"`
	Size       int64  `json:"size"`
}

// LsPath ls path and return a json
func LsPath(path string) (string, error) {
	parse_fileInfo := func(info os.FileInfo) (dent Dentry) {
		dent.Name = info.Name()
		dent.Date = info.ModTime().String()
		dent.Ftype = "file"
		dent.Permission = info.Mode().String()
		dent.Size = fmt.Sprintf("%d bytes", info.Size())
		return dent
	}
	// if it's a file, return its info
	if IsFileExist(path) {
		info, statErr := os.Stat(path)
		if statErr != nil {
			logging.Debugf("LsPath: %v", statErr)
			return "", statErr
		}
		dents := []Dentry{parse_fileInfo(info)}
		jsonData, err := json.Marshal(dents)
		if err != nil {
			logging.Debugf("LsPath: %v", err)
			return "", err
		}
		return string(jsonData), nil
	}

	files, err := os.ReadDir(path)
	if err != nil {
		logging.Debugf("LsPath: %v", err)
		return "", err
	}

	// parse
	var dents []Dentry
	for _, f := range files {
		info, statErr := f.Info()
		if statErr != nil {
			logging.Debugf("LsPath: %v", statErr)
			continue
		}
		dents = append(dents, parse_fileInfo(info))
	}

	// json
	jsonData, err := json.Marshal(dents)
	return string(jsonData), err
}

// IsCommandExist check if an executable is in $PATH
func IsCommandExist(exe string) bool {
	_, err := exec.LookPath(exe)
	return err == nil
}

// IsFileExist check if a file exists
func IsFileExist(path string) bool {
	f, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err == nil {
		return !f.IsDir()
	}

	return true
}

// IsExist check if a path exists
func IsExist(path string) bool {
	_, statErr := os.Stat(path)
	return !os.IsNotExist(statErr)
}

// IsDirExist check if a directory exists
func IsDirExist(path string) bool {
	f, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err == nil {
		return f.IsDir()
	}

	return false
}

// RemoveItemFromArray remove string/int from slice
func RemoveItemFromArray[T string | int](to_remove T, sliceList []T) []T {
	list := []T{}
	for _, item := range sliceList {
		if item != to_remove {
			list = append(list, item)
		}
	}
	return list
}

// RemoveDupsFromArray remove duplicated string/int from slice
func RemoveDupsFromArray[T string | int](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// IntArrayToStringArray convert int array to string array
func IntArrayToStringArray(arr []int) []string {
	var res []string
	for _, v := range arr {
		res = append(res, fmt.Sprintf("%d", v))
	}
	return res
}

// AppendToFile append bytes to a file
func AppendToFile(filename string, data []byte) (err error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()

	if _, err = f.Write(data); err != nil {
		return
	}
	return
}

// AppendTextToFile append text to a file
func AppendTextToFile(filename string, text string) (err error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		return
	}
	return
}

// IsStrInFile works like grep, check if a string is in a text file
func IsStrInFile(text, filepath string) bool {
	f, err := os.Open(filepath)
	if err != nil {
		logging.Debugf("IsStrInFile: %v", err)
		return false
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if strings.Contains(s.Text(), text) {
			return true
		}
	}

	return false
}

// Copy copy file or directory from src to dst
func Copy(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// if destination is a directory
	f, err := os.Stat(dst)
	if err == nil {
		if f.IsDir() {
			dst = filepath.Join(dst, filepath.Base(src))
		}
	}

	// if dst is a file and exists
	if IsFileExist(dst) {
		err = os.RemoveAll(dst)
		if err != nil {
			logging.Debugf("Copy: %s exists and cannot be removed: %v", dst, err)
		}
	}

	return os.WriteFile(dst, in, 0o755)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, d.Type().Perm())
		}

		return copyFile(path, targetPath)
	})
}

// FileBaseName extracts the base name of the file from a given path.
func FileBaseName(path string) string {
	// Use the standard library to safely get the base name
	return filepath.Base(filepath.Clean(path))
}

// FileAllocate allocate n bytes for a file, will delete the target file if already exists
func FileAllocate(filepath string, n int64) (err error) {
	if IsExist(filepath) {
		err = os.Remove(filepath)
		if err != nil {
			return
		}
	}
	f, err := os.Create(filepath)
	if err != nil {
		return
	}
	defer f.Close()
	f.Truncate(n)

	return
}

// FileSize calc file size
func FileSize(path string) (size int64) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	size = fi.Size()
	return
}

func ReplaceBytesInFile(path string, old []byte, replace_with []byte) (err error) {
	file_bytes, err := os.ReadFile(path)
	if err != nil {
		return
	}

	to_write := bytes.ReplaceAll(file_bytes, old, replace_with)
	return os.WriteFile(path, to_write, 0o644)
}

// FindHolesInBinary find holes in a binary file that are big enough for a payload
func FindHolesInBinary(fdata []byte, size int64) (indexes []int64, err error) {
	// find_hole finds a hole from start
	find_hole := func(start int64) (end int64) {
		for i := start; i < int64(len(fdata)); i++ {
			if fdata[i] == 0 {
				end = i
			} else {
				break
			}
		}
		return
	}

	// find holes
	for i := int64(0); i < int64(len(fdata)); i++ {
		if fdata[i] == 0 {
			end := find_hole(i)
			// if hole is big enough
			if end-i >= size {
				indexes = append(indexes, i)
			}
			i = end
		}
	}

	return
}

// IsDirWritable check if a directory is writable
func IsDirWritable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return false
	}
	// Check if the current user can write to the directory
	testFile := filepath.Join(path, RandMD5String())
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)
	return true
}

// GetWritablePaths get all writable paths in a directory up to a given depth
func GetWritablePaths(root_path string, depth, max int) ([]string, error) {
	if depth < 0 {
		return nil, fmt.Errorf("invalid depth: %d", depth)
	}

	var writablePaths []string
	var searchPaths func(path string, currentDepth int) error

	searchPaths = func(path string, currentDepth int) error {
		if currentDepth > depth {
			return nil
		}

		files, err := os.ReadDir(path)
		if err != nil {
			logging.Debugf("Skipping unreadable directory %s: %v", path, err)
			return nil
		}

		for _, file := range files {
			fullPath := filepath.Join(path, file.Name())
			if file.IsDir() {
				if IsDirWritable(fullPath) {
					writablePaths = append(writablePaths, fullPath)
				}
				if len(writablePaths) >= max {
					return nil
				}
				TakeABlink() // avoid being too fast and causing high CPU usage
				if err := searchPaths(fullPath, currentDepth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := searchPaths(root_path, 0); err != nil {
		return nil, err
	}

	if len(writablePaths) == 0 {
		return nil, fmt.Errorf("no writable paths found in %s", root_path)
	}

	return writablePaths, nil
}

// Agent-specific file operations for centralized control

// WriteFileAgent is a centralized file writing function for agent operations.
// This function wraps all file writing operations to allow for future modifications
// such as encryption, steganography, or other security enhancements.
func WriteFileAgent(filename string, data []byte, perm os.FileMode) error {
	// Future enhancements can be added here:
	// - File encryption before writing
	// - Steganography to hide files
	// - Anti-forensics techniques
	// - Atomic writes with temporary files
	// - Logging for debugging (but be careful with OpSec)

	logging.Debugf("Agent: Writing %d bytes to %s with permissions %o", len(data), filename, perm)

	// Currently just wraps os.WriteFile, but can be enhanced later
	return os.WriteFile(filename, data, perm)
}

// CreateFileAgent is a centralized file creation function for agent operations.
// This function wraps file creation operations to allow for future modifications.
func CreateFileAgent(filename string) (*os.File, error) {
	logging.Debugf("Agent: Creating file %s", filename)

	// Future enhancements can be added here:
	// - Hidden file attributes
	// - Special file creation flags
	// - Anti-forensics techniques

	return os.Create(filename)
}

// OpenFileAgent is a centralized file opening function for agent operations.
// This function wraps file opening operations to allow for future modifications.
func OpenFileAgent(filename string, flag int, perm os.FileMode) (*os.File, error) {
	logging.Debugf("Agent: Opening file %s with flags %d and permissions %o", filename, flag, perm)

	// Future enhancements can be added here:
	// - Special file opening flags
	// - Anti-forensics techniques
	// - File locking mechanisms

	return os.OpenFile(filename, flag, perm)
}

// AppendToFileAgent is a centralized file appending function for agent operations.
// This function wraps file appending operations to allow for future modifications.
func AppendToFileAgent(filename string, data []byte) error {
	logging.Debugf("Agent: Appending %d bytes to %s", len(data), filename)

	// Future enhancements can be added here:
	// - Encryption before appending
	// - Anti-forensics techniques

	f, err := OpenFileAgent(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write(data); err != nil {
		return err
	}
	return nil
}

// AppendTextToFileAgent is a centralized text appending function for agent operations.
// This function wraps text appending operations to allow for future modifications.
func AppendTextToFileAgent(filename string, text string) error {
	logging.Debugf("Agent: Appending text to %s", filename)

	// Future enhancements can be added here:
	// - Encryption before appending
	// - Anti-forensics techniques

	f, err := OpenFileAgent(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		return err
	}
	return nil
}
