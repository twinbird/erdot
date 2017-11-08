# erdot

'erdot lang' to 'dot lang' translator.

## install

## Usage

```console
$ erdot [erdot file] | dot -Tpng -o output.png
```

## what's erdot lang

### sample

```
// This is a comment

# Tables

// Indent is '[tab]' or '2 or more [space]'
TableName[(alias name)]
	ColumnName[(alias name)] Type [Unique|PrimaryKey|NotNull|Default(value)]
	ColumnName[(alias name)] Type [Unique|PrimaryKey|NotNul|Default(value)l]
	ColumnName[(alias name)] Type [Unique|PrimaryKey|NotNul|Default(value)l]

# Relations

// one to one
ChildTableName.ColumnName 1-1 ParentTableName.ColumnName 

// 0 or 1 to one
ChildTableName.ColumnName ?-1 ParentTableName.ColumnName 

// 0 or more to one
ChildTableName.ColumnName *-1 ParentTableName.ColumnName 

// 1 or more to one
ChildTableName.ColumnName +-1 ParentTableName.ColumnName 
```

### Cardinality

| Cardinality   | Syntax   |
| ------------- | -------- |
| 0 or 1        | ?        |
| exactly 1     | 1        |
| 0 or more     | *        |
| 1 or more     | +        |
