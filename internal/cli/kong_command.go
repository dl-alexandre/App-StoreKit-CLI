package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kong"
)

// Command is the command description used by the CLI handlers. Kong builds and
// parses the actual command grammar from this tree at execution time.
type Command struct {
	Use               string
	Short             string
	Aliases           []string
	RunE              func(*Command, []string) error
	Run               func(*Command, []string)
	PreRunE           func(*Command, []string) error
	PersistentPreRunE func(*Command, []string) error
	Args              func(*Command, []string) error

	parent   *Command
	children []*Command
	flags    *flagSet
	root     *Command
	ctx      context.Context
	stdout   io.Writer
	stderr   io.Writer
	args     []string
	options  *rootOptions
	app      *App
}

func newCommand() *Command { return &Command{flags: &flagSet{}} }

func (c *Command) AddCommand(children ...*Command) {
	for _, child := range children {
		child.parent = c
		c.children = append(c.children, child)
	}
}

func (c *Command) Flags() *flagSet {
	if c.flags == nil {
		c.flags = &flagSet{}
	}
	return c.flags
}

func (c *Command) PersistentFlags() *flagSet { return c.Flags() }
func (c *Command) Root() *Command {
	if c.root != nil {
		return c.root
	}
	for c.parent != nil {
		c = c.parent
	}
	return c
}
func (c *Command) Context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}
func (c *Command) SetContext(ctx context.Context) { c.ctx = ctx }
func (c *Command) OutOrStdout() io.Writer {
	if c.stdout != nil {
		return c.stdout
	}
	return os.Stdout
}
func (c *Command) CommandPath() string {
	parts := []string{}
	for p := c; p != nil; p = p.parent {
		if p.Use != "" {
			parts = append(parts, p.Use)
		}
	}
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, " ")
}
func (c *Command) SetArgs(args []string) { c.args = args }
func (c *Command) SetOut(w io.Writer)    { c.stdout = w }
func (c *Command) SetErr(w io.Writer)    { c.stderr = w }
func (c *Command) RegisterFlagCompletionFunc(name string, complete any) error {
	for _, flag := range c.Flags().defs {
		if flag.name != name {
			continue
		}
		if fn, ok := complete.(func(*Command, []string, string) ([]string, int)); ok {
			flag.completions, _ = fn(c, nil, "")
		}
	}
	return nil
}
func (c *Command) GenBashCompletion(w io.Writer) error { return c.generateCompletion(w, "bash") }
func (c *Command) GenZshCompletion(w io.Writer) error  { return c.generateCompletion(w, "zsh") }
func (c *Command) GenFishCompletion(w io.Writer, _ bool) error {
	return c.generateCompletion(w, "fish")
}
func (c *Command) GenPowerShellCompletionWithDesc(w io.Writer) error {
	return c.generateCompletion(w, "powershell")
}

func (c *Command) generateCompletion(w io.Writer, shell string) error {
	root := c.Root()
	words := []string{}
	var walk func(*Command)
	walk = func(node *Command) {
		if node != root {
			words = append(words, node.Use)
			words = append(words, node.Aliases...)
		}
		for _, f := range node.Flags().defs {
			words = append(words, "--"+f.name)
			words = append(words, f.completions...)
		}
		for _, child := range node.children {
			walk(child)
		}
	}
	walk(root)
	words = append(words, "--help", "--config", "--format", "--jq", "--table-columns", "--request-id", "--timeout", "--debug", "--issuer-id", "--key-id", "--bundle-id", "--private-key-path", "--private-key", "--env", "--max-retries", "--retry-backoff-ms")
	words = uniqueWords(words)
	wordList := strings.Join(words, " ")
	switch shell {
	case "bash":
		_, err := fmt.Fprintf(w, "_ask_completions() {\n  COMPREPLY=( $(compgen -W %q -- \"${COMP_WORDS[COMP_CWORD]}\") )\n}\ncomplete -F _ask_completions ask\n", wordList)
		return err
	case "zsh":
		_, err := fmt.Fprintf(w, "#compdef ask\n_arguments '*: :(%s)'\n", wordList)
		return err
	case "fish":
		_, err := fmt.Fprintf(w, "complete -c ask -f -a '%s'\n", strings.ReplaceAll(wordList, "'", "\\'"))
		return err
	case "powershell":
		_, err := fmt.Fprintf(w, "Register-ArgumentCompleter -Native -CommandName ask -ScriptBlock { param($wordToComplete) '%s'.Split(' ') | Where-Object { $_ -like \"$wordToComplete*\" } }\n", strings.ReplaceAll(wordList, "'", "''"))
		return err
	default:
		return fmt.Errorf("unsupported completion shell %q", shell)
	}
}

func uniqueWords(words []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(words))
	for _, word := range words {
		if word != "" && !seen[word] {
			seen[word] = true
			out = append(out, word)
		}
	}
	return out
}

func (c *Command) Execute() error { return c.execute(c.args) }

func (c *Command) execute(args []string) error {
	root := c.Root()
	root.root = root
	root.args = append([]string(nil), args...)
	model, values, err := buildKongModel(root)
	if err != nil {
		return err
	}
	k := kong.Must(model, kong.Name(root.Use), kong.Description(root.Short), kong.Writers(root.OutOrStdout(), root.stderrOrStderr()), kong.Exit(func(int) {}))
	_, err = k.Parse(args)
	if err != nil {
		if hasHelpFlag(args) {
			return nil
		}
		return err
	}
	if hasHelpFlag(args) {
		return nil
	}
	applyFlagValues(values)
	selected := commandFromArgs(root, args)
	if selected == nil {
		return errors.New("no command selected")
	}
	selected.root = root
	selected.ctx = context.Background()
	selected.stdout = root.stdout
	selected.stderr = root.stderr
	selected.options = root.options
	if root.PersistentPreRunE != nil {
		if err := root.PersistentPreRunE(selected, nil); err != nil {
			return err
		}
	}
	if selected.Args != nil {
		if err := selected.Args(selected, nil); err != nil {
			return err
		}
	}
	if selected.PreRunE != nil {
		if err := selected.PreRunE(selected, nil); err != nil {
			return err
		}
	}
	if selected.RunE != nil {
		return selected.RunE(selected, nil)
	}
	if selected.Run != nil {
		selected.Run(selected, nil)
	}
	return nil
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func (c *Command) stderrOrStderr() io.Writer {
	if c.stderr != nil {
		return c.stderr
	}
	return os.Stderr
}

type flagDef struct {
	name, help   string
	ptr          any
	kind         string
	defaultValue string
	repeat       bool
	completions  []string
}
type flagSet struct{ defs []*flagDef }

func (f *flagSet) add(ptr any, name, help, kind, def string, repeat bool) {
	f.defs = append(f.defs, &flagDef{name: name, help: help, ptr: ptr, kind: kind, defaultValue: def, repeat: repeat})
}
func (f *flagSet) StringVar(p *string, n, d, h string) { f.add(p, n, h, "string", d, false) }
func (f *flagSet) BoolVar(p *bool, n string, d bool, h string) {
	f.add(p, n, h, "bool", strconv.FormatBool(d), false)
}
func (f *flagSet) IntVar(p *int, n string, d int, h string) {
	f.add(p, n, h, "int", strconv.Itoa(d), false)
}
func (f *flagSet) DurationVar(p *time.Duration, n string, d time.Duration, h string) {
	f.add(p, n, h, "duration", d.String(), false)
}
func (f *flagSet) StringArrayVar(p *[]string, n string, d []string, h string) {
	if d != nil {
		*p = append([]string(nil), d...)
	}
	f.add(p, n, h, "string-array", "", true)
}
func (f *flagSet) StringSliceVar(p *[]string, n string, d []string, h string) {
	if d != nil {
		*p = append([]string(nil), d...)
	}
	f.add(p, n, h, "string-slice", "", false)
}
func (f *flagSet) Var(v interface{ Set(string) error }, n, h string) {
	f.add(v, n, h, "optional-bool", "", false)
}
func (f *flagSet) GetString(name string) (string, error) {
	for _, d := range f.defs {
		if d.name == name {
			if p, ok := d.ptr.(*string); ok {
				return *p, nil
			}
		}
	}
	return "", fmt.Errorf("unknown flag %q", name)
}

type kongFieldValue struct {
	flag  *flagDef
	path  []int
	value reflect.Value
}

func buildKongModel(root *Command) (any, []kongFieldValue, error) {
	selected := commandFromArgs(root, root.args)
	globalNames := map[string]bool{}
	for _, d := range root.Flags().defs {
		globalNames[d.name] = true
	}
	selectedNames := map[string]bool{}
	for _, d := range selected.Flags().defs {
		selectedNames[d.name] = true
	}
	rootDefs := make([]*flagDef, 0, len(root.Flags().defs))
	for _, d := range root.Flags().defs {
		if !selectedNames[d.name] {
			rootDefs = append(rootDefs, d)
		}
	}
	fields := []reflect.StructField{}
	values := []kongFieldValue{}
	for i, d := range rootDefs {
		f, err := flagStructField(d, fmt.Sprintf("Global%d", i))
		if err != nil {
			return nil, nil, err
		}
		fields = append(fields, f)
	}
	for i, child := range root.children {
		t, nested, err := commandType(child, fmt.Sprintf("Cmd%d", i), selected, globalNames)
		if err != nil {
			return nil, nil, err
		}
		childIndex := len(fields)
		fields = append(fields, reflect.StructField{Name: fmt.Sprintf("Cmd%d", i), Type: t, Tag: commandTag(child)})
		for _, item := range nested {
			item.path = append([]int{childIndex}, item.path...)
			values = append(values, item)
		}
	}
	rootType := reflect.StructOf(fields)
	rootValue := reflect.New(rootType)
	for i, d := range rootDefs {
		values = append(values, kongFieldValue{flag: d, path: []int{i}})
	}
	// Retain the backing value on the root so that Kong mutates the model we parse.
	return rootValue.Interface(), bindKongTargets(rootValue, values), nil
}

func commandType(cmd *Command, fieldName string, selected *Command, globalNames map[string]bool) (reflect.Type, []kongFieldValue, error) {
	fields := []reflect.StructField{}
	defs := make([]*flagDef, 0, len(cmd.Flags().defs))
	for _, d := range cmd.Flags().defs {
		if cmd == selected || !globalNames[d.name] {
			defs = append(defs, d)
		}
	}
	for i, d := range defs {
		f, err := flagStructField(d, fmt.Sprintf("Flag%d", i))
		if err != nil {
			return nil, nil, err
		}
		fields = append(fields, f)
	}
	values := []kongFieldValue{}
	for i, child := range cmd.children {
		t, nested, err := commandType(child, fmt.Sprintf("Cmd%d", i), selected, globalNames)
		if err != nil {
			return nil, nil, err
		}
		childIndex := len(fields)
		fields = append(fields, reflect.StructField{Name: fmt.Sprintf("Cmd%d", i), Type: t, Tag: commandTag(child)})
		for _, item := range nested {
			item.path = append([]int{childIndex}, item.path...)
			values = append(values, item)
		}
	}
	for i, d := range defs {
		values = append(values, kongFieldValue{flag: d, path: []int{i}})
	}
	t := reflect.StructOf(fields)
	return t, values, nil
}

func bindKongTargets(root reflect.Value, values []kongFieldValue) []kongFieldValue {
	for i := range values {
		v := root.Elem()
		for _, index := range values[i].path {
			v = v.Field(index)
		}
		values[i].value = v
	}
	return values
}

func commandTag(c *Command) reflect.StructTag {
	return reflect.StructTag(fmt.Sprintf("cmd:\"\" name:%q help:%q aliases:%q", c.Use, c.Short, strings.Join(c.Aliases, ",")))
}

func flagStructField(d *flagDef, name string) (reflect.StructField, error) {
	var typ reflect.Type
	switch d.kind {
	case "string":
		typ = reflect.TypeOf("")
	case "bool":
		typ = reflect.TypeOf(false)
	case "int":
		typ = reflect.TypeOf(int(0))
	case "duration":
		typ = reflect.TypeOf(time.Duration(0))
	case "string-array", "string-slice":
		typ = reflect.TypeOf([]string{})
	case "optional-bool":
		typ = reflect.TypeOf((*bool)(nil))
	default:
		return reflect.StructField{}, fmt.Errorf("unsupported flag kind %q", d.kind)
	}
	tag := fmt.Sprintf("name:%q help:%q", d.name, d.help)
	if d.defaultValue != "" {
		tag += fmt.Sprintf(" default:%q", d.defaultValue)
	}
	if d.kind == "string-array" {
		tag += ` sep:"\x00"`
	}
	if d.kind == "string-slice" {
		tag += ` sep:","`
	}
	return reflect.StructField{Name: name, Type: typ, Tag: reflect.StructTag(tag)}, nil
}

func applyFlagValues(values []kongFieldValue) {
	for _, item := range values {
		if !item.value.IsValid() {
			continue
		}
		value := item.value.Interface()
		switch p := item.flag.ptr.(type) {
		case *string:
			*p = value.(string)
		case *bool:
			*p = value.(bool)
		case *int:
			*p = value.(int)
		case *time.Duration:
			*p = value.(time.Duration)
		case *[]string:
			*p = append((*p)[:0], value.([]string)...)
		case *optionalBool:
			if v, ok := value.(*bool); ok && v != nil {
				p.value = v
			}
		}
	}
}

func commandFromArgs(root *Command, args []string) *Command {
	selected := root
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			if strings.Contains(arg, "=") {
				continue
			}
			name := strings.TrimPrefix(arg, "--")
			var found *flagDef
			for _, d := range selected.Flags().defs {
				if d.name == name {
					found = d
					break
				}
			}
			if found == nil {
				for _, d := range root.Flags().defs {
					if d.name == name {
						found = d
						break
					}
				}
			}
			if found != nil && found.kind != "bool" {
				i++
			}
			continue
		}
		if selected == nil || len(selected.children) == 0 {
			continue
		}
		for _, child := range selected.children {
			match := child.Use == arg
			for _, alias := range child.Aliases {
				match = match || alias == arg
			}
			if match {
				selected = child
				break
			}
		}
	}
	return selected
}

func NoArgs(_ *Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unknown argument %q", args[0])
	}
	return nil
}
