package parser

import (
	"fmt"

	"github.com/grantwforsythe/monkeylang/pkg/ast"
	"github.com/grantwforsythe/monkeylang/pkg/lexer"
	"github.com/grantwforsythe/monkeylang/pkg/token"
	"github.com/grantwforsythe/monkeylang/pkg/utils"
)

type Parser struct {
	l         *lexer.Lexer
	currToken token.Token
	peekToken token.Token
	errors    []utils.ErrorString
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []utils.ErrorString{}}

	// Read the first two tokens so both curr and peek are set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []utils.ErrorString {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := utils.ErrorString{s: fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)}
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.currToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.currToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// TODO: Fix this. We are skipping the expression until we encounter a semicolon
	for p.currToken.Type != token.SEMICOLON {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.currToken}

	p.nextToken()

	// TODO: Fix this. We are skipping the expression until we encounter a semicolon
	for p.currToken.Type != token.SEMICOLON {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.currToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return nil
	}
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.currToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}