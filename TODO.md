# TODO

+ Medium: improve error handling, especially in `plugin.go`. The app should
  return a non-zero status if any problem occurs, but it should only abort for
  grave errors.
+ Larger: improve the implementation and testing of the `git` package cleaner.
    + Stage one, which is simpler: write a `pathAdapter` interface, so that we
      can easily swap real git paths for golden files in a `testdata` directory.
    + Stage two, which is more complex: use `io/fs` to create an adapter. In
      this case, the tests won't require the file system at all.
