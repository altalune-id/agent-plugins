# Go Proverbs: Examples

One bad and one good example per proverb, in the order of `SKILL.md`. Each block is a
complete file that passes `go vet`.

## Don't communicate by sharing memory, share memory by communicating.

Look for: goroutines writing one map or slice under a shared lock.

```go
// Avoid
package tally

import "sync"

func Count(lines []string, word func(string) string) map[string]int {
	var mu sync.Mutex
	var wg sync.WaitGroup
	counts := map[string]int{}
	for _, l := range lines {
		wg.Go(func() { mu.Lock(); counts[word(l)]++; mu.Unlock() })
	}
	wg.Wait()
	return counts
}
```

```go
// Prefer
package tally

func Count(lines []string, word func(string) string) map[string]int {
	words := make(chan string, len(lines))
	for _, l := range lines {
		go func() { words <- word(l) }()
	}
	counts := map[string]int{}
	for range lines {
		counts[<-words]++
	}
	return counts
}
```

Why: each word is handed over once, and only the receiving loop owns the map.

## Concurrency is not parallelism.

Look for: a goroutine per item for cheap work, with no benchmark.

```go
// Avoid
package stats

import "sync"

func Sum(xs []int) int {
	var mu sync.Mutex
	var wg sync.WaitGroup
	total := 0
	for _, x := range xs {
		wg.Go(func() { mu.Lock(); total += x; mu.Unlock() })
	}
	wg.Wait()
	return total
}
```

```go
// Prefer
package stats

func Sum(xs []int) int {
	total := 0
	for _, x := range xs {
		total += x
	}
	return total
}
```

Why: the work has no independent parts, so goroutines add scheduling and lock cost, not speed.

## Channels orchestrate; mutexes serialize.

Look for: a buffered channel of size 1 used as a lock.

```go
// Avoid
package counter

type Counter struct {
	lock chan struct{}
	n    int
}

func New() *Counter { return &Counter{lock: make(chan struct{}, 1)} }

func (c *Counter) Inc() {
	c.lock <- struct{}{}
	c.n++
	<-c.lock
}
```

```go
// Prefer
package counter

import "sync"

type Counter struct {
	mu sync.Mutex
	n  int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}
```

Why: guarding one field is serialization, which is the mutex's job; it also makes the zero value usable.

## The bigger the interface, the weaker the abstraction.

Look for: a many-method interface beside its only implementation.

```go
// Avoid
package user

type User struct{ ID, Email string }

type Store interface {
	Find(id string) (User, error)
	FindByEmail(email string) (User, error)
	List() ([]User, error)
	Create(u User) error
	Update(u User) error
	Delete(id string) error
}
```

```go
// Prefer
package notify

type User struct{ ID, Email string }
type finder interface {
	Find(id string) (User, error)
}

func Email(f finder, id string) (string, error) {
	u, err := f.Find(id)
	if err != nil {
		return "", err
	}
	return u.Email, nil
}
```

Why: the consumer asks for the one method it calls, so a fake in a test is three lines.

## Make the zero value useful.

Look for: a type that panics when declared without its constructor.

```go
// Avoid
package cache

type Cache struct{ items map[string]string }

func New() *Cache { return &Cache{items: map[string]string{}} }

func (c *Cache) Set(k, v string) { c.items[k] = v }
```

```go
// Prefer
package cache

type Cache struct{ items map[string]string }

func (c *Cache) Set(k, v string) {
	if c.items == nil {
		c.items = map[string]string{}
	}
	c.items[k] = v
}

func (c *Cache) Get(k string) string { return c.items[k] }
```

Why: `var c Cache` works, so callers and struct fields need no setup step.

## interface{} says nothing.

Look for: `any` parameters plus type assertions where the types are known.

```go
// Avoid
package mathx

func Max(xs []any) any {
	best := xs[0]
	for _, x := range xs[1:] {
		if x.(int) > best.(int) {
			best = x
		}
	}
	return best
}
```

```go
// Prefer
package mathx

import "cmp"

func Max[T cmp.Ordered](xs []T) T {
	best := xs[0]
	for _, x := range xs[1:] {
		best = max(best, x)
	}
	return best
}
```

Why: the constraint states the contract, and a wrong type fails at compile time, not as a panic.

## Gofmt's style is no one's favorite, yet gofmt is everyone's favorite.

Look for: files that `gofmt -l` lists, or review threads about layout.

```go
// Avoid
package config

type Config struct {
  Host    string
  Port int
  Debug     bool
}

func Default() Config { return Config{Host: "localhost",Port: 8080} }
```

```go
// Prefer
package config

type Config struct {
	Host  string
	Port  int
	Debug bool
}

func Default() Config { return Config{Host: "localhost", Port: 8080} }
```

Why: the Prefer block is exactly what `gofmt` prints; run it on save and in CI, and stop debating layout.

## A little copying is better than a little dependency.

Look for: a new `go.mod` entry used for one small function.

```go
// Avoid
package report

import "github.com/dustin/go-humanize"

func Rank(n int) string { return humanize.Ordinal(n) }
```

```go
// Prefer
package report

import "strconv"

func Rank(n int) string {
	s := "th"
	if d := n % 10; d >= 1 && d <= 3 && n%100/10 != 1 {
		s = [...]string{"st", "nd", "rd"}[d-1]
	}
	return strconv.Itoa(n) + s
}
```

Why: six owned lines cost less than a module to vet, update and audit.

## Syscall must always be guarded with build tags.

Look for: `syscall` in a file with no `//go:build` line or OS suffix.

```go
// Avoid
package proc

import "syscall"

func Stop(pid int) error { return syscall.Kill(pid, syscall.SIGTERM) }
```

```go
// Prefer
//go:build unix

package proc

import "syscall"

func Stop(pid int) error { return syscall.Kill(pid, syscall.SIGTERM) }
```

Why: `syscall.Kill` does not exist on Windows; the tag keeps that build green and a `proc_windows.go` supplies `Stop` there.

## Cgo must always be guarded with build tags.

Look for: `import "C"` in a file every build compiles.

```go
// Avoid
package fastabs

// #include <stdlib.h>
import "C"

func Abs(n int) int { return int(C.abs(C.int(n))) }
```

```go
// Prefer
//go:build cgo

package fastabs

// #include <stdlib.h>
import "C"

func Abs(n int) int { return int(C.abs(C.int(n))) }
```

Why: with `CGO_ENABLED=0` the file drops out, and a `//go:build !cgo` twin provides the pure Go path.

## Cgo is not Go.

Look for: cgo for something the standard library or a pure Go module already does.

```go
// Avoid
package geo

// #cgo LDFLAGS: -lm
// #include <math.h>
import "C"

func Hypot(x, y float64) float64 {
	return float64(C.hypot(C.double(x), C.double(y)))
}
```

```go
// Prefer
package geo

import "math"

func Hypot(x, y float64) float64 { return math.Hypot(x, y) }
```

Why: the pure Go call cross-compiles, needs no C toolchain and skips the cgo call overhead.

## With the unsafe package there are no guarantees.

Look for: `unsafe` in service code, often justified by speed with no benchmark.

```go
// Avoid
package text

import "unsafe"

func String(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }
```

```go
// Prefer
package text

func String(b []byte) string { return string(b) }
```

Why: the cast aliases mutable bytes as an immutable string and relies on header layout Go does not promise.

## Clear is better than clever.

Look for: a one-liner that takes a second read to decode.

```go
// Avoid
package grade

func Result(score int) string {
	return map[bool]string{true: "pass", false: "fail"}[score >= 50]
}
```

```go
// Prefer
package grade

func Result(score int) string {
	if score >= 50 {
		return "pass"
	}
	return "fail"
}
```

Why: the `if` reads in one pass and does not build a map on every call.

## Reflection is never clear.

Look for: `reflect` in handlers or services where a type switch would do.

```go
// Avoid
package format

import "reflect"

func Size(v any) int {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.String {
		return rv.Len()
	}
	return int(rv.Int())
}
```

```go
// Prefer
package format

func Size(v any) int {
	switch x := v.(type) {
	case string:
		return len(x)
	case int:
		return x
	}
	return 0
}
```

Why: the type switch names every case it handles; the reflect version panics at run time on a bool.

## Errors are values.

Look for: errors matched by their message text.

```go
// Avoid
package store

import "strings"

func IsMissing(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}
```

```go
// Prefer
package store

import "errors"

var ErrNotFound = errors.New("store: not found")

func IsMissing(err error) bool { return errors.Is(err, ErrNotFound) }
```

Why: a sentinel survives wrapping and rewording; a substring match breaks on both.

## Don't just check errors, handle them gracefully.

Look for: bare `return err`, or log-and-return at every layer.

```go
// Avoid
package config

import (
	"log"
	"os"
)

func Load(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
	}
	return b, err
}
```

```go
// Prefer
package config

import (
	"fmt"
	"os"
)

func Load(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return b, nil
}
```

Why: the wrap says what failed and keeps the cause for `errors.Is`; the edge logs it once.

## Design the architecture, name the components, document the details.

Look for: packages and types named `util`, `common`, `manager` or `helper`.

```go
// Avoid
package util

type Manager struct{ rates map[string]float64 }

func (m *Manager) Process(code string, amount float64) float64 {
	return amount * m.rates[code]
}
```

```go
// Prefer
package currency

type Converter struct{ rates map[string]float64 }

func (c *Converter) ToBase(code string, amount float64) float64 {
	return amount * c.rates[code]
}
```

Why: `currency.Converter.ToBase` tells the reader the design without opening the file.

## Documentation is for users.

Look for: godoc that narrates the implementation.

```go
// Avoid
package duration

import "time"

// Parse calls time.ParseDuration and then passes the result through min.
func Parse(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	return min(d, 24*time.Hour), err
}
```

```go
// Prefer
package duration

import "time"

// Parse reads a duration such as "90m" and caps it at 24 hours.
func Parse(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	return min(d, 24*time.Hour), err
}
```

Why: the caller needs the accepted input and the cap, not the steps inside.

## Don't panic.

Look for: `panic` on bad input, config or a failed request.

```go
// Avoid
package port

import "strconv"

func Parse(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return n
}
```

```go
// Prefer
package port

import (
	"fmt"
	"strconv"
)

func Parse(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parse port %q: %w", s, err)
	}
	return n, nil
}
```

Why: bad input is an expected failure; the caller decides what to do, and the process stays up.
