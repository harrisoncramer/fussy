# fussy

A golangci-lint plugin, and a standalone binary, holding a repository to a handful of rules that
a general-purpose linter has no opinion about. Most of them are about comments.

## The rules

`commentlength` holds a comment to one sentence. A package comment may run as long as it needs
to, and a free paragraph inside a file may run long behind an `Explainer:` prefix. Set
`max_chars` to cap the sentence as well.

`commentclause` reports the trailing justification clause. A one-sentence rule does not stop an
author saying two things; it pushes them to glue the second onto the first with a comma and a
"since", "because", "so that" or "rather than". Package comments are exempt.

`commentreference` reports a comment naming a test or another file. The name rots the moment
either side is renamed and nothing fails when it does, and a reader who has to open a second file
to finish reading a line has been sent away by the thing meant to save them the trip. A doc
comment naming the identifier it documents is fine, and so is `pkg.go.dev`.

`commentprefix` requires an exported function's doc comment to begin with the identifier it
documents. Nothing else is required to carry a comment.

`forbidgetenv` forbids reading the environment outside the package that loads configuration.

`forbidnilnil` forbids returning a nil pointer with a nil error.

`contexttimeout` asks the paths it is pointed at to bound the contexts they start.

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
satisfies the rule and produces a worse sentence than the two it replaced, which is what
`commentclause` is for.

Files a tool wrote are skipped by every comment rule, since a generated header is not an author's
comment to fix.

## Configuration

Both the plugin and the binary take the same settings block. Under golangci-lint it is the
plugin's `settings`; standalone it is a YAML file the `FUSSY_CONFIG` environment variable points
at.

```yaml
comment_clause:
  skip: false
  exclude:
    - config/appconfig\.go$
  # The joins reported, replacing the built-in list of since, because, so that,
  # rather than, which is, and so as to. A codebase adopting the rule mid-life
  # will want to start with the unambiguous ones and add the rest later.
  clauses:
    - since
    - because
    - so that
    - rather than
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
```

Every `exclude` and `include` is an unanchored regular expression over the file path. Generated
files are skipped by every rule.

## Running it

As a golangci-lint plugin, name the module in `.custom-gcl.yml` and build the custom binary.

Standalone, `go run github.com/harrisoncramer/fussy/cmd/fussy ./...` runs every rule the
configuration leaves on, and a rule can be selected by name: `fussy -commentclause ./...`.

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
