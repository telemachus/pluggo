# TODO

## `internal/git`

+ Larger: improve the implementation and testing of the `git` package.
    + One idea, which is simple: move the git commands from internal/cli into
      internal/git.
    + Another idea, also relatively simple: write a `pathAdapter` interface, so
      that we can easily swap real git paths for golden files in a `testdata`
      directory.
    + A third idea, which is more complex: use `io/fs` to create an adapter. In
      this case, the tests won't require the file system at all.
