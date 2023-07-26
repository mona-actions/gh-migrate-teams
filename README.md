# gh-migrate-teams

`gh-migrate-teams` is a [GitHub CLI](https://cli.github.com) extension to assist in the migration of teams between GitHub organizations. [GitHub Enterprise Importer](https://github.com/github/gh-gei) provides an excellent feature set when migrating organizations, but there are some gaps when it comes to migrating teams. This extension aims to fill those gaps. Wether you are consolidating organizations in an [EMU](https://docs.github.com/en/enterprise-cloud@latest/admin/identity-and-access-management/using-enterprise-managed-users-for-iam/about-enterprise-managed-users), or auditing teams and collaborators in an existing organization, this extension can help.

## Install

```bash
gh extension install mona-actions/gh-migrate-teams
```

## Usage: Export

Export team membership, team repository access, and repository collaborator access to CSV files.

```bash
Usage:
  migrate-teams export [flags]

Flags:
  -f, --file-prefix string    Output filenames prefix
  -h, --help                  help for export
  -o, --organization string   Organization to export
  -t, --token string          GitHub token
```

## Usage: Sync

Recreates teams, membership, and team repo roles from a source organization to a target organization.

```bash
Usage:
  migrate-teams sync [flags]

Flags:
  -h, --help                         help for sync
  -s, --source-organization string   Source Organization to sync teams from
  -a, --source-token string          Source Organization GitHub token
  -t, --target-organization string   Target Organization to sync teams from
  -b, --target-token string          Target Organization GitHub token
```

## License

- [MIT](./license) (c) [Mona-Actions](https://github.com/mona-actions)
- [Contributing](./contributing.md)
