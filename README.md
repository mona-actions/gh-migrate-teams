# gh-migrate-teams

`gh-migrate-teams` is a [GitHub CLI](https://cli.github.com) extension to assist in the migration of teams between GitHub organizations. Currently it provides the ability to export teams, team membership, repos, and repo permissions to CVS files for reference.

## Install

```bash
gh extension install mona-actions/gh-migrate-teams
```

## Usage

```bash
Usage:
  migrate-teams export [flags]

Flags:
  -f, --file-prefix string    Output filenames prefix
  -h, --help                  help for export
  -o, --organization string   Organization to export
  -t, --token string          GitHub token
```

## License

- [MIT](./license) (c) [Mona-Actions](https://github.com/mona-actions)
