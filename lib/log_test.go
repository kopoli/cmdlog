package cmdlib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestLog(t *testing.T) {
	testdir := "test-log"

	filterfile := filepath.Join(testdir, "filter")
	logfile := filepath.Join(testdir, "log")

	contentsForRemoval := "# remove"

	var log *Log

	type opfunc func() error

	opSave := func() func() error {
		return func() error {
			return log.SaveDefaultFilters()
		}
	}

	opExists := func(filename string, should bool) func() error {
		return func() error {
			exist := FileExists(filename)
			if should != exist {
				msg := map[bool]string{
					true:  "exist",
					false: "not exist",
				}
				return fmt.Errorf("File %s is expected to %s but does %s",
					filename, msg[should], msg[exist])
			}
			return nil
		}
	}

	opLoad := func() func() error {
		return func() error {
			return log.LoadFilters()
		}
	}

	opLoadExpectError := func() func() error {
		return func() error {
			err := opLoad()()
			if err == nil {
				return fmt.Errorf("Expected error loading filters")
			}
			return nil
		}
	}

	opAppendLine := func(session, args string) func() error {
		return func() error {
			return log.AppendLine(session, args)
		}
	}

	opAppendLongLine := func(session, args string, times int) func() error {
		return func() error {
			sb := strings.Builder{}
			for i := 0; i < times; i++ {
				_, _ = sb.WriteString(args)
			}
			return log.AppendLine(session, sb.String())
		}
	}

	opExpectLogfile := func(contentsRe string) func() error {
		re := regexp.MustCompile(contentsRe)
		return func() error {
			data, err := ioutil.ReadFile(logfile)
			if err != nil {
				return err
			}

			if !re.Match(data) {
				return fmt.Errorf("Expected regexp: \"%s\" to match contents:\n%s", contentsRe, string(data))
			}
			return nil
		}
	}

	tests := []struct {
		name             string
		filterData       string
		logData          string
		ops              []opfunc
		expectFilterData []string
	}{
		{"empty defaults",
			contentsForRemoval,
			contentsForRemoval,
			[]opfunc{
				opExists(filterfile, false),
				opExists(logfile, false),
			},
			defaultFilters,
		},

		{"Save defaults",
			contentsForRemoval,
			contentsForRemoval,
			[]opfunc{
				opSave(),
				opExists(filterfile, true),
			},
			defaultFilters,
		},

		{"Load empty filterlist, changes nothing",
			contentsForRemoval,
			contentsForRemoval,
			[]opfunc{
				opExists(filterfile, false),
				opLoad(),
			},
			defaultFilters,
		},
		{"Filterlist with a single line (no newline)",
			"a",
			contentsForRemoval,
			[]opfunc{
				opExists(filterfile, true),
				opLoad(),
			},
			[]string{"a"},
		},
		{"Filterlist with a single line",
			"a\n",
			contentsForRemoval,
			[]opfunc{
				opLoad(),
			},
			[]string{"a"},
		},
		{"Filterlist with two lines (no newline)",
			"a\nb",
			contentsForRemoval,
			[]opfunc{
				opLoad(),
			},
			[]string{"a", "b"},
		},
		{"Filterlist with two lines",
			"a\nb\n",
			contentsForRemoval,
			[]opfunc{
				opLoad(),
			},
			[]string{"a", "b"},
		},
		{"Filterlist invalid regexp",
			`$[\\`,
			contentsForRemoval,
			[]opfunc{
				opLoadExpectError(),
			},
			[]string{},
		},
		{"Filterlist invalid regexp but next line is ok",
			"$[\\\nabc",
			contentsForRemoval,
			[]opfunc{
				opLoadExpectError(),
			},
			[]string{"abc"},
		},
		{"Filterlist with more regexps",
			";.*\n[a-z]$\n",
			contentsForRemoval,
			[]opfunc{
				opLoad(),
			},
			[]string{";.*", "[a-z]$"},
		},
		{"Logfile empty, unchanged",
			contentsForRemoval,
			"",
			[]opfunc{
				opExpectLogfile("^$"),
			},
			defaultFilters,
		},
		{"Logfile created",
			contentsForRemoval,
			contentsForRemoval,
			[]opfunc{
				opAppendLine("ses", "something"),
				opExpectLogfile(`^[0-9]+\tses\tsomething\n$`),
			},
			defaultFilters,
		},
		{"Logfile empty, add line",
			contentsForRemoval,
			"",
			[]opfunc{
				opAppendLine("ses", "something"),
				opExpectLogfile(`^[0-9]+\tses\tsomething\n$`),
			},
			defaultFilters,
		},
		{"Logfile already filled, add line",
			contentsForRemoval,
			"abc\n",
			[]opfunc{
				opAppendLine("ses", "something"),
				opExpectLogfile(`^abc\n[0-9]+\tses\tsomething\n$`),
			},
			defaultFilters,
		},
		{"Logfile default filter out",
			contentsForRemoval,
			"",
			[]opfunc{
				opAppendLine("remove", "ls"),
				opExpectLogfile("^$"),
				opAppendLine("ok", "ls something"),
				opExpectLogfile(`^[0-9]+\tok\tls something\n$`),
			},
			defaultFilters,
		},
		{"Logfile custom filter",
			"abc.*\n",
			"",
			[]opfunc{
				opLoad(),
				opAppendLine("ok", "ls"),
				opExpectLogfile(`^[0-9]+\tok\tls\n$`),
				opAppendLine("remove", "abc something here"),
				opExpectLogfile(`^[0-9]+\tok\tls\n$`),
			},
			[]string{"abc.*"},
		},
		{"Logfile create, add a very long line",
			contentsForRemoval,
			contentsForRemoval,
			[]opfunc{
				opAppendLongLine("ses", "something", 32 * 1024),
				opExpectLogfile(`^[0-9]+\tses\t(something)+\n$`),
			},
			defaultFilters,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := func(err error, args ...interface{}) {
				if err != nil {
					str := fmt.Sprint(args...)
					t.Fatalf("%s: %v", str, err)
				}
			}

			var err error
			err = os.RemoveAll(testdir)
			check(err, "Could not remove test directory")

			err = os.MkdirAll(testdir, 0755)
			check(err, "Could not create test directory")

			if tt.filterData != contentsForRemoval {
				err = ioutil.WriteFile(filterfile, []byte(tt.filterData), 0600)
				check(err, "Could not create filterfile")
			}
			if tt.logData != contentsForRemoval {
				err = ioutil.WriteFile(logfile, []byte(tt.logData), 0600)
				check(err, "Could not create logfile")
			}

			log = CreateLog(logfile, filterfile)

			for i := range tt.ops {
				err = tt.ops[i]()
				check(err, "Op ", i, " failed")
			}

			compare(t, "Filter data differs", tt.expectFilterData, log.Filters)

			err = os.RemoveAll(testdir)
			check(err, "Could not remove test directory")
		})
	}
}
