package ctxshell

import "regexp"

type Hook struct {
	Before  func() error
	After   func() error
	Pattern string
}

func NewHook(pattern string, before, after func() error) Hook {
	return Hook{
		Before:  before,
		After:   after,
		Pattern: pattern,
	}
}

// Match returns true if the pattern matches the hook's pattern by regexp.
func (h Hook) Match(pattern string) bool {
	re := regexp.MustCompile(h.Pattern)
	return re.MatchString(pattern)
}

func (t *Cshell) AddHook(hook Hook) {
	t.hooks = append(t.hooks, hook)
}

func (t *Cshell) GetHooksByPattern(pattern string) []Hook {
	hooks := []Hook{}
	for _, hook := range t.hooks {
		if hook.Match(pattern) {
			hooks = append(hooks, hook)
		}
	}
	return hooks
}

func (t *Cshell) executeHooksBefore(hooks []Hook) error {
	for _, hook := range hooks {
		if hook.Before != nil {
			err := hook.Before()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Cshell) executeHooksAfter(hooks []Hook) error {
	for _, hook := range hooks {
		if hook.After != nil {
			err := hook.After()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
