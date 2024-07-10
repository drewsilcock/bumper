# Bumper

> This is my version bumper. There are many like it, but this one is mine. 

## Getting Started

```bash
git install github.com/drewsilcock/bumper@latest

cd ~/my-project

# Assuming you have GOBIN in your PATH, otherwise prepend with "$(go env GOPATH)/bin/"
bumper
```

## Assumptions

This is made with my personal workflow in mind, so we make certain assumptions:

- The readme is called `README.md` and contains as the first line `# {Project Name}`.
- The changelog is called `CHANGELOG.md` and contains a list of versions in the format `## v{Version} - {Date}` with the unreleased changes in a section at the top called either `## Unreleased` or `## Development`.
- Git flow is being with the development branch called `dev` and the main branch called `main`.
- Tags are added to the main branch but the tagged commits are merged into dev so that they are accessible on the dev branch.

## Configuration

You can specify the bump type (major, minor, patch) via the the CLI or via a prompt.

When you first try to create a GitLab release, you will be prompted for a personal access token with the `api` permission. This is stored in the config file `~/.config/bumper/config.toml` for future use.