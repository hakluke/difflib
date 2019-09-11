// Copyright 2012 Aryan Naraghi (aryan.naraghi@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package difflib provides functionality for computing the difference
// between two sequences of strings.
package difflib

import (
	"bytes"
	"fmt"
	"html"
	"math"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// DeltaType describes the relationship of elements in two
// sequences. The following table provides a summary:
//
//    Constant    Code   Meaning
//   ----------  ------ ---------------------------------------
//    Common      " "    The element occurs in both sequences.
//    LeftOnly    "-"    The element is unique to sequence 1.
//    RightOnly   "+"    The element is unique to sequence 2.
type DeltaType int

const (
	Common DeltaType = iota
	LeftOnly
	RightOnly
)

// String returns a string representation for DeltaType.
func (t DeltaType) String() string {
	switch t {
	case Common:
		return " "
	case LeftOnly:
		return "-"
	case RightOnly:
		return "+"
	}
	return "?"
}

type DiffRecord struct {
	Payload string
	Delta   DeltaType
}

type Line struct {
	Number  []int
	Delta   string
	Payload string
}

// String returns a string representation of d. The string is a
// concatenation of the delta type and the payload.
func (d DiffRecord) String() string {
	return fmt.Sprintf("%s %s", d.Delta, d.Payload)
}

// Diff returns the result of diffing the seq1 and seq2.
func Diff(seq1, seq2 []string, trim bool) (withLines []Line) {
	// Trims any common elements at the heads and tails of the
	// sequences before running the diff algorithm. This is an
	// optimization.
	start, end := numEqualStartAndEndElements(seq1, seq2)

	var startedAt int
	var diff []DiffRecord
	if !trim {
		for _, content := range seq1[:start] {
			diff = append(diff, DiffRecord{content, Common})
		}
	} else {
		if start > 0 {
			diff = append([]DiffRecord{DiffRecord{seq1[start-1 : start][0], Common}}, diff...)
			startedAt = start - 1
		}
		if start > 1 {
			diff = append([]DiffRecord{DiffRecord{seq1[start-2 : start][0], Common}}, diff...)
			startedAt = start - 2
		}
		if start > 2 {
			diff = append([]DiffRecord{DiffRecord{seq1[start-3 : start][0], Common}}, diff...)
			startedAt = start - 3
		}
	}

	diffRes := compute(seq1[start:len(seq1)-end], seq2[start:len(seq2)-end])
	diff = append(diff, diffRes...)

	if !trim {
		for _, content := range seq1[len(seq1)-end:] {
			diff = append(diff, DiffRecord{content, Common})
		}
	} else {
		if end > 0 {
			diff = append(diff, DiffRecord{seq1[len(seq1)-end : len(seq1)-end+1][0], Common})
		}
		if end > 1 {
			diff = append(diff, DiffRecord{seq1[len(seq1)-end : len(seq1)-end+2][1], Common})
		}
		if end > 2 {
			diff = append(diff, DiffRecord{seq1[len(seq1)-end : len(seq1)-end+3][2], Common})
		}
	}

	var l, r int

	if trim {
		l, r = startedAt, startedAt
	}

	for dIndex, d := range diff {
		if d.Delta == LeftOnly {
			l++
			// num = ++l
		} else if d.Delta == RightOnly {
			r++
			// num = ++r
		} else {
			r++
			l++
		}

		num := []int{l, r}

		firstBefore := (dIndex != 0 && (diff[dIndex-1].Delta == RightOnly || diff[dIndex-1].Delta == LeftOnly))
		secondBefore := (dIndex > 1 && (diff[dIndex-2].Delta == RightOnly || diff[dIndex-2].Delta == LeftOnly))
		thirdBefore := (dIndex > 2 && (diff[dIndex-3].Delta == RightOnly || diff[dIndex-3].Delta == LeftOnly))

		firstAfter := (dIndex < len(diff)-1 && (diff[dIndex+1].Delta == RightOnly || diff[dIndex+1].Delta == LeftOnly))
		secondAfter := (dIndex < len(diff)-2 && (diff[dIndex+2].Delta == RightOnly || diff[dIndex+2].Delta == LeftOnly))
		thirdAfter := (dIndex < len(diff)-3 && (diff[dIndex+3].Delta == RightOnly || diff[dIndex+3].Delta == LeftOnly))

		if !trim || (d.Delta == RightOnly || d.Delta == LeftOnly) || firstBefore || secondBefore || thirdBefore || firstAfter || secondAfter || thirdAfter {
			line := Line{num, d.Delta.String(), d.Payload}
			withLines = append(withLines, line)
		}
	}

	return withLines
}

// HTMLDiff returns the results of diffing seq1 and seq2 as an HTML
// string. The resulting HTML is a table without the opening and
// closing table tags. Each table row represents a DiffRecord. The
// first and last columns contain the "line numbers" for seq1 and
// seq2, respectively (the function assumes that seq1 and seq2
// represent the lines in a file). The second and third columns
// contain the actual file contents.
//
// The cells that contain line numbers are decorated with the class
// "line-num". The cells that contain deleted elements are decorated
// with "deleted" and the cells that contain added elements are
// decorated with "added".
func HTMLDiff(difference []Line, header string) string {
	if header == "" {
		header = "Difference"
	}

	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, `<table class="diff-table"><tr class="table-header"><td><i class="fa fa-chevron-down collapse-icon"></i></td><td colspan="3">%s</td></tr>`, html.EscapeString(header))

	dmp := diffmatchpatch.New()
	var wDiffs []diffmatchpatch.Diff
	for index, d := range difference {
		if index != 0 && difference[index].Number[0] != difference[index-1].Number[0] && difference[index].Number[0]-1 != difference[index-1].Number[0] {
			fmt.Fprintf(buf, `<tr class="new-part"><td colspan="2">...</td><td>...</td></tr>`)
		}
		if index == 0 && ((difference[index].Number[0] != 1 && difference[index].Number[0] != 0) || (difference[index].Number[1] != 1 && difference[index].Number[0] != 0)) {
			fmt.Fprintf(buf, `<tr class="new-part"><td colspan="2">...</td><td>...</td></tr>`)
		}
		fmt.Fprintf(buf, `<tr>`)
		num := d.Number
		if d.Delta == "-" {
			if len(difference) != index+1 && difference[index+1].Delta == "+" && difference[index].Number[0] == difference[index+1].Number[1] {
				var content string
				wDiffs = dmp.DiffMain(d.Payload, difference[index+1].Payload, false)
				for _, w := range wDiffs {

					if w.Type == -1 {
						content += "<span class=\"deleted-text\">" + html.EscapeString(w.Text) + "</span>"
					} else if w.Type != 1 {
						content += html.EscapeString(w.Text)
					}
				}

				fmt.Fprintf(buf, `<td class="line-num line-num-deleted">%d</td><td class="line-num line-num-deleted"></td><td class="deleted code"><span class="delta-type">%s</span><pre><code>%s</code></pre></td>`, num[0], d.Delta, content)
			} else {
				fmt.Fprintf(buf, `<td class="line-num line-num-deleted">%d</td><td class="line-num line-num-deleted"></td><td class="deleted code"><span class="delta-type">%s</span><pre><code>%s</code></pre></td>`, num[0], d.Delta, html.EscapeString(d.Payload))
			}
		} else if d.Delta == "+" {
			if index != 0 && difference[index-1].Delta == "-" && difference[index].Number[1] == difference[index-1].Number[0] {
				var content string
				for _, w := range wDiffs {
					if w.Type == 1 {
						content += "<span class=\"added-text\">" + html.EscapeString(w.Text) + "</span>"
					} else if w.Type != -1 {
						content += html.EscapeString(w.Text)
					}
				}
				fmt.Fprintf(buf, `<td class="line-num line-num-added"></td><td class="line-num line-num-added">%d</td><td class="added code"><span class="delta-type">%s</span><pre><code>%s</code></pre></td>`, num[1], d.Delta, content)
			} else {
				fmt.Fprintf(buf, `<td class="line-num line-num-added"></td><td class="line-num line-num-added">%d</td><td class="added code"><span class="delta-type">%s</span><pre><code>%s</code></pre></td>`, num[1], d.Delta, html.EscapeString(d.Payload))
			}
		} else {
			fmt.Fprintf(buf, `<td class="line-num line-num-normal">%d</td><td class="line-num line-num-normal">%d</td><td class="code"><span class="delta-type">%s</span><pre><code>%s</code></pre></td>`, num[0], num[1], d.Delta, html.EscapeString(d.Payload))
		}
		buf.WriteString("</tr>\n")
	}
	buf.WriteString("</table>")
	return buf.String()
}

// numEqualStartAndEndElements returns the number of elements a and b
// have in common from the beginning and from the end. If a and b are
// equal, start will equal len(a) == len(b) and end will be zero.
func numEqualStartAndEndElements(seq1, seq2 []string) (start, end int) {
	for start < len(seq1) && start < len(seq2) && seq1[start] == seq2[start] {
		start++
	}
	i, j := len(seq1)-1, len(seq2)-1
	for i > start && j > start && seq1[i] == seq2[j] {
		i--
		j--
		end++
	}
	return
}

// intMatrix returns a 2-dimensional slice of ints with the given
// number of rows and columns.
func intMatrix(rows, cols int) [][]int {
	matrix := make([][]int, rows)
	for i := 0; i < rows; i++ {
		matrix[i] = make([]int, cols)
	}
	return matrix
}

// longestCommonSubsequenceMatrix returns the table that results from
// applying the dynamic programming approach for finding the longest
// common subsequence of seq1 and seq2.
func longestCommonSubsequenceMatrix(seq1, seq2 []string) [][]int {
	matrix := intMatrix(len(seq1)+1, len(seq2)+1)
	for i := 1; i < len(matrix); i++ {
		for j := 1; j < len(matrix[i]); j++ {
			if seq1[len(seq1)-i] == seq2[len(seq2)-j] {
				matrix[i][j] = matrix[i-1][j-1] + 1
			} else {
				matrix[i][j] = int(math.Max(float64(matrix[i-1][j]),
					float64(matrix[i][j-1])))
			}
		}
	}
	return matrix
}

// compute is the unexported helper for Diff that returns the results of
// diffing left and right.
func compute(seq1, seq2 []string) (diff []DiffRecord) {
	matrix := longestCommonSubsequenceMatrix(seq1, seq2)
	i, j := len(seq1), len(seq2)
	for i > 0 || j > 0 {
		if i > 0 && matrix[i][j] == matrix[i-1][j] {
			diff = append(diff, DiffRecord{seq1[len(seq1)-i], LeftOnly})
			i--
		} else if j > 0 && matrix[i][j] == matrix[i][j-1] {
			diff = append(diff, DiffRecord{seq2[len(seq2)-j], RightOnly})
			j--
		} else if i > 0 && j > 0 {
			diff = append(diff, DiffRecord{seq1[len(seq1)-i], Common})
			i--
			j--
		}
	}
	return
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
