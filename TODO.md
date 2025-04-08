# TODO

+ Minor: move the `updateHelpTags` to `cmd.go` and call it from `pluggo.go` as
  part of the sequence of larger commands. Add a `noOp` check at the top of the
  function.
+ Medium: require a `basedir` object in the configuration file. The object
  should be an array of strings, which will be joined together to become the
  base directory of the plugins, not including `start/` or `opt/` which will be
  added by the tool depending on the plugins `opt` setting. If the first item in
  the array is `HOME`, the tool will replace this with the full path to the
  user's home directory.
+ Medium: improve error handling, especially in `plugin.go`. The app should
  return a non-zero status if any problem occurs, but it should only abort for
  grave errors.
+ Larger: improve the implementation and testing of the `git` package cleaner.
    + Stage one, which is simpler: write a `pathAdapter` interface, so that we
      can easily swap real git paths for golden files in a `testdata` directory.
    + Stage two, which is more complex: use `io/fs` to create an adapter. In
      this case, the tests won't require the file system at all.
