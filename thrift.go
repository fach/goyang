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

package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/openconfig/goyang/pkg/indent"
	"github.com/openconfig/goyang/pkg/yang"
)

func init() {
	register(&formatter{
		name: "thrift",
		f:    doThrift,
		help: "display tree in a thrift format",
	})
}

func doThrift(w io.Writer, entries []*yang.Entry) {
	for _, e := range entries {
		for _, e := range flatten(e) {
			FormatStruct(w, e)
		}
	}
}

// kind2thrift maps base yang types to protocol buffer types.
var kind2thrift = map[yang.TypeKind]string{
	yang.Yint8:   "byte",        // int in range [-128, 127]
	yang.Yint16:  "i16",         // int in range [-32768, 32767]
	yang.Yint32:  "i32",         // int in range [-2147483648, 2147483647]
	yang.Yint64:  "i64",         // int in range [-9223372036854775808, 9223372036854775807]
	yang.Yuint8:  "TODO-uint8",  // int in range [0, 255]
	yang.Yuint16: "TODO-uint16", // int in range [0, 65535]
	yang.Yuint32: "TODO-uint32", // int in range [0, 4294967295]
	yang.Yuint64: "TODO-uint64", // int in range [0, 18446744073709551615]

	yang.Ybinary:             "TODO-bytes",     // arbitrary data
	yang.Ybits:               "TODO-bits",      // set of bits or flags
	yang.Ybool:               "bool",           // true or false
	yang.Ydecimal64:          "TODO-decimal64", // signed decimal number
	yang.Yenum:               "enum",           // enumerated strings
	yang.Yidentityref:        "string",         // reference to abstrace identity
	yang.YinstanceIdentifier: "TODO-ii",        // reference of a data tree node
	yang.Yleafref:            "string",         // reference to a leaf instance
	yang.Ystring:             "string",         // human readable string
	yang.Yunion:              "TODO-union",     // choice of types
}

func FormatStruct(w io.Writer, e *yang.Entry) {
	var names []string

	// TODO: Implement RPC support

	names = nil
	for k, se := range e.Dir {
		if se.RPC == nil {
			names = append(names, k)
		}
	}
	if len(names) == 0 {
		return
	}

	fmt.Fprintln(w)
	if e.Description != "" {
		fmt.Fprintln(indent.NewWriter(w, "// "), e.Description)
	}
	fmt.Fprintf(w, "struct %s {\n", fixName(e.Name))

	sort.Strings(names)
	for x, k := range names {
		se := e.Dir[k]
		k := strings.Replace(k, "-", "_", -1)
		if se.Description != "" {
			fmt.Fprintln(indent.NewWriter(w, "  // "), se.Description)
		}
		fmt.Fprintf(w, "    %d: ", x+1)
		if se.ListAttr != nil {
			fmt.Fprint(w, "list")
		} else {
			fmt.Fprint(w, "optional ")
		}
		if len(se.Dir) == 0 && se.Type != nil {
			// TODO(borman): this is probably an empty container.
			kind := "UNKNOWN TYPE"
			if se.Type != nil {
				kind = kind2proto[se.Type.Kind]
			}
			if se.ListAttr != nil {
				fmt.Fprintf(w, "<%s> ", kind)
			} else {
				fmt.Fprintf(w, "%s ", kind)
			}
			fmt.Fprintf(w, "%s; // %s\n", k, yang.Source(se.Node))
			//fmt.Fprintf(w, "%s;\n", k)
			continue
		}
		if se.ListAttr != nil {
			fmt.Fprintf(w, "<%s> ", fixName(se.Name))
		} else {
			fmt.Fprintf(w, "%s ", fixName(se.Name))
		}
		fmt.Fprintf(w, "%s; // %s\n", k, yang.Source(se.Node))
		//fmt.Fprintf(w, "%s;\n", k)

	}
	// { to match the brace below to keep brace matching working
	fmt.Fprintln(w, "}")
}
