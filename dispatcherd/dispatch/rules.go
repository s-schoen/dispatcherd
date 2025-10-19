package dispatch

import (
	"context"
	"fmt"
	"log/slog"
)

var DefaultDispatcher = []string{""}

type RuleOperator string

const (
	EQUALS RuleOperator = "eq"
)

type RuleMatch struct {
	TagName  string
	Operator RuleOperator
	Value    string
}

type Rule struct {
	ID             string
	DispatcherName string
	Match          []RuleMatch
}

type RuleEngine interface {
	ProcessMessage(ctx context.Context, msg *Message) ([]string, error)
}

type DefaultRuleEngine struct {
	logger *slog.Logger
	rules  []Rule
}

func NewRuleEngine(logger *slog.Logger) *DefaultRuleEngine {
	return &DefaultRuleEngine{
		logger: logger,
		rules:  make([]Rule, 0),
	}
}

func (e *DefaultRuleEngine) SetRules(rules []Rule) {
	e.rules = rules
}

func (e *DefaultRuleEngine) ProcessMessage(ctx context.Context, msg *Message) ([]string, error) {
	if msg.Tags == nil {
		// no tags, so matching does not make sense
		return DefaultDispatcher, nil
	}

	var dispatchers []string
	for _, rule := range e.rules {
		e.logger.DebugContext(ctx, fmt.Sprintf("validating rule '%s'", rule.ID))
		if e.ruleMatch(rule, msg.Tags) {
			e.logger.DebugContext(ctx, fmt.Sprintf("matched rule '%s'", rule.ID))
			dispatchers = append(dispatchers, rule.DispatcherName)
		}
	}

	if len(dispatchers) == 0 {
		// no match, return default
		return DefaultDispatcher, nil
	}

	return dispatchers, nil
}

func (e *DefaultRuleEngine) ruleMatch(rule Rule, tags map[string]string) bool {
	matched := false
	for _, match := range rule.Match {
		val, ok := tags[match.TagName]
		if !ok {
			// required tag does not exist
			return false
		}

		switch match.Operator {
		case EQUALS:
			if val == match.Value {
				matched = true
			}
		}
	}

	return matched
}
