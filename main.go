package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
)

const (
	TableDeclearSectionCode = iota
	RelationDeclearSectionCode
	TableNameDeclearCode
	ColumnDeclearCode
	RelationDeclearCode
	CommentDeclearCode
	BadOperationCode

	NoOperationSection = iota
	TablesSection
	RelationSection
)

var (
	RegexpCommentDeclear         *regexp.Regexp
	RegexpTableDeclearSection    *regexp.Regexp
	RegexpRelationDeclearSection *regexp.Regexp
	RegexpTableNameDeclear       *regexp.Regexp
	RegexpColumnDeclear          *regexp.Regexp
	RegexpRelationDeclear        *regexp.Regexp
	RegexpColumnPrimaryKey       *regexp.Regexp
	RegexpColumnUnique           *regexp.Regexp
	RegexpColumnNotNull          *regexp.Regexp
)

func init() {
	RegexpTableDeclearSection = regexp.MustCompile(`^#\s+Tables\s*`)
	RegexpRelationDeclearSection = regexp.MustCompile(`^#\s+Relations\s*`)
	RegexpTableNameDeclear = regexp.MustCompile(`^([^\s]+)(\s+\((.+?)\))?`)
	RegexpColumnDeclear = regexp.MustCompile(`(\t|  +)(.+?)\s+(\((.+?)\)\s+)?(.+)`)
	RegexpCommentDeclear = regexp.MustCompile(`^//.*`)
	RegexpRelationDeclear = regexp.MustCompile(`^(.+)\.(.+)\s(.+)\s(.+)\.(.+)`)
	RegexpColumnPrimaryKey = regexp.MustCompile(`PrimaryKey`)
	RegexpColumnUnique = regexp.MustCompile(`Unique`)
	RegexpColumnNotNull = regexp.MustCompile(`NotNull`)
}

type TableIRCode struct {
	name      string
	aliasName string
}

func (t *TableIRCode) String() string {
	return fmt.Sprintf("%s(%s)", t.name, t.aliasName)
}

type ColumnIRCode struct {
	name      string
	aliasName string
	isNotNull bool
	isUnique  bool
	isPrimary bool
	fullType  string
}

func (c *ColumnIRCode) String() string {
	constraint := ""
	if c.isNotNull == true {
		constraint += "NOT NULL"
	}
	if c.isUnique == true {
		constraint += "UNIQUE"
	}
	if c.isPrimary == true {
		constraint += "PRIMARY KEY"
	}
	return fmt.Sprintf("\t%s %s %s: %s", c.name, c.aliasName, c.fullType, constraint)
}

type RelationIRCode struct {
	fromTable   string
	fromColumn  string
	toTable     string
	toColumn    string
	cardinarity string
}

func (r *RelationIRCode) String() string {
	return fmt.Sprintf("%s.%s %s %s.%s", r.fromTable, r.fromColumn, r.cardinarity, r.toTable, r.toColumn)
}

type IRCode struct {
	opKind   int
	lineNo   int
	code     string
	table    *TableIRCode
	relation *RelationIRCode
	column   *ColumnIRCode
}

func (c *IRCode) String() string {
	var k string
	var desc string

	switch c.opKind {
	case TableDeclearSectionCode:
		k = "TableDeclearSectionCode"
	case RelationDeclearSectionCode:
		k = "RelationDeclearSectionCode"
	case TableNameDeclearCode:
		k = "TableNameDeclearCode"
		desc = c.table.String()
	case ColumnDeclearCode:
		k = "ColumnDeclearCode"
		desc = c.column.String()
	case RelationDeclearCode:
		k = "RelationDeclearCode"
		desc = c.relation.String()
	case CommentDeclearCode:
		k = "CommentDeclearCode"
	case BadOperationCode:
		k = "BadOperationCode"
	default:
		log.Fatal("Unknown IRCode to String:", c)
	}

	str := fmt.Sprintf("%d:%s:%s", c.lineNo, k, c.code)
	str += "\n" + desc
	return str
}

func main() {
	var fp *os.File
	var err error

	if len(os.Args) < 2 {
		fp = os.Stdin
	} else {
		fp, err = os.Open(os.Args[1])
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		defer fp.Close()
	}
	code, err := parse(fp)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(code); i++ {
		fmt.Println(code[i])
	}
}

func parse(r io.Reader) ([]*IRCode, error) {
	var code []*IRCode
	var l int
	section := NoOperationSection
	reader := bufio.NewReaderSize(r, 4096)

	for {
		l += 1
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		c, sec, err := lineToIRCode(string(line), l, section)
		if err != nil {
			return nil, err
		}
		section = sec
		if c != nil {
			code = append(code, c)
		}
	}
	return code, nil
}

func lineToIRCode(line string, no int, section int) (*IRCode, int, error) {
	// comments
	if RegexpCommentDeclear.MatchString(line) == true {
		return &IRCode{CommentDeclearCode, no, line, nil, nil, nil}, section, nil
	}
	// into table declear section
	if RegexpTableDeclearSection.MatchString(line) == true {
		return &IRCode{TableDeclearSectionCode, no, line, nil, nil, nil}, TablesSection, nil
	}
	// into relation declear section
	if RegexpRelationDeclearSection.MatchString(line) == true {
		return &IRCode{RelationDeclearSectionCode, no, line, nil, nil, nil}, RelationSection, nil
	}
	// table declear
	if section == TablesSection {
		// table
		if RegexpTableNameDeclear.MatchString(line) == true {
			g := RegexpTableNameDeclear.FindStringSubmatch(line)
			tn := g[1]
			alias := g[3]
			t := &TableIRCode{tn, alias}
			return &IRCode{TableNameDeclearCode, no, line, t, nil, nil}, section, nil
		}
		// column
		if RegexpColumnDeclear.MatchString(line) == true {
			g := RegexpColumnDeclear.FindStringSubmatch(line)
			c := &ColumnIRCode{
				name:      g[2],
				aliasName: g[4],
				fullType:  g[5],
			}
			// constraint check
			if RegexpColumnPrimaryKey.MatchString(c.fullType) == true {
				c.isPrimary = true
			}
			if RegexpColumnUnique.MatchString(c.fullType) == true {
				c.isUnique = true
			}
			if RegexpColumnNotNull.MatchString(c.fullType) == true {
				c.isNotNull = true
			}
			return &IRCode{ColumnDeclearCode, no, line, nil, nil, c}, section, nil
		}
	}
	// relation declear
	if section == RelationSection {
		if RegexpRelationDeclear.MatchString(line) == true {
			g := RegexpRelationDeclear.FindStringSubmatch(line)
			tn1 := g[1]
			c1 := g[2]
			card := g[3]
			tn2 := g[4]
			c2 := g[5]
			r := &RelationIRCode{tn1, c1, tn2, c2, card}
			return &IRCode{RelationDeclearCode, no, line, nil, r, nil}, section, nil
		}
	}
	return nil, section, nil
}
