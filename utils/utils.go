package diffutils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/hakluke/difflib"
	"github.com/google/uuid"
)

func Diff(seq1, seq2 []string, trim bool) ([]difflib.Line, error) {
	diff := []difflib.Line{}

	seq1File := fmt.Sprintf("diff-input-%s.txt", uuid.New().String())
	seq2File := fmt.Sprintf("diff-input-%s.txt", uuid.New().String())
	output := fmt.Sprintf("diff-output-%s.txt", uuid.New().String())

	err := saveToHTMLFile(seq1, seq1File)
	if err != nil {
		return diff, err
	}
	defer os.Remove(path.Join("/tmp/", seq1File))

	err = saveToHTMLFile(seq2, seq2File)
	if err != nil {
		return diff, err
	}
	defer os.Remove(path.Join("/tmp/", seq2File))

	defer os.Remove(path.Join("/tmp/", output))

	cmd := exec.Command("difflib_standalone", []string{path.Join("/tmp/", seq1File), path.Join("/tmp/", seq2File), path.Join("/tmp/", output), "true"}...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return diff, fmt.Errorf("an error has occured in running the format tool. stderr: %s, error: %s", strings.TrimSuffix(stderr.String(), "\n"), err.Error())
	}

	diff, err = fileToDiff(output)
	if err != nil {
		fmt.Println(err)
		return diff, err
	}

	return diff, nil
}

func saveToHTMLFile(lines []string, filename string) error {
	file := path.Join("/tmp/", filename)

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, d := range lines {
		_, err = f.WriteString(d + "\n")
		if err != nil {
			return err
		}
	}

	return f.Sync()
}

func fileToDiff(inputFilename string) ([]difflib.Line, error) {
	diff := []difflib.Line{}
	inputFile := path.Join("/tmp/", inputFilename)

	f, err := os.Open(inputFile)
	if err != nil {
		return diff, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		var buffer bytes.Buffer

		var l []byte
		var isPrefix bool
		for {
			l, isPrefix, err = reader.ReadLine()
			buffer.Write(l)
			if !isPrefix {
				break
			}
			if err != nil {
				if err != io.EOF {
					return diff, err
				}
				break
			}
		}

		x := buffer.String()
		parts := strings.SplitN(x, ":", 4)

		// format is "leftLineNum rightLineNum delta payload"
		line := difflib.Line{Number: []int{0, 0}}
		if len(parts) > 1 {
			line.Number[0], err = strconv.Atoi(parts[0])
			if err != nil {
				return diff, err
			}

			line.Number[1], err = strconv.Atoi(parts[1])
			if err != nil {
				return diff, err
			}

			line.Delta = parts[2]
			line.Payload = parts[3]
			diff = append(diff, line)
		}

		if err == io.EOF {
			break
		}
	}

	return diff, nil
}
