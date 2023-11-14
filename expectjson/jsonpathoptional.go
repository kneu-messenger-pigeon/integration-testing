package expectjson

import (
	"github.com/vitorsalgado/mocha/v3/expect"
)

func JSONPathOptional(p string, matcher expect.Matcher) expect.Matcher {
	jsonMatcher := expect.JSONPath(p, matcher)

	m := expect.Matcher{}
	m.Name = "JSONPathOptional"
	m.DescribeMismatch = jsonMatcher.DescribeMismatch
	m.Matches = func(v any, args expect.Args) (bool, error) {
		matched, err := jsonMatcher.Matches(v, args)
		if err != nil && err.Error() == "could not find a field using provided json path" {
			return false, nil
		}

		return matched, err
	}

	return m
}
