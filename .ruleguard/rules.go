//go:build ignore
// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

// Common Go best practices and performance rules

func emptySlice(m dsl.Matcher) {
	m.Match(`len($x) == 0`).
		Where(m["x"].Type.Is(`[]$_`)).
		Report(`replace len($x) == 0 with $x == nil`)
}

func emptyMap(m dsl.Matcher) {
	m.Match(`len($x) == 0`).
		Where(m["x"].Type.Is(`map[$_]$_`)).
		Report(`replace len($x) == 0 with $x == nil`)
}

func appendAssign(m dsl.Matcher) {
	m.Match(`for $*_ { $*_; $x = append($x, $*args); $*_ }`).
		Report(`$x = append($x, ...) should be avoided in loops, consider pre-allocating`)
	m.Match(`for $*_ { $*_; $x = append($x, $*args) }`).
		Report(`$x = append($x, ...) should be avoided in loops, consider pre-allocating`)
}

func stringConcatInLoop(m dsl.Matcher) {
	m.Match(`$x += $y`).
		Where(m["x"].Type.Is(`string`)).
		Report(`string concatenation in loop is inefficient, use strings.Builder`)
}

func httpBodyNotClosed(m dsl.Matcher) {
	m.Match(`$resp, $err := http.$method($*args)`).
		Report(`http response body should be closed with defer resp.Body.Close()`)
}

func contextTODO(m dsl.Matcher) {
	m.Match(`context.TODO()`).
		Report(`context.TODO() should be replaced with proper context`)
}

func errorStringFormat(m dsl.Matcher) {
	m.Match(`errors.New(fmt.Sprintf($*args))`).
		Report(`use fmt.Errorf instead of errors.New(fmt.Sprintf(...))`)
}

func unnecessaryElse(m dsl.Matcher) {
	m.Match(`if $cond { return $*_ } else { $*body }`).
		Report(`unnecessary else after return`)
}

func sliceInit(m dsl.Matcher) {
	m.Match(`var $x = make([]$T, 0)`).
		Report(`use var $x []$T instead of var $x = make([]$T, 0)`)
}

func mapInit(m dsl.Matcher) {
	m.Match(`make(map[$K]$V, 0)`).
		Report(`use make(map[$K]$V) instead of make(map[$K]$V, 0)`)
}

func regexpCompile(m dsl.Matcher) {
	m.Match(`regexp.Compile($pattern)`).
		Report(`consider using regexp.MustCompile for static patterns or cache compiled regexp`)
}

func timeFormat(m dsl.Matcher) {
	m.Match(`$t.Format("2006-01-02 15:04:05")`).
		Report(`use time.RFC3339 or other predefined layouts when possible`)
}

func sqlRowsClose(m dsl.Matcher) {
	m.Match(`$rows, $err := $db.Query($*args)`).
		Report(`sql rows should be closed with defer rows.Close()`)
}

func mutexCopy(m dsl.Matcher) {
	m.Match(`$x = $y`).
		Where(m["y"].Type.Is(`sync.Mutex`) || m["y"].Type.Is(`sync.RWMutex`)).
		Report(`mutex should not be copied, use pointer`)
}

func channelClose(m dsl.Matcher) {
	m.Match(`close($ch)`).
		Report(`ensure channel is not closed multiple times`)
}

func goroutineLeak(m dsl.Matcher) {
	m.Match(`go func() { $*body }()`).
		Report(`ensure goroutine can exit to prevent leaks`)
}

func interfaceNil(m dsl.Matcher) {
	m.Match(`$x == nil`).
		Where(m["x"].Type.Is(`interface{}`)).
		Report(`interface{} comparison with nil might not work as expected`)
}

func typeAssertion(m dsl.Matcher) {
	m.Match(`$x.($T)`).
		Where(!m.File().Name.Matches(`.*_test\.go$`)).
		Report(`type assertion should check ok value: $x, ok := $x.($T)`)
}

func rangeOverMap(m dsl.Matcher) {
	m.Match(`for $_, $_ := range $m { $*_ }`).
		Where(m["m"].Type.Is(`map[$_]$_`)).
		Report(`map iteration order is not guaranteed`)
}
