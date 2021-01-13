# Guide for repository maintainers

## Run integration tests from fork repository

To run integration tests from fork repositories maintainer must carefully check changes that a PR is trying to make. **It is the maintainer's responsibility to avoid secrets leak.**

Integration tests are executed after adding a comment:

```bash
/ok-to-test sha=<short-commit>
```

After the `ok-to-test` job has started, move to the GitHub actions page to see the output from integration-tests jobs (click `ok-to-test` job).

*Note: the `integration-fork` job status is updated on the main PR page once it is finished.*
 