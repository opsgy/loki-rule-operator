package v1beta1

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/grafana/loki/pkg/logql"
	"github.com/prometheus/prometheus/pkg/labels"
)

// ValidateExpressions validates the expressions of the rules
func (lokiRule *GlobalLokiRule) ValidateExpressions() (*GlobalLokiRuleSpec, error) {
	specCopy := lokiRule.Spec.DeepCopy()
	for _, group := range specCopy.Groups {
		for _, rule := range group.Rules {
			alertName := rule.Expr
			if rule.Alert != "" {
				alertName = rule.Alert
			}

			// validate expression
			expr, err := logql.ParseExpr(rule.Expr)
			if err != nil {
				return nil, fmt.Errorf("%s: %s", alertName, err.Error())
			}

			rule.Expr = expr.String()
		}
	}

	return specCopy, nil
}

// ValidateExpressions validates the expressions of the rules
func (lokiRule *LokiRule) ValidateExpressions() (*LokiRuleSpec, error) {
	specCopy := lokiRule.Spec.DeepCopy()
	for _, group := range specCopy.Groups {
		for _, rule := range group.Rules {
			alertName := rule.Expr
			if rule.Alert != "" {
				alertName = rule.Alert
			}

			// validate expression
			expr, err := logql.ParseExpr(rule.Expr)
			if err != nil {
				return nil, fmt.Errorf("%s: %s", alertName, err.Error())
			}
			if err := enforceNode(lokiRule.Namespace, expr); err != nil {
				return nil, fmt.Errorf("%s: %s", alertName, err.Error())
			}

			rule.Expr = expr.String()
		}
	}

	return specCopy, nil
}

// EnforceNode walks the given node recursively
// and enforces the given label enforcer on it.
//
// Whenever a parser.MatrixSelector or parser.VectorSelector AST node is found,
// their label enforcer is being potentially modified.
// If a node's label matcher has the same name as a label matcher
// of the given enforcer, then it will be replaced.
func enforceNode(ns string, node logql.Expr) error {
	t := getType(node)
	switch t {
	case "*matchersExpr":
		rs := reflect.ValueOf(node).Elem()
		rf := rs.FieldByName("matchers")
		re := reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
		matchers := re.Interface().([]*labels.Matcher)
		enforcedMatchers, err := enforceMatchers(ns, matchers)
		if err != nil {
			return err
		}
		re.Set(reflect.ValueOf(enforcedMatchers))

	case "*filterExpr":
		rs := reflect.ValueOf(node).Elem()
		rf := rs.FieldByName("left")
		re := reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
		expr := re.Interface().(logql.Expr)
		if err := enforceNode(ns, expr); err != nil {
			return err
		}

	case "*rangeAggregationExpr":
		rs := reflect.ValueOf(node).Elem()
		rf := rs.FieldByName("left")
		re := reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
		left := re.Interface()

		leftType := getType(left)
		if leftType == "*logRange" {
			if err := enforceRange(ns, left); err != nil {
				return err
			}
		} else {
			if err := enforceNode(ns, left.(logql.Expr)); err != nil {
				return err
			}
		}

	case "*vectorAggregationExpr":
		rs := reflect.ValueOf(node).Elem()
		rf := rs.FieldByName("left")
		re := reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
		left := re.Interface()

		leftType := getType(left)
		if leftType == "*logRange" {
			if err := enforceRange(ns, left); err != nil {
				return err
			}
		} else {
			if err := enforceNode(ns, left.(logql.Expr)); err != nil {
				return err
			}
		}

	case "*pipelineExpr":
		rs := reflect.ValueOf(node).Elem()
		rf := rs.FieldByName("left")
		re := reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
		left := re.Interface()

		leftType := getType(left)
		if leftType == "*logRange" {
			if err := enforceRange(ns, left); err != nil {
				return err
			}
		} else {
			if err := enforceNode(ns, left.(logql.Expr)); err != nil {
				return err
			}
		}

	case "*binOpExpr":
		rs := reflect.ValueOf(node).Elem()
		rf := rs.FieldByName("SampleExpr")
		re := reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
		left := re.Interface()

		leftType := getType(left)
		if leftType == "*logRange" {
			if err := enforceRange(ns, left); err != nil {
				return err
			}
		} else {
			if err := enforceNode(ns, left.(logql.Expr)); err != nil {
				return err
			}
		}

	default:
		panic(fmt.Errorf("parser.Walk: unhandled node type %s", t))
	}

	return nil
}

func enforceRange(ns string, logRange interface{}) error {
	rs := reflect.ValueOf(logRange).Elem()
	rf := rs.FieldByName("left")
	re := reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	left := re.Interface()
	leftType := getType(left)
	if leftType == "*logRange" {
		if err := enforceRange(ns, left); err != nil {
			return err
		}
	} else {
		if err := enforceNode(ns, left.(logql.Expr)); err != nil {
			return err
		}
	}
	return nil
}

func enforceMatchers(ns string, targets []*labels.Matcher) ([]*labels.Matcher, error) {
	var res []*labels.Matcher

	for _, target := range targets {
		if target.Name == "namespace" {
			if target.Type != labels.MatchEqual || target.Value != ns {
				return nil, fmt.Errorf("'namespace' selector should equals '%s'", ns)
			}
		} else {
			res = append(res, target)
		}
	}

	nsMatcher := &labels.Matcher{
		Name:  "namespace",
		Type:  labels.MatchEqual,
		Value: ns,
	}
	res = append(res, nsMatcher)

	return res, nil
}

func getType(myvar interface{}) string {
	t := reflect.TypeOf(myvar)
	if t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	}
	return t.Name()
}
