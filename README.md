# fussy

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

Methods are not reported yet, because unexporting one can break interface satisfaction silently
and there is no implements check here to clear it. Files no build configuration compiled, such as
a generator behind a go:build ignore tag, are read for the names they use and named in the report,
so a caller behind a build tag does not read as an export nobody wants.

Pass -json for a machine-readable result, -kinds to narrow the sweep to some of func, type, var
and const, and -generated to include identifiers a generator wrote. The command exits 1 when it
has findings.
