package main

const dotCodeTemplate = `
{{define "layout"}}
digraph "ERD" {
	node [shape="box" style="filled" color="whitesmoke" fillcolor="whitesmoke" fontname="monospace" fontsize="10"];
	edge [dir="both" fontname="sans-serif" fontsize="10"];

	// Entities
	{{range .Entities}}
		{{template "entity" .}}
	{{end}}

	// Relations
	{{range .Relations}}
		{{template "relation" .}}
	{{end}}
}
{{end}}

{{define "entity"}}
	{{.Name}}[label = <
		<table border="0" cellspacing="0" cellpadding="0" bgcolor="gray">
		<tr>
			<td align="left" valign="bottom" cellpadding="4" border="1" color="gray">{{.Name}}{{addParenthesis .AliasName}}</td>
		</tr>
		<tr>
			<td>
				<table bgcolor="white" color="gray" border="1" cellborder="0" cellspacing="1" cellpadding="1">
				{{range .PrimaryColumns}}
					{{template "attribute" .}}
				{{end}}
				<hr />
				{{range .Columns}}
					{{template "attribute" .}}
				{{end}}
				</table>
			</td>
		</tr>
		</table>>];
{{end}}

{{define "attribute"}}
				<tr>
					<td align="left"><font color="red">{{.Name}}</font></td>
					<td align="left">{{.AliasName}}</td>
					<td align="left">{{.FullType}}</td>
				</tr>
{{end}}

{{define "relation"}}
	{{.FromTable}} -> {{.ToTable}}{{toDotCardinarity .Cardinarity}}
{{end}}
`
