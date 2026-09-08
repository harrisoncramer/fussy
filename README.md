# fussy

A golangci-lint plugin, and a standalone binary, holding a repository to a handful of rules that
a general-purpose linter has no opinion about. Several of them are about comments, and the rest
are naming conventions a repository has settled on and wants held.

## The rules

`commentlength` holds a comment to one sentence. A package comment may run as long as it needs
to, and a free paragraph inside a file may run long behind an `Explainer:` prefix. Set
`max_chars` to cap the sentence as well.

`commentreference` reports a comment naming a test or another file. The name rots the moment
either side is renamed and nothing fails when it does, and a reader who has to open a second file
to finish reading a line has been sent away by the thing meant to save them the trip. A doc
comment naming the identifier it documents is fine, and so is `pkg.go.dev`.

`commentprefix` requires an exported function's doc comment to begin with the identifier it
documents. Nothing else is required to carry a comment.

`forbidgetenv` forbids reading the environment outside the package that loads configuration.

`forbidnilnil` forbids returning a nil pointer with a nil error.

`contexttimeout` asks the paths it is pointed at to bound the contexts they start.

`paramsstruct` holds a params struct to the name of the function that takes it, and holds the
parameter itself to `p`. `NewMailService` takes a `NewMailServiceParams` called `p`, and `Record`
takes a `RecordParams`. The alternative rule, naming the struct after the type built, reads just
as well in isolation, which is why a repository ends up with both and no way to tell which one a
given struct is following. A struct declared in another package is held only to the parameter
name, since it cannot be renamed from the call site, and a struct two functions in the package
take is held only to the parameter name as well, since a wrapper or a retrying variant can
satisfy no name that mentions one of them. Test files are left alone.

`tabletest` holds a table-driven test to one shape: the slice is `tests`, the range binds `tt`,
and the subtest is named `tt.name`. It anchors on a slice of structs carrying a `name` field, so
a cross-product sweep over an enum with a computed subtest name is a different shape and is left
alone.

`testdouble` reports a type in a test file named with a `stub`, `mock` or `spy` prefix. The words
are used interchangeably for the same shape, so a repository picks one; `fake` is the default and
both halves are configurable.

`storeverb` asks an exported method to open with a verb the repository has settled on. It reaches
only the paths `include` names, and names none by default, since only a repository that has
decided its verbs has anything for it to hold. The `allow` list carries the method names an
interface has already picked, such as `Scan` and `String`, which no rule of the repository's own
can reach.

## Why the comment rules read the way they do

A comment says what the thing is. Most of the reasoning an author reaches for is worth nothing
and should be deleted outright. The reasoning that is genuinely load-bearing goes where a reader
finds it without being sent looking: the package comment, which may run as long as it needs, or
one `Explainer:` paragraph beside the code it defends. There are a handful of those in a healthy
repository, not one per function.

A commit message is not that place. It describes a delta, and is read by someone who already
knows which delta they care about; a reader who opens the file has no way to know which of a
thousand commits explains the shape they are looking at.

Most declarations want no comment at all. A name that already says what the thing is has nothing
left for a comment to add, and one written anyway fills up with reasoning to justify existing.

When `commentlength` rejects a second sentence, delete it. Folding it into the first with a comma
satisfies the rule and produces a worse sentence than the two it replaced.

Files a tool wrote are skipped by every comment rule, since a generated header is not an author's
comment to fix.

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
store_verb:
  skip: false
  # The paths the rule reaches. None leaves it switched off.
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

Every `exclude` and `include` is an unanchored regular expression over the file path. Generated
files are skipped by every rule.

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

It draws three verdicts rather than two. An identifier only its own package uses should be
unexported. An identifier only its own tests use is dead and should be deleted, since unexporting
it would turn a visibly unclaimed export into a package-private function no tool will flag again.
An identifier an external test package uses is left alone, since unexporting it would stop those
tests compiling.

A module rooted inside the one being swept is not swept with it, since `./...` resolves to the
module at the working directory and a nested module has a build of its own. A caller there is a
caller in another module, which is the plainest reason an identifier has to stay exported, so
those files are read for the names they use and the module is named in the report.

Methods are not reported yet, because unexporting one can break interface satisfaction silently
and there is no implements check here to clear it. Files no build configuration compiled, such as
a generator behind a go:build ignore tag, are read for the names they use and named in the report,
so a caller behind a build tag does not read as an export nobody wants.

Pass -json for a machine-readable result, -kinds to narrow the sweep to some of func, type, var
and const, and -generated to include identifiers a generator wrote. The command exits 1 when it
has findings.
