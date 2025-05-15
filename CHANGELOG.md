# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Add generated CHANGELOG.md

### Changed

- Bump github.com/charmbracelet/bubbletea from 1.3.4 to 1.3.5
- Configure Github permissions for build workflow
- Enable more revive linting rules
- Extract keyboard bindings diagram to a separate file and add a Makefile
- Update to go v1.24 and golangci v2

## [0.5.3] - 2025-03-16

### Added

- Add a simple input prompt for the export path
- Add bubbletea model method tests
- Add export tests
- Add prompt cursor movement
- Add prompt home and end actions
- Add separate exportMarkedPath
- Add test and resolve selection bug

### Changed

- Ensure single-mark selection is exported
- Refactor bubbletea model into edits package
- Scroll and truncate prompt responses
- Update deprecated goreleaser properties

## [0.5.2] - 2025-03-08

### Added

- Add overwrite confirmation if export file is not empty

### Changed

- Bump github.com/charmbracelet/bubbletea from 1.1.1 to 1.1.2
- Bump github.com/charmbracelet/bubbletea from 1.1.2 to 1.2.1
- Bump github.com/charmbracelet/bubbletea from 1.2.1 to 1.2.2
- Bump github.com/charmbracelet/bubbletea from 1.2.2 to 1.2.4
- Bump github.com/charmbracelet/bubbletea from 1.2.4 to 1.3.0
- Bump github.com/charmbracelet/bubbletea from 1.3.0 to 1.3.3
- Bump github.com/charmbracelet/bubbletea from 1.3.3 to 1.3.4
- Bump github.com/muesli/termenv from 0.15.2 to 0.16.0
- Bump github.com/stretchr/testify from 1.9.0 to 1.10.0
- Extract EBNF diagram to a separate file and create a PNG
- Simplify type arguments
- Update copyright notices
- Update documentation

## [0.5.1] - 2024-10-05

### Added

- Implement disjoint cuts as documented

## [0.5.0] - 2024-10-04

### Added

- Add Table of Contents
- Add initial permascroll to Init() for testing
- Implement selecting from multiple cuts

### Changed

- Change module URL
- Extract terminal styles to a separate file
- Improve Japanese translations
- Precompute cursor strings
- Preserve all cuts with timestamps
- Use ANSI color numbers

## [0.4.0] - 2024-09-26

### Added

- Implement Join action
- Implement ReplaceText

### Changed

- Update documentation

## [0.3.1] - 2024-09-25

### Changed

- Skip Copy operation when no text is available

## [0.3.0] - 2024-09-23

### Added

- Implement text translations

### Changed

- Refactor edits.ShowHelp to edits.Mode

## [0.2.0] - 2024-09-23

### Added

- Implement help screen

## [0.1.0] - 2024-09-23

### Added

- Add GPLv3 license
- Add command line help and increment version
- Add documentation and extract func drawLine from DrawWindow
- Add goreleaser configuration to create releases
- Add test and resolve issue when inserting over pending deletion
- Create dependabot.yml
- Create go.yml
- Create jotty.go using tcell v2
- Create project
- Implement Backspace()
- Implement Copy, cut and InsertCut actions
- Implement Delete()
- Implement DeleteText() and InsertText() in document package
- Implement Export action
- Implement ExportText
- Implement combining characters and word wrap
- Implement cursor
- Implement dispatch table
- Implement doc.MoveParagraphs and nav.MoveParas
- Implement document hashing for automatic versioning
- Implement edit marks
- Implement exchange operation and disjoint selections using fourth mark
- Implement initial capital letters
- Implement linear Undo() and Redo()
- Implement minimal status line
- Implement modal exit confirmation and error messages
- Implement navigation
- Implement paragraph counts
- Implement paragraph splitting
- Implement permascroll
- Implement permascroll backing store
- Implement primary and secondary selections
- Implement replacing text in primary selection
- Implement resize indicator when screen is too small
- Implement scope highlight in status bar
- Implement scrolling
- Implement sections
- Implement sentence counts
- Implement single-entry cut buffer
- Implement word count

### Changed

- Also show the terminal cursor
- Always DrawCursor at end of DrawStatusBar
- Automatically embed version string
- Bump github.com/charmbracelet/bubbletea from 0.26.6 to 0.27.0
- Bump github.com/charmbracelet/bubbletea from 0.27.0 to 0.27.1
- Bump github.com/charmbracelet/bubbletea from 0.27.1 to 1.1.0
- Bump github.com/charmbracelet/bubbletea from 1.1.0 to 1.1.1
- Bump github.com/stretchr/testify from 1.8.4 to 1.9.0
- Change AppendByte back to AppendRune and refactor tests
- Change character count symbol to "@"
- Change document API to prepare for undo
- Convert ABNF to EBNF for PlantUML diagrams
- Convert svgbob to ditaa for PlantUML diagram
- Destructure var cursor
- Enable linters and implement recommendations
- Enhance AppendRunes() into InsertRunes()
- Extract document operations into package document
- Extract func advanceLine from DrawWindow
- Extract miscellaneous actions to a new file and tidy up
- Extract screenLine() to DRY Screen()
- Generate StatusLine as part of func Screen
- Improve state diagram
- Initialise state at start of each line
- Make func drawCursor and drawStatusBar private
- Move blank line between paras to Screen()
- Multiple sections can appear on screen, so maintain indexes for them
- Only dispatch actions when window size is sufficient
- Port from ncurses to bubbletea
- Provide one preceding paragraph of context if possible
- Refactor Backspace()
- Refactor DrawStatusBar
- Refactor and implement AppendRune
- Refactor and implement character count
- Refactor and implement redraw on resize
- Refactor func drawLine
- Refactor into func newBuffer
- Refactor tests to be table-driven
- Refactor to only call drawWindow from Screen
- Rename "sect" to "sectn" to more clearly distinguish it from "sent" for sentences
- Rename buffer.go to edits.go
- Rename variables for clarity
- Replace line.r with line.brk for clarity
- Rewrite to replace tcell with go-ncursesw
- Simplify TestDrawWindow()
- Simplify cache
- Simplify test assertions
- Unify index and cache
- Update tests
- Update tests and implement Space and Enter functions
- Var total is only needed within drawStatusBar

### Removed

- Completely remove "sections" design
- Remove unused "text" field in buffer lines
- Remove unused var Screen

[unreleased]: https://github.com/xanni/jotty/compare/v0.5.3...HEAD
[0.5.3]: https://github.com/xanni/jotty/compare/v0.5.2...v0.5.3
[0.5.2]: https://github.com/xanni/jotty/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/xanni/jotty/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/xanni/jotty/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/xanni/jotty/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/xanni/jotty/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/xanni/jotty/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/xanni/jotty/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/xanni/jotty/compare/bb69f47296ccbe2b...v0.1.0

<!-- generated by git-cliff -->
