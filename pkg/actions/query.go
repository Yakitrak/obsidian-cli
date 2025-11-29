package actions

import (
	"fmt"
	"strings"
)

type exprType int

const (
	exprLeaf exprType = iota
	exprAnd
	exprOr
	exprNot
)

// InputExpression represents a boolean expression composed of ListInputs.
type InputExpression struct {
	Type  exprType
	Input *ListInput
	Left  *InputExpression
	Right *InputExpression
}

// ParseInputsWithExpression parses args into ListInputs and a boolean expression tree.
// Operators: AND/OR/NOT (case-insensitive), with parentheses for grouping.
// When no operator is provided between terms, OR is assumed for compatibility.
func ParseInputsWithExpression(args []string) ([]ListInput, *InputExpression, error) {
	if len(args) == 0 {
		return nil, nil, nil
	}

	rawTokens := tokenizeArgs(args)
	if len(rawTokens) == 0 {
		return nil, nil, nil
	}

	tokens := insertImplicitOrs(rawTokens)

	parser := expressionParser{
		tokens: tokens,
	}
	expr, err := parser.parseExpression()
	if err != nil {
		return nil, nil, err
	}
	if parser.pos != len(parser.tokens) {
		return nil, nil, fmt.Errorf("unexpected token %q", parser.tokens[parser.pos])
	}

	return parser.inputs, expr, nil
}

// ParseInputs keeps the original signature while delegating to the expression parser.
func ParseInputs(args []string) ([]ListInput, error) {
	inputs, _, err := ParseInputsWithExpression(args)
	return inputs, err
}

type expressionParser struct {
	tokens []string
	pos    int
	inputs []ListInput
}

func (p *expressionParser) parseExpression() (*InputExpression, error) {
	return p.parseOr()
}

func (p *expressionParser) parseOr() (*InputExpression, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.peekIsOperator("or") {
		p.pos++
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &InputExpression{
			Type:  exprOr,
			Left:  left,
			Right: right,
		}
	}
	return left, nil
}

func (p *expressionParser) parseAnd() (*InputExpression, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.peekIsOperator("and") {
		p.pos++
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &InputExpression{
			Type:  exprAnd,
			Left:  left,
			Right: right,
		}
	}
	return left, nil
}

func (p *expressionParser) parseUnary() (*InputExpression, error) {
	if p.peekIsOperator("not") {
		p.pos++
		child, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &InputExpression{
			Type: exprNot,
			Left: child,
		}, nil
	}
	return p.parsePrimary()
}

func (p *expressionParser) parsePrimary() (*InputExpression, error) {
	if p.match("(") {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if !p.match(")") {
			return nil, fmt.Errorf("expected ')'")
		}
		return expr, nil
	}

	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of expression")
	}

	token := p.tokens[p.pos]
	if classifyToken(token) == "operator" {
		return nil, fmt.Errorf("unexpected operator %q", token)
	}
	p.pos++
	input, err := parseSingleInput(token)
	if err != nil {
		return nil, err
	}
	p.inputs = append(p.inputs, input)

	return &InputExpression{
		Type:  exprLeaf,
		Input: &input,
	}, nil
}

func (p *expressionParser) match(tok string) bool {
	if p.pos < len(p.tokens) && p.tokens[p.pos] == tok {
		p.pos++
		return true
	}
	return false
}

func (p *expressionParser) peekIsOperator(op string) bool {
	if p.pos >= len(p.tokens) {
		return false
	}
	tok := strings.ToLower(p.tokens[p.pos])
	switch op {
	case "and":
		return tok == "and" || tok == "&&"
	case "or":
		return tok == "or" || tok == "||"
	case "not":
		return tok == "not" || tok == "!"
	default:
		return false
	}
}

func tokenizeArgs(args []string) []string {
	var tokens []string
	for _, arg := range args {
		for _, piece := range splitParens(arg) {
			if piece == "" {
				continue
			}
			if containsOperatorWord(piece) {
				for _, field := range strings.Fields(piece) {
					if field == "" {
						continue
					}
					tokens = append(tokens, field)
				}
				continue
			}
			tokens = append(tokens, piece)
		}
	}
	return tokens
}

// containsOperatorWord checks if a string contains boolean operators that need
// to be split into separate tokens. We use heuristics to avoid expensive field
// splitting on simple inputs: a single colon indicates a simple pattern like
// "tag:foo" which won't contain operators. Two or more colons (e.g., "tag:foo AND tag:bar")
// or parentheses/symbolic operators suggest a compound expression worth scanning.
func containsOperatorWord(piece string) bool {
	colonCount := strings.Count(piece, ":")
	hasParens := strings.Contains(piece, "(") || strings.Contains(piece, ")")
	if colonCount < 2 && !hasParens && !strings.Contains(piece, "&&") && !strings.Contains(piece, "||") {
		return false
	}
	fields := strings.Fields(piece)
	for _, f := range fields {
		switch strings.ToUpper(f) {
		case "AND", "OR", "NOT", "&&", "||", "!":
			return true
		}
	}
	return false
}

// splitParens separates leading/trailing parentheses from a token without
// splitting parentheses that appear in the middle of a value.
func splitParens(token string) []string {
	if strings.Contains(token, " ") && !containsOperatorWord(token) {
		return []string{token}
	}

	var tokens []string
	leading := token
	for len(leading) > 0 && (leading[0] == '(' || leading[0] == ')') {
		tokens = append(tokens, string(leading[0]))
		leading = leading[1:]
	}

	trailingParens := ""
	for len(leading) > 0 && (leading[len(leading)-1] == '(' || leading[len(leading)-1] == ')') {
		trailingParens = string(leading[len(leading)-1]) + trailingParens
		leading = leading[:len(leading)-1]
	}

	if strings.TrimSpace(leading) != "" {
		tokens = append(tokens, leading)
	}

	for i := 0; i < len(trailingParens); i++ {
		tokens = append(tokens, string(trailingParens[i]))
	}
	return tokens
}

// insertImplicitOrs inserts OR between adjacent operands/groups when no explicit operator is provided.
func insertImplicitOrs(tokens []string) []string {
	var result []string
	for i, tok := range tokens {
		kind := classifyToken(tok)
		if i > 0 {
			prevKind := classifyToken(tokens[i-1])
			if (prevKind == "operand" || prevKind == "rparen") && (kind == "operand" || kind == "lparen") {
				result = append(result, "OR")
			}
		}
		result = append(result, tok)
	}
	return result
}

func classifyToken(tok string) string {
	switch strings.ToLower(tok) {
	case "and", "or", "not", "&&", "||", "!":
		return "operator"
	case "(":
		return "lparen"
	case ")":
		return "rparen"
	default:
		return "operand"
	}
}

func parseSingleInput(arg string) (ListInput, error) {
	if strings.HasPrefix(arg, "tag:") {
		tag := strings.TrimPrefix(arg, "tag:")
		if strings.HasPrefix(tag, "\"") && strings.HasSuffix(tag, "\"") {
			tag = strings.Trim(tag, "\"")
		}

		if tag == "" || tag == "*" {
			return ListInput{}, fmt.Errorf("invalid tag value in %q: tag cannot be empty or a wildcard (*)", arg)
		}

		return ListInput{
			Type:  InputTypeTag,
			Value: tag,
		}, nil
	}

	if strings.HasPrefix(arg, "find:") {
		searchTerm := strings.TrimPrefix(arg, "find:")
		if strings.HasPrefix(searchTerm, "\"") && strings.HasSuffix(searchTerm, "\"") {
			searchTerm = strings.Trim(searchTerm, "\"")
		}

		if searchTerm == "" || searchTerm == "*" {
			return ListInput{}, fmt.Errorf("invalid find value in %q: find cannot be empty or a wildcard (*)", arg)
		}

		return ListInput{
			Type:  InputTypeFind,
			Value: searchTerm,
		}, nil
	}

	if strings.Contains(arg, ":") {
		parts := strings.SplitN(arg, ":", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		if key == "" || val == "" || val == "*" {
			return ListInput{}, fmt.Errorf("invalid property input %q: both key and value are required", arg)
		}

		return ListInput{
			Type:     InputTypeProperty,
			Value:    val,
			Property: key,
		}, nil
	}

	return ListInput{
		Type:  InputTypeFile,
		Value: arg,
	}, nil
}
