// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package locale

import (
	"strings"

	"github.com/benoitkugler/textprocessing/fribidi"
)

func (l Lingua) shapeArabic(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	var parType fribidi.ParType = fribidi.RTL
	for _, l := range lines {
		r := []rune(l)
		v, _ := fribidi.LogicalToVisual(fribidi.DefaultFlags, r, &parType)
		r = v.Str
		for i, j := 0, 0; i < len(r); i++ {
			ch := r[i]
			// Skip ZWNBSP as they have no purpose at render time and GNU Unifont shows them.
			if ch == 0xFEFF {
				continue
			}
			r[j] = ch
			j++
		}
		out = append(out, string(v.Str))
	}
	return strings.Join(out, "\n")
}
