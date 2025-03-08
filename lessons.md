# Lessons from the Jotty project

This project aims to demonstrate some concepts from Ted Nelson's software and
user interface designs.  I chose Go as the implementation language because of
the ability to produce cross-platform executables that stand alone and do not
require any special installation or additional dependencies.  This document will
describe features and design choices that illuminate Ted Nelson's ideas, the Go
language, and general software development techniques and pracices.

## Documentation

When planning a project it's critical to first establish and document the goals
and the audience as the very first step, even before gathering the requirements.
In determining the boundaries of the project it's also useful to document some
non-goals in order to clarify what is in and out of scope.

For Jotty, the audiences are writers interested in a tool that prioritises rapid
and efficient keyboard-only text entry and rearragement, and also software
developers interested in learning about some of Ted Nelson's early designs.

## Tests

Every project should have tests, and the tests need to be an integrated part of
the codebase.  Test coverage alone is neither necessary nor sufficient for
verifying correctness, but achieving good coverage is very helpful in many
respects.  It ensures that all the code is actually executed, which helps the
developers consider not only whether the code is correct but also whether it is
necessary or useful.
