# TODO

## `internal/git`

+ Medium: do I need both `git.HeadDigestString` and `git.HeadDigest`? Maybe
  I should just give `Digest` a `String` method.
+ Medium: look over the entire package to decide what should be exported and
  what not.
+ Larger: improve the implementation and testing of the `git` package.
    + One idea, which is simpler: write a `pathAdapter` interface, so that we
      can easily swap real git paths for golden files in a `testdata` directory.
    + A second idea, which is more complex: use `io/fs` to create an adapter. In
      this case, the tests won't require the file system at all.
