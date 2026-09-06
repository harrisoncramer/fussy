# agentslinter

A set of Go analyzers that hold code, and the comments on it, to rules an agent
writing that code tends to drift away from.

They ship two ways from the one module, so a repository can either enforce them or
be swept by them without adopting anything.

As a golangci-lint plugin, compiled into a self-contained golangci-lint binary
analogous to custom linters in other ecosystems such as ESLint. This is the form a
repository enforces in CI.

As `cmd/agentslint`, a standalone binary that runs the same analyzers over the
packages it is given. This is the form you point at somebody else's repository: it
reads nothing from the tree it is checking and writes nothing to it, so the
repository gains no config file and its git state does not move.

## Analyzers

`forbid_getenv` flags direct `os.Getenv` and `os.LookupEnv` calls so
configuration flows through the config package instead of bare os calls.

`forbid_nil_nil` flags a `return nil, nil` from a function whose last result is an
error and which returns a pointer alongside it, since the caller is handed nothing
and told nothing about why. A sentinel error the caller can match with `errors.Is`
says which of the two happened. It is held to pointers, where nil carries no meaning
of its own, rather than to a nil map or slice that already reads as empty.

`comment_length` flags any comment that runs to more than one sentence, so a
comment says one thing rather than narrating the change that made it. A free
paragraph standing on its own in a file may run long if it sets that reasoning
aside behind an `Explainer:` prefix. A doc comment on a declaration is held to its
one sentence whatever it starts with, and using the prefix there is its own
finding.

The package comment is the exception and has no length limit at all. It is the one
place a package gets to explain what it is for, and it is what a reader meets first
on pkg.go.dev, so it is meant to run to paragraphs. For the same reason it may not
use the `Explainer:` prefix, which would be published as part of the documentation
rather than read as a marker. Every non-main package has to have one, which
staticcheck's `ST1000` enforces. Where the comment outgrows the top of the file it
lives in, move it to a `doc.go` holding nothing else.

Whole files are left alone by listing path patterns under `exclude`, which are
unanchored regexps matched against the file path.

`context_timeout` flags a `context.Background` or `context.TODO` that is not immediately
bounded by `context.WithTimeout`, and a timeout written inline rather than named by a
constant, `time.Minute` included. It applies only to files matching the unanchored path
regexps under `include`, since a background context is the right answer nearly everywhere
else, and an empty `include` fails the analyzer rather than reading as on while checking
nothing. It is meant for the layer that reaches a server over the network, where a call that
never times out is a wait that never ends.

`comment_prefix` checks that a doc comment on an exported function or method
begins with the identifier name, so an LSP rename keeps the comment in sync.
With `require: true` the comment has to be there at all. In a test file the rule
covers the tests themselves, named `Test`, `Benchmark`, `Fuzz` or `Example`, and
leaves alone the exported stubs a fake needs to satisfy an interface.

Each analyzer can be turned off individually with `skip: true`, in the
`custom.agentslinter.settings` block under golangci-lint or in the file `AGENTSLINT_CONFIG`
names under the standalone binary.

## Use as a golangci-lint plugin

Name the module in the consuming repository's `.custom-gcl.yml`:

```yaml
version: v2.7.0
destination: local
plugins:
  - module: 'github.com/harrisoncramer/agentslinter'
    version: v0.1.0
```

Then enable `agentslinter` in `.golangci.yaml` and give it a `custom.agentslinter.settings`
block. Running `golangci-lint custom` fetches the core linter source and bundles the plugin in.

While iterating on a rule, add `path: ../agentslinter` alongside `module` to build against a
working tree instead of a tag.

## Use as a standalone binary

```bash
go install github.com/harrisoncramer/agentslinter/cmd/agentslint@latest
```

Run it from inside any Go module:

```bash
agentslint ./...
```

With no configuration it runs `comment_length`, `comment_prefix` and `forbid_nil_nil`, which
say something about any Go package, and leaves off `forbid_getenv` and `context_timeout`, which
assume a repository's own shape. Point `AGENTSLINT_CONFIG` at a YAML file to change that:

```yaml
comment_length:
  skip: false
  exclude:
    - config/appconfig\.go$
comment_prefix:
  skip: false
  require: true
forbid_nil_nil:
  skip: false
forbid_getenv:
  skip: true
context_timeout:
  skip: true
```

The keys are the same ones the `custom.agentslinter.settings` block takes, so a repository that
enforces the plugin and a sweep of one that does not are configured the same way. The config
file lives wherever you keep it rather than in the repository being checked. Individual
analyzers can also be switched off for one run with `-commentlength=false` and the like.

## Development

After adding or changing an analyzer, tag a release. Consumers pin the tag.

## Testing

Analyzers are tested with the `analysistest` package. Test data lives under
`testdata/src/<pkg>`, with expectations encoded as `// want` comments in the
source. Run the tests from this module:

```bash
go test ./...
```
