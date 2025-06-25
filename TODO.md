# TODO

## `internal/git`

+ Use `io/fs` rather than my thinner interface of `ReadFile`?

## `internal/cli/sync`

+ The `manageUpdate` method handles both moves and updates. I think that this
  may lead to some repetition (especially surrounding `pSpec.Pinned`). I should
  separate `move` from `update`.
