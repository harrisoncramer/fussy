# fussy

A golangci-lint plugin, and a standalone binary, holding a repository to a handful of rules that
a general-purpose linter has no opinion about. Several of them are about comments, and the rest
are naming conventions and other "fussy" fixes designed to enforce strict patterns for Go agents.

## The rules

`commentlength` holds a comment to one sentence. A package comment may run as long as it needs
to, and a free paragraph inside a file may run long behind an `Explainer:` prefix. Set
`max_chars` to cap the sentence as well.

`commentreference` reports a comment naming a test or another file. The name rots the moment
either side is renamed and nothing fails when it does.

`commentprefix` requires an exported function's doc comment to begin with the identifier it documents.

`forbidgetenv` forbids reading the environment outside the package that loads configuration.

`forbidnilnil` forbids returning a nil pointer with a nil error.

`contexttimeout` asks the paths it is pointed at to bound the contexts they start.

`paramsstruct` holds a params struct to the name of the function that takes it, and holds the
parameter itself to one name, `params` by default and whatever `parameter_name` says otherwise.
`NewMailService` takes a `NewMailServiceParams` called `params`, for instance.

`tabletest` holds a table-driven test to one shape: the slice is `tests`, the range binds `tt`,
and the subtest is named `tt.name`. It anchors on a slice of structs carrying a `name` field that
the test then ranges over, so a cross-product sweep over an enum with a computed subtest name is
a different shape and is left alone, and so is a slice of that shape the test builds as its own
data and never drives.

`testdouble` reports a type in a test file named with a `stub`, `mock` or `spy` prefix. The words
are used interchangeably for the same shape, so a repository picks one; `fake` is the default and
both halves are configurable.

`storeverb` asks an exported method to open with a verb the repository has settled on. It reaches
only the paths `include` names, and names none by default, since only a repository that has
decided its verbs has anything for it to hold. The `allow` list carries the method names an
interface has already picked, such as `Scan` and `String`, which no rule of the repository's own
can reach.

## Configuration

Both the plugin and the binary take the same settings block. Under golangci-lint it is the
plugin's `settings`; standalone it is a YAML file the `FUSSY_CONFIG` environment variable points
at.

```yaml
comment_length:
  skip: false
  max_chars: 0
  exclude:
    - config/appconfig\.go$
comment_prefix:
  skip: false
  require: true
comment_reference:
  skip: false
  exclude: []
context_timeout:
  skip: false
  include:
    - internal/ui/
forbid_getenv:
  skip: false
forbid_nil_nil:
  skip: false
params_struct:
  skip: false
  exclude: []
  # The name every params struct is taken under.
  parameter_name: params
  # The structs held to no function name, still held to the parameter name.
  allow: []
store_verb:
  skip: false
  include:
    - internal/store/
  # The verbs a method may open with, replacing the built-in list of Abandon, Archive,
  # Begin, Claim, Clear, Count, Create, Delete, Exists, Finish, Forget, Get, Lock, List,
  # Mark, Move, Open, Prune, Record, Release, Rename, Restore, Search, Send, Set, Unmark,
  # Update and Upsert.
  verbs: []
  # The methods held to no verb, replacing the built-in Close, Error, MarshalJSON,
  # MarshalText, Scan, ServeHTTP, String, UnmarshalJSON, UnmarshalText and Value.
  allow: []
table_test:
  skip: false
  exclude: []
test_double:
  skip: false
  exclude: []
  # The prefixes reported, replacing the built-in stub, mock and spy.
  forbidden:
    - stub
    - mock
    - spy
  # The word the report points at.
  preferred: fake
```

## Running it

As a golangci-lint plugin, name the module in `.custom-gcl.yml` and build the custom binary.

Standalone, `go run github.com/harrisoncramer/fussy/cmd/fussy ./...` runs every rule the
configuration leaves on, and a rule can be selected by name: `fussy -commentlength ./...`.

## fussy unexported

Sweeps a whole module for exported identifiers no other package uses.

    fussy unexported ./...

This cannot be one of the analyzers. An analyzer runs over one package at a time and sees its
dependencies as export data, so it can never see whether some other package imports the symbol in
front of it. The sweep loads every package of the module at once instead, which is why it is a
subcommand.

An identifier only its own package uses should be unexported. 
An identifier only its own tests use is dead and should be deleted.
An identifier an external test package uses is left alone.
