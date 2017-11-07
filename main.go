package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"text/template"
)

const (
	// IRCode
	TableDeclearSectionIRCode = iota
	RelationDeclearSectionIRCode
	TableNameDeclearIRCode
	ColumnDeclearIRCode
	RelationDeclearIRCode
	CommentDeclearIRCode
	BadOperationIRCode

	// section flag (for parse to IR)
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
	RegexpColumnPrimaryKey = regexp.MustCompile(`Primary Key`)
	RegexpColumnUnique = regexp.MustCompile(`Unique`)
	RegexpColumnNotNull = regexp.MustCompile(`Not Null`)
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
	case TableDeclearSectionIRCode:
		k = "TableDeclearSectionCode"
	case RelationDeclearSectionIRCode:
		k = "RelationDeclearSectionCode"
	case TableNameDeclearIRCode:
		k = "TableNameDeclearCode"
		desc = c.table.String()
	case ColumnDeclearIRCode:
		k = "ColumnDeclearCode"
		desc = c.column.String()
	case RelationDeclearIRCode:
		k = "RelationDeclearCode"
		desc = c.relation.String()
	case CommentDeclearIRCode:
		k = "CommentDeclearCode"
	case BadOperationIRCode:
		k = "BadOperationCode"
	default:
		log.Fatal("Unknown IRCode to String:", c)
	}

	str := fmt.Sprintf("%d:%s:%s", c.lineNo, k, c.code)
	str += "\n" + desc
	return str
}

// presentation for template
type ViewCode struct {
	Entities  []*EntityViewCode
	Relations []*RelationViewCode
}

type EntityViewCode struct {
	Name           string
	AliasName      string
	PrimaryColumns []*EntityColumnViewCode
	Columns        []*EntityColumnViewCode
}

type EntityColumnViewCode struct {
	Name         string
	AliasName    string
	IsNotNull    bool
	IsUnique     bool
	IsPrimary    bool
	FullType     string
	ForeignTable string
}

type RelationViewCode struct {
	FromTable   string
	FromColumn  string
	ToTable     string
	ToColumn    string
	Cardinarity string
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
		return &IRCode{CommentDeclearIRCode, no, line, nil, nil, nil}, section, nil
	}
	// into table declear section
	if RegexpTableDeclearSection.MatchString(line) == true {
		return &IRCode{TableDeclearSectionIRCode, no, line, nil, nil, nil}, TablesSection, nil
	}
	// into relation declear section
	if RegexpRelationDeclearSection.MatchString(line) == true {
		return &IRCode{RelationDeclearSectionIRCode, no, line, nil, nil, nil}, RelationSection, nil
	}
	// table declear
	if section == TablesSection {
		// table
		if RegexpTableNameDeclear.MatchString(line) == true {
			g := RegexpTableNameDeclear.FindStringSubmatch(line)
			tn := g[1]
			alias := g[3]
			t := &TableIRCode{tn, alias}
			return &IRCode{TableNameDeclearIRCode, no, line, t, nil, nil}, section, nil
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
			return &IRCode{ColumnDeclearIRCode, no, line, nil, nil, c}, section, nil
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
			return &IRCode{RelationDeclearIRCode, no, line, nil, r, nil}, section, nil
		}
	}
	return nil, section, nil
}

func translateViewCode(code []*IRCode) (*ViewCode, error) {
	viewCode := &ViewCode{
		Entities:  make([]*EntityViewCode, 0),
		Relations: make([]*RelationViewCode, 0),
	}

	var entity *EntityViewCode
	for _, c := range code {
		switch c.opKind {
		case TableNameDeclearIRCode:
			if entity != nil {
				viewCode.Entities = append(viewCode.Entities, entity)
			}
			entity = &EntityViewCode{Name: c.table.name, AliasName: c.table.aliasName}
		case ColumnDeclearIRCode:
			col := &EntityColumnViewCode{
				Name:      c.column.name,
				AliasName: c.column.aliasName,
				IsNotNull: c.column.isNotNull,
				IsUnique:  c.column.isUnique,
				IsPrimary: c.column.isPrimary,
				FullType:  c.column.fullType,
			}
			if col.IsPrimary {
				entity.PrimaryColumns = append(entity.PrimaryColumns, col)
			} else {
				entity.Columns = append(entity.Columns, col)
			}
		case RelationDeclearIRCode:
			if entity != nil {
				viewCode.Entities = append(viewCode.Entities, entity)
			}
			r := &RelationViewCode{
				FromTable:   c.relation.fromTable,
				FromColumn:  c.relation.fromColumn,
				ToTable:     c.relation.toTable,
				ToColumn:    c.relation.toColumn,
				Cardinarity: c.relation.cardinarity,
			}
			viewCode.Relations = append(viewCode.Relations, r)
		}
	}
	if entity != nil {
		viewCode.Entities = append(viewCode.Entities, entity)
	}
	setForeignTable(viewCode)
	return viewCode, nil
}

func setForeignTable(vc *ViewCode) {
	for _, r := range vc.Relations {
		for _, e := range vc.Entities {
			if r.FromTable == e.Name {
				for _, c := range e.PrimaryColumns {
					if c.Name == r.FromColumn {
						c.ForeignTable = r.ToTable
					}
				}
				for _, c := range e.Columns {
					if c.Name == r.FromColumn {
						c.ForeignTable = r.ToTable
					}
				}
			}
		}
	}
}

// template utility
func toDotCardinarity(c string) string {
	ret := "["

	// head
	switch c[0] {
	case '1':
		// one
		ret += `arrowhead="tee"`
	case '*':
		// 0 or more
		ret += `arrowhead="crowodot"`
	case '?':
		// 0 or 1
		ret += `arrowhead=""`
	case '+':
		// 1 or more
		ret += `arrowhead="teecrow"`
	default:
		log.Fatalf("unknown cardinarity %s", c)
	}

	ret += " "

	// tail
	switch c[2] {
	case '1':
		// one
		ret += `arrowtail="tee"`
	case '*':
		// 0 or more
		ret += `arrowtail="crowodot"`
	case '?':
		// 1 or 0
		ret += `arrowtail=""`
	case '+':
		// 1 or more
		ret += `arrowtail="teecrow"`
	default:
		log.Fatalf("unknown cardinarity %s", c)
	}

	return ret + "]"
}

// template utility
func addParenthesis(s string) string {
	if s == "" {
		return ""
	}
	return "(" + s + ")"
}

// template utility
func isForeignColumn(c *EntityColumnViewCode) bool {
	if c.ForeignTable != "" {
		return true
	}
	return false
}

// template utility
func isUniqueColumn(c *EntityColumnViewCode) bool {
	return c.IsUnique
}

// template utility
func isNotNullColumn(c *EntityColumnViewCode) bool {
	if c.IsPrimary || c.IsNotNull {
		return true
	}
	return false
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

	// erdot text => IRCode
	code, err := parse(fp)
	if err != nil {
		log.Fatal(err)
	}

	// IRCode => ViewCode
	vc, err := translateViewCode(code)
	if err != nil {
		log.Fatal(err)
	}

	// make dot code from ViewCode
	funcMap := template.FuncMap{
		"toDotCardinarity": toDotCardinarity,
		"addParenthesis":   addParenthesis,
		"isForeignColumn":  isForeignColumn,
		"isUniqueColumn":   isUniqueColumn,
		"isNotNullColumn":  isNotNullColumn,
	}
	tmpl, err := template.New("dotCodeTemplate").Funcs(funcMap).Parse(dotCodeTemplate)
	if err != nil {
		log.Fatal(err)
	}
	err = tmpl.ExecuteTemplate(os.Stdout, "layout", vc)
	if err != nil {
		log.Fatal(err)
	}
}
