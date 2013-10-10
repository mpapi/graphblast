// Bundle extracts file data from the arbitrary data sections of executables,
// with fallback to the filesystem.
package bundle

// TODO Add/split out versions for different platforms

import (
	"debug/elf"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

func ReadFromBinary(filename string) ([]byte, error) {
	file, err := elf.Open(os.Args[0])
	if err != nil {
		return []byte{}, err
	}

	sec := file.Section(filename)
	if sec == nil {
		return []byte{}, errors.New("no section for filename")
	}

	return sec.Data()
}

func ReadFile(filename string) []byte {
	bytes, err := ReadFromBinary(filename)
	if err == nil {
		return bytes
	}

	bytes, err = ioutil.ReadFile(filename)
	if err == nil {
		return bytes
	}

	panic(fmt.Sprintf("unable to open file %s", filename))
}
