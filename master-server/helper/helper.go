package helper

import (
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type DiskUsage struct {
	stat *syscall.Statfs_t
}

func MoveFile(sourcePath string, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}

func Symlink(sourcePath string, destPath string) error {
	app := "/bin/ln"

	arg0 := "-s"

	exec.Command(app, arg0, sourcePath, destPath).Run()

	return nil
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func isSlaveWritable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}

func ExportsExist(line string) (bool, error) {
	b, err := ioutil.ReadFile("/etc/fstab")
	if err != nil {
		return false, err
	}

	content := strings.Split(string(b), "\n")

	for _, eline := range content {
		if standardizeSpaces(eline) == line {
			return true, nil
		}
	}

	return false, nil
}

func AppendExports(line string) error {
	f, err := os.OpenFile("/etc/fstab", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.WriteString(line); err != nil {
		return err
	}

	return nil
}

func AvailableBrick() (string, error) {
	files, err := ioutil.ReadDir("/data")
	if err != nil {
		return "", err
	}

	for _, f := range files {
		if f.IsDir() == true {
			path := "/data/" + f.Name()
			usage := NewDiskUsage(path)
			used_percent := usage.Usage() * 100

			if used_percent < 92 {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("Cannot get available disk !")
}

// Returns an object holding the disk usage of volumePath
// This function assumes volumePath is a valid path
// https://github.com/ricochet2200/go-disk-usage/blob/master/du/diskusage.go
func NewDiskUsage(volumePath string) *DiskUsage {

	var stat syscall.Statfs_t
	syscall.Statfs(volumePath, &stat)
	return &DiskUsage{&stat}
}

// Total free bytes on file system
func (this *DiskUsage) Free() uint64 {
	return this.stat.Bfree * uint64(this.stat.Bsize)
}

// Total size of the file system
func (this *DiskUsage) Size() uint64 {
	return this.stat.Blocks * uint64(this.stat.Bsize)
}

// Total bytes used in file system
func (this *DiskUsage) Used() uint64 {
	return this.Size() - this.Free()
}

// Percentage of use on the file system
func (this *DiskUsage) Usage() float32 {
	return float32(this.Used()) / float32(this.Size())
}
