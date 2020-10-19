package concourse

/*
 Where are all the tests?

So we had a ton of tests that made it super annoying to change the pipeline renderer as they
were far to granular and coupled to the implementation...

Given that we have good e2e tests we test the renderer in there.

But! We have a test suite in `cmd/cmds/root_test`.
This suite basically invoke the e2e tests (without diff, as we diff in the build script)
so we can check the code coverage of the renderer.

In idea, just right click on the test and `Run '...' with coverage`
*/
