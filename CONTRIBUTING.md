# Contributing

Authorization Proxy uses GitHub to manage reviews of pull requests.

- The versioning scheme we use is [SemVer](http://semver.org/).
- Relevant coding style guidelines are the [Go Code Review Comments](https://code.google.com/p/go-wiki/wiki/CodeReviewComments) and the _Formatting and style_ section of Peter Bourgon's [Go: Best Practices for Production Environments](https://peter.bourgon.org/go-in-production/#formatting-and-style).

## Steps to Contribute

Should you wish to work on an issue, please claim it first by commenting on the GitHub issue that you want to work on it. This is to prevent duplicated efforts from contributors on the same issue.

For quickly compiling and testing your changes, do:
```bash
# for building
go build
./authorization-proxy

# for testing (Make sure all the tests pass before you commit and push :))
make test
```

## Pull Request Process

- Branch from the master branch and, if needed, rebase to the current master branch before submitting your pull request. If it doesn't merge cleanly with master you may be asked to rebase your changes.
    - Branches should have descriptive names and start with prefixes like `patch/`, `fix/`, `feature/`. Good examples are: `fix/vulnerability-issue` or `feature/issue-templates`.
- If your patch is not getting reviewed or you need a specific person to review it, you can @-reply a reviewer asking for a review in the pull request or a comment.
- Add tests relevant to the fixed bug or new feature.
- Update the README.md with details of changes to the interface, this includes new environment variables, exposed ports, useful file locations and container parameters.
- Add prefix `[major]`, `[minor]`, `[patch]` or `[skip]` in the PR title for releasing.
- Please use `Squash and merge` to merge a PR and double-confirm the merging message.
- Merging PR to master will increase the version no. and create a new release automatically.
    - The new version no. depends on the commit message prefix when new PRs is merged to master branch.
    - Commit message prefix and release tag reference table:
        | **Commit Msg Prefix** | **New Version No.**  | **Release `latest` Tag** | **Release `nightly` Tag** |
        |:---------------------:|:--------------------:|:------------------------:|:-------------------------:|
        | `[major] *`           | `v1.2.3` => `v2.0.0` | ✅                        | ✅                         |
        | `[minor] *`           | `v1.2.3` => `v1.3.0` | ✅                        | ✅                         |
        | `[patch] *`           | `v1.2.3` => `v1.2.4` | ✅                        | ✅                         |
        | `[skip] *`            | ❌                    | ❌                        | ✅                         |
        - Other prefixes will cause the pipeline to **FAIL**❌.

## Dependency management

The Authorization Proxy project uses [Go modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) to manage dependencies on external packages. This requires a working Go environment with version 1.18 or greater installed.

To add or update a new dependency, use the `go get` command:

```bash
# Pick the latest tagged release.
go get example.com/some/module/pkg

# Pick a specific version.
go get example.com/some/module/pkg@vX.Y.Z
```

Tidy up the `go.mod` and `go.sum` files:

```bash
# The GO111MODULE variable can be omitted when the code isn't located in GOPATH.
GO111MODULE=on go mod tidy
```

You have to commit the changes to `go.mod` and `go.sum` before submitting the pull request.

## Contributor Covenant Code of Conduct

### Attribution

This Code of Conduct is adapted from the [Contributor Covenant](https://www.contributor-covenant.org/), version 2.0, available at <https://www.contributor-covenant.org/version/2/0/code_of_conduct.html>.
