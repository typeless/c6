package c6

import "io/ioutil"
import "path/filepath"
import "c6/ast"
import "strconv"

import "fmt"

var fileAstMap map[string]interface{} = map[string]interface{}{}

const (
	UnknownFileType = iota
	ScssFileType
	SassFileType
)

type ParserContext struct {
	ParentRuleSet  *ast.RuleSet
	CurrentRuleSet *ast.RuleSet
}

type ParserError struct {
	ExpectingToken string
	ActualToken    string
}

const debugParser = true

func debug(format string, args ...interface{}) {
	if debugParser {
		fmt.Printf(format+"\n", args...)
	}
}

func (e ParserError) Error() string {
	return fmt.Sprintf("Expecting '%s', but the actual token we got was '%s'.", e.ExpectingToken, e.ActualToken)
}

func getFileTypeByExtension(extension string) uint {
	switch extension {
	case "scss":
		return ScssFileType
	case "sass":
		return SassFileType
	}
	return UnknownFileType
}

type Parser struct {
	Input chan *ast.Token

	// integer for counting token
	Pos         int
	RollbackPos int
	Tokens      []*ast.Token
}

func NewParser() *Parser {
	p := Parser{}
	p.Pos = 0
	p.Tokens = []*ast.Token{}
	return &p
}

func (parser *Parser) parseFile(path string) error {
	ext := filepath.Ext(path)
	filetype := getFileTypeByExtension(ext)
	_ = filetype
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var code string = string(data)
	_ = code
	return nil
}

func (self *Parser) backup() {
	self.Pos--
}

func (self *Parser) remember() {
	self.RollbackPos = self.Pos
}

func (self *Parser) rollback() {
	self.Pos = self.RollbackPos
}

func (self *Parser) accept(tokenType ast.TokenType) bool {
	var tok = self.next()
	if tok.Type == tokenType {
		return true
	}
	self.backup()
	return false
}

func (self *Parser) expect(tokenType ast.TokenType) *ast.Token {
	var tok = self.next()
	if tok.Type != tokenType {
		self.backup()
		panic(fmt.Errorf("Expecting %s, Got %s", tokenType, tok))
	}
	return tok
	return nil

}

func (self *Parser) acceptTypes(types []ast.TokenType) bool {
	var p = self.Pos
	var match = true
	for _, tokType := range types {
		var tok = self.next()
		if tok.Type != tokType {
			match = false
			break
		}
	}
	// restore the position if it doesn't match
	if !match {
		self.Pos = p
	}
	return match
}

func (self *Parser) next() *ast.Token {
	var p = self.Pos
	self.Pos++
	if p < len(self.Tokens) {
		return self.Tokens[p]
	} else {
		if len(self.Tokens) > 1 {
			// get the last token
			var tok = self.Tokens[len(self.Tokens)-1]
			if tok == nil {
				return nil
			}
		}
		token := <-self.Input
		self.Tokens = append(self.Tokens, token)
		return token
	}
	return nil
}

func (self *Parser) peekBy(offset int) *ast.Token {
	if self.Pos+offset < len(self.Tokens) {
		return self.Tokens[self.Pos+offset]
	}
	token := <-self.Input
	for token != nil {
		self.Tokens = append(self.Tokens, token)
		if self.Pos+offset < len(self.Tokens) {
			return self.Tokens[self.Pos+offset]
		}
		token = <-self.Input
	}
	return nil
}

func (self *Parser) advance() {
	self.Pos++
}

func (self *Parser) current() *ast.Token {
	return self.Tokens[self.Pos]
}

func (self *Parser) peek() *ast.Token {
	if self.Pos < len(self.Tokens) {
		return self.Tokens[self.Pos]
	}
	token := <-self.Input
	self.Tokens = append(self.Tokens, token)
	return token
}

func (self *Parser) isSelector() bool {
	var tok = self.peek()
	if tok.Type == ast.T_ID_SELECTOR ||
		tok.Type == ast.T_TYPE_SELECTOR ||
		tok.Type == ast.T_CLASS_SELECTOR ||
		tok.Type == ast.T_PSEUDO_SELECTOR ||
		tok.Type == ast.T_PARENT_SELECTOR {
		return true
	} else if tok.Type == ast.T_BRACKET_LEFT {
		return true
	}
	return false
}

func (self *Parser) eof() bool {
	var tok = self.next()
	self.backup()
	return tok == nil
}

func (parser *Parser) parseScss(code string) *ast.Block {
	l := NewLexerWithString(code)
	l.run()
	parser.Input = l.getOutput()

	block := ast.Block{}
	for !parser.eof() {
		stm := parser.ParseStatement(nil)
		if stm != nil {
			block.AppendStatement(stm)
		}
	}
	return &block
}

func (parser *Parser) ParseStatement(parentRuleSet *ast.RuleSet) ast.Statement {
	var token = parser.peek()

	if token.Type == ast.T_IMPORT {
		return parser.ParseImportStatement()
	} else if token.IsSelector() {
		return parser.ParseRuleSet(parentRuleSet)
	}
	return nil
}

func (parser *Parser) ParseRuleSet(parentRuleSet *ast.RuleSet) ast.Statement {
	var ruleset = ast.RuleSet{}
	var tok = parser.next()

	for tok.IsSelector() {

		switch tok.Type {

		case ast.T_TYPE_SELECTOR:
			sel := ast.TypeSelector{tok.Str}
			ruleset.AppendSelector(sel)

		case ast.T_UNIVERSAL_SELECTOR:
			sel := ast.UniversalSelector{}
			ruleset.AppendSelector(sel)

		case ast.T_ID_SELECTOR:
			sel := ast.IdSelector{tok.Str}
			ruleset.AppendSelector(sel)

		case ast.T_CLASS_SELECTOR:
			sel := ast.ClassSelector{tok.Str}
			ruleset.AppendSelector(sel)

		case ast.T_PARENT_SELECTOR:
			sel := ast.ParentSelector{parentRuleSet}
			ruleset.AppendSelector(sel)

		case ast.T_PSEUDO_SELECTOR:
			sel := ast.PseudoSelector{tok.Str, ""}
			if nextTok := parser.peek(); nextTok.Type == ast.T_LANG_CODE {
				sel.C = nextTok.Str
			}
			ruleset.AppendSelector(sel)
		case ast.T_ADJACENT_SELECTOR:
			ruleset.AppendSelector(ast.AdjacentSelector{})
		case ast.T_CHILD_SELECTOR:
			ruleset.AppendSelector(ast.ChildSelector{})
		case ast.T_DESCENDANT_SELECTOR:
			ruleset.AppendSelector(ast.DescendantSelector{})
		default:
			panic(fmt.Errorf("Unexpected selector token: %+v", tok))
		}
		tok = parser.next()
	}
	parser.backup()

	// parse declaration block
	ruleset.DeclarationBlock = parser.ParseDeclarationBlock(&ruleset)
	return &ruleset
}

/**
This method returns objects with ast.Number interface

works for:

	'10'
	'10' 'px'
	'10' 'em'
	'0.2' 'em'
*/
func (parser *Parser) ReduceNumber() ast.Number {
	// the number token
	var tok = parser.next()

	debug("ReduceNumber => next: %s", tok)

	var tok2 = parser.peek()
	var number ast.Number
	if tok.Type == ast.T_INTEGER {
		i, err := strconv.ParseInt(tok.Str, 10, 64)
		if err != nil {
			panic(err)
		}
		number = ast.NewIntegerNumber(i)
	} else {

		f, err := strconv.ParseFloat(tok.Str, 64)
		if err != nil {
			panic(err)
		}
		number = ast.NewFloatNumber(f)
	}

	if tok2.IsOneOfTypes([]ast.TokenType{ast.T_UNIT_PX, ast.T_UNIT_PT, ast.T_UNIT_CM, ast.T_UNIT_EM, ast.T_UNIT_MM, ast.T_UNIT_REM, ast.T_UNIT_DEG, ast.T_UNIT_PERCENT}) {
		// consume the unit token
		parser.next()
		number.SetUnit(int(tok2.Type))
	}
	return number
}

func (parser *Parser) ReduceFunctionCall() *ast.FunctionCall {
	var identTok = parser.next()

	debug("ReduceFunctionCall => next: %s", identTok)

	var fcall = ast.NewFunctionCall(identTok)

	parser.expect(ast.T_PAREN_START)

	var argTok = parser.peek()
	for argTok.Type != ast.T_PAREN_END {
		var arg = parser.ReduceFactor()
		fcall.AppendArgument(arg)
		debug("ReduceFunctionCall => arg: %+v", arg)

		argTok = parser.peek()
		if argTok.Type == ast.T_COMMA {
			parser.next() // skip comma
			argTok = parser.peek()
		} else if argTok.Type == ast.T_PAREN_END {
			parser.next() // consume ')'
			break
		}
	}
	return fcall
}

func (parser *Parser) ReduceIdent() *ast.Ident {
	var tok = parser.next()
	debug("ReduceIndent => next: %s", tok)

	if tok.Type != ast.T_IDENT {
		panic("Invalid token for ident.")
	}
	return ast.NewIdent(tok.Str, *tok)
}

/**
The ReduceFactor must return an Expression interface compatible object
*/
func (parser *Parser) ReduceFactor() ast.Expression {
	var tok = parser.peek()
	debug("ReduceFactor => peek: %s", tok)

	if tok.Type == ast.T_PAREN_START {

		parser.expect(ast.T_PAREN_START)
		var expr = parser.ReduceExpression()
		parser.expect(ast.T_PAREN_END)
		return expr

	} else if tok.Type == ast.T_INTERPOLATION_START {

		parser.expect(ast.T_INTERPOLATION_START)
		parser.ReduceExpression()
		parser.expect(ast.T_INTERPOLATION_END)
		// TODO:

	} else if tok.Type == ast.T_QQ_STRING || tok.Type == ast.T_Q_STRING {

		tok = parser.next()
		var str = ast.NewString(tok)
		return ast.Expression(str)

	} else if tok.Type == ast.T_INTEGER || tok.Type == ast.T_FLOAT {

		// reduce number
		var number = parser.ReduceNumber()
		return ast.Expression(number)

	} else if tok.Type == ast.T_FUNCTION_NAME {

		var fcall = parser.ReduceFunctionCall()
		return ast.Expression(*fcall)

	} else if tok.Type == ast.T_IDENT {

		var ident = parser.ReduceIdent()
		return ast.Expression(ident)

	} else if tok.Type == ast.T_HEX_COLOR {
		panic("hex color is not implemented yet")
	} else {
		panic(fmt.Errorf("Unknown Token: %s", tok))
	}
	return nil
}

func (parser *Parser) ReduceTerm() ast.Expression {
	debug("ReduceTerm")

	var expr1 = parser.ReduceFactor()

	// see if the next token is '*' or '/'
	var tok = parser.peek()
	if tok.Type == ast.T_MUL || tok.Type == ast.T_DIV {
		var opTok = parser.next()
		var op = ast.NewOp(opTok)
		var expr2 = parser.ReduceFactor()
		return ast.NewBinaryExpression(op, expr1, expr2)
	}
	return expr1
}

/**

We here treat the property values as expressions:

	padding: {expression} {expression} {expression};
	margin: {expression};

Expression := "#{" Expression "}"
			| '+' Expression
			| '-' Expression
			| Term '+' Term
			| Term '-' Term
			| Term
*/
func (parser *Parser) ReduceExpression() ast.Expression {
	debug("ReduceExpression")

	if parser.accept(ast.T_INTERPOLATION_START) {
		debug("ReduceExpression => accept: T_INTERPOLATION_START")

		debug("ReduceExpression => ReduceExpression")
		var expr = parser.ReduceExpression()

		parser.expect(ast.T_INTERPOLATION_END)
		debug("ReduceExpression => expect: T_INTERPOLATION_START")
		return expr
	}

	// plus or minus. this creates an unary expression that holds the later term.
	var tok = parser.peek()
	if tok.Type == ast.T_PLUS || tok.Type == ast.T_MINUS {
		parser.next()
		var op = ast.NewOp(tok)
		var expr = parser.ReduceExpression()
		return ast.NewUnaryExpression(op, expr)
	}

	var leftTerm = parser.ReduceTerm()
	var rightTok = parser.peek()
	if rightTok.Type == ast.T_PLUS {
		parser.next()
		var op = ast.NewOp(rightTok)
		var rightTerm = parser.ReduceTerm()
		return ast.NewBinaryExpression(op, leftTerm, rightTerm)
	} else if rightTok.Type == ast.T_MINUS {
		parser.next()
		var op = ast.NewOp(rightTok)
		var rightTerm = parser.ReduceTerm()
		return ast.NewBinaryExpression(op, leftTerm, rightTerm)
	} else {
		return ast.NewUnaryExpression(nil, leftTerm)
	}
	return nil
}

/**
The returned Expression is an interface
*/
func (parser *Parser) ParsePropertyListValue(parentRuleSet *ast.RuleSet, property *ast.Property) []ast.Expression {
	var tok = parser.peek()
	var valueList []ast.Expression = []ast.Expression{}

	// a list can end with ';' or '}'
	for tok.Type != ast.T_SEMICOLON && tok.Type != ast.T_BRACE_END {
		var expr = parser.ReduceExpression()
		tok = parser.peek()

		// see if the next is a comma
		if tok.Type == ast.T_COMMA {
			parser.next()
			tok = parser.peek()
		}
		if expr != nil {
			valueList = append(valueList, expr)
		}
	}
	parser.accept(ast.T_SEMICOLON)
	return valueList
}

func (parser *Parser) ParseDeclarationBlock(parentRuleSet *ast.RuleSet) *ast.DeclarationBlock {
	var declBlock = ast.DeclarationBlock{}

	var tok = parser.next() // should be '{'
	if tok.Type != ast.T_BRACE_START {
		panic(ParserError{"{", tok.Str})
	}

	tok = parser.next()
	for tok != nil && tok.Type != ast.T_BRACE_END {

		if tok.Type == ast.T_PROPERTY_NAME_TOKEN {
			parser.expect(ast.T_COLON)

			var property = ast.NewProperty(tok)
			var valueList = parser.ParsePropertyListValue(parentRuleSet, property)
			property.Values = valueList
			declBlock.Append(property)
			_ = property

		} else if tok.IsSelector() {
			// parse subrule
			panic("subselector unimplemented")
		} else {
			panic("unexpected token")
		}

		tok = parser.next()
	}

	return &declBlock
}

func (parser *Parser) ParseImportStatement() ast.Statement {
	// skip the ast.T_IMPORT token
	var tok = parser.next()

	// Create the import statement node
	var rule = ast.ImportStatement{}

	tok = parser.peek()
	// expecting url(..)
	if tok.Type == ast.T_IDENT {
		parser.advance()

		if tok.Str != "url" {
			panic("invalid function for @import rule.")
		}

		if tok = parser.next(); tok.Type != ast.T_PAREN_START {
			panic("expecting parenthesis after url")
		}

		tok = parser.next()
		rule.Url = ast.Url(tok.Str)

		if tok = parser.next(); tok.Type != ast.T_PAREN_END {
			panic("expecting parenthesis after url")
		}

	} else if tok.IsString() {
		parser.advance()
		rule.Url = ast.RelativeUrl(tok.Str)
	}

	/*
		TODO: parse media query for something like:

		@import url(color.css) screen and (color);
		@import url('landscape.css') screen and (orientation:landscape);
		@import url("bluish.css") projection, tv;
	*/
	tok = parser.peek()
	if tok.Type == ast.T_MEDIA {
		parser.advance()
		rule.MediaList = append(rule.MediaList, tok.Str)
	}

	// must be ast.T_SEMICOLON
	tok = parser.next()
	if tok.Type != ast.T_SEMICOLON {
		panic(ParserError{";", tok.Str})
	}
	return &rule
}
